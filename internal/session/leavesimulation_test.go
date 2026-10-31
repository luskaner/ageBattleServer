package session

import (
	"bytes"
	"testing"
)

// The LeaveSimulation message (0x1F) is the densest thing this server used to get
// wrong by name, and the tests here are the ones that pin it. Its handler is three
// lines: it takes the sender's session, does nothing if there is none, and hands
// the sender's peer id to the leave routine.
//
// The leave routine empties two queues, clears one byte and, in one 16-bit store,
// ends the recording. It does not remove the peer. The only thing that removes
// a peer is the disconnect path.

// recordingSession is a session with one peer and the frame store already created,
// which is the state the game start leaves behind: Recording set and the
// peer marked as being in the started game (Participating).
func recordingSession(t *testing.T, id int64, peerIDs ...byte) (*Session, []*mockConn) {
	t.Helper()
	s := NewSession(id)
	conns := make([]*mockConn, 0, len(peerIDs))
	for _, pid := range peerIDs {
		c := &mockConn{}
		s.AddPeer(NewPeer(pid, c))
		conns = append(conns, c)
	}
	s.RecordFrame(ackFrameType, ackFrame(0))
	if !s.Recording {
		t.Fatal("the frame store was created and Recording should be set")
	}
	// GetPeers and not a walk of s.Peers: the peer table is a slot table
	// ([256]*Peer, indexed by peer id) and a slot table has no
	// order and 255 empty slots. Everything that wants the peers asks for the
	// walk, which is peerOrder.
	for _, p := range s.GetPeers() {
		if !p.Participating {
			t.Fatalf("peer %d should carry the Participating mark the store creation sets", p.ID)
		}
	}
	return s, conns
}

// TestLeaveSimulationEndsTheRecordingWhenTheLastPeerLeaves pins
// the 16-bit store that ends the match: the value written is 256, which clears
// the recording flag and sets GameStarted, and it only happens when the session
// was recording and no peer is left in the game.
func TestLeaveSimulationEndsTheRecordingWhenTheLastPeerLeaves(t *testing.T) {
	s, _ := recordingSession(t, 1, 1)
	s.QueueGameData(1, []byte{0xaa, 0xbb})

	if body, _ := s.PopAckTick(); body == nil {
		t.Fatal("the session should still be acking before the leave")
	}

	if !s.LeaveSimulation(1) {
		t.Fatal("LeaveSimulation did not find peer 1")
	}
	if s.Recording {
		t.Error("Recording should be clear once the match is over")
	}
	if !s.GameStarted {
		t.Error("GameStarted should be set: it is the other byte of the same store")
	}
	if s.GetPeer(1) == nil {
		t.Fatal("the peer must stay in the session: the leave does not remove it")
	}
	if s.GetPeer(1).Participating {
		t.Error("Participating should be clear for the leaver")
	}
	// The tick then stops building acks, because the loop that decides whether to
	// keep recording reads the same flag.
	if body, notice := s.PopAckTick(); body != nil || notice != nil {
		t.Errorf("PopAckTick after the match ended returned %d bytes / %v, want nothing", len(body), notice)
	}
}

// TestLeaveSimulationKeepsRecordingWhileAPeerRemains is the other half of the same
// condition: one peer leaving is not the end of the match.
func TestLeaveSimulationKeepsRecordingWhileAPeerRemains(t *testing.T) {
	s, _ := recordingSession(t, 1, 1, 2)
	s.QueueGameData(2, []byte{0x33})

	s.LeaveSimulation(1)

	if !s.Recording || s.GameStarted {
		t.Fatalf("recording=%v gameStarted=%v, want true/false: peer 2 is still in the game "+
			"(the peers are scanned for a Participating one before the recording is ended)", s.Recording, s.GameStarted)
	}
	if body, _ := s.PopAckTick(); body == nil {
		t.Error("the session should still be acking while a peer remains")
	}
}

// TestLeaveSimulationDropsOnlyTheLeaversQueuedData pins the leave: both
// of the leaver's queues are emptied and the other peer's are untouched.
func TestLeaveSimulationDropsOnlyTheLeaversQueuedData(t *testing.T) {
	s, _ := recordingSession(t, 1, 1, 2)
	s.QueueGameData(1, []byte{0x11, 0x22})
	s.QueueGameData(2, []byte{0x33, 0x44})
	entry := []byte{1, 0, 0, 0, 7, 0, 0, 0}
	s.QueueChecksumEntries(1, entry)
	s.QueueChecksumEntries(2, entry)

	s.LeaveSimulation(1)

	body, _ := s.PopAckTick()
	if body == nil {
		t.Fatal("no ack built")
	}
	if bytes.Contains(body, []byte{0x11, 0x22}) {
		t.Errorf("the leaver's queued game data is still in the ack %x: the leave empties it", body)
	}
	if !bytes.Contains(body, []byte{0x33, 0x44}) {
		t.Errorf("the remaining peer's game data is missing from the ack %x", body)
	}
	// The leaver's checksum entry goes with it, and with it goes the reconciled
	// block: the reconcile needs an entry from EVERY peer, so the remaining peer's
	// entries are held back rather than flushed on their own.
	if bytes.Contains(body, entry) {
		t.Errorf("the ack carries checksum entries with one peer left holding none: %x", body)
	}
	if !bytes.HasSuffix(body, []byte{0x00, 0x00}) {
		t.Errorf("the ack does not end in the empty block %x", body)
	}
	// One sub made it out of the two that were queued: the flush is [0xfe][count].
	if body[0] != 0xfe || body[1] != 1 {
		t.Errorf("the ack flushes %d sub-messages, want 1: %x", body[1], body)
	}
}

// TestEndOfGameMarkerNeedsTheClientsMessageVersion pins the
// end-of-game marker: it is for an observer whose own connection announces message
// version 4 or more, on top of the arm that says the store's end is its target.
// The observers of the committed captures connect and go straight to 0x1C without
// a LobbyConnect, and a connection's message version is only ever set by the
// LobbyConnect handler, so they are at 0 and the
// original sent them none.
func TestEndOfGameMarkerNeedsTheClientsMessageVersion(t *testing.T) {
	run := func(version uint16) []mockMsg {
		s := NewSession(1)
		s.SetObserverConfig(1, 1)
		// No peers: the store's end is the target, so the only thing left
		// to decide is the version.
		recordAcks(s, 4)
		conn := &mockConn{msgVersion: version}
		if status, _ := s.StartObserve(conn, 1, 0); status != 0 {
			t.Fatalf("status %d, want 0", status)
		}
		if !s.ActivateObserver(conn, 8) {
			t.Fatal("activation failed")
		}
		conn.msgs = nil
		for range 3 {
			s.Tick()
		}
		return conn.msgs
	}

	if hasMsg(run(0), 0x22) {
		t.Error("an observer that announced no message version was sent the type-34 marker")
	}
	if !hasMsg(run(4), 0x22) {
		t.Error("an observer that announced message version 4 was not sent the type-34 marker")
	}
}

// TestLeaveSimulationKeepsServingObservers is what ends the recording does NOT do:
// the store is not destroyed and the observers keep being served from it. The
// session tick serves the observers BEFORE it tests whether the recording is
// still running, and that update takes the live end of the store as the target.
func TestLeaveSimulationKeepsServingObservers(t *testing.T) {
	s, _ := recordingSession(t, 1, 1)
	s.SetObserverConfig(1, 1)
	recordAcks(s, 5)

	obs := &mockConn{msgVersion: 4}
	if status, _ := s.StartObserve(obs, 1, 0); status != 0 {
		t.Fatalf("status %d, want 0", status)
	}
	if !s.ActivateObserver(obs, 8) {
		t.Fatal("activation failed")
	}
	obs.msgs = nil

	s.LeaveSimulation(1)
	if !s.GameStarted {
		t.Fatal("the match should be over after the last peer leaves")
	}
	for range 3 {
		s.Tick()
	}
	if n := countMsgs(obs.msgs, ackFrameType); n == 0 {
		t.Fatalf("the observer was served nothing after the match ended: %v", obs.msgs)
	}
	if !hasMsg(obs.msgs, 0x22) {
		t.Error("a caught-up observer of a finished match should get the end-of-game marker")
	}
}