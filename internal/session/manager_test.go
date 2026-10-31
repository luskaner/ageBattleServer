package session

import (
	"fmt"
	"testing"
)

func TestManagerNew(t *testing.T) {
	mgr := NewManager(0, 0)
	if mgr == nil {
		t.Fatal("NewManager(0, 0) returned nil")
	}
	if mgr.SessionCount() != 0 {
		t.Errorf("expected 0 sessions, got %d", mgr.SessionCount())
	}
}

// --- frame store tests (Replay) ---

// storeFrame builds a frame body of a size the tests can do arithmetic on: a
// body of n bytes costs n+frameOverhead against its block.
func storeFrame(n int, fill byte) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = fill
	}
	return b
}

func TestReplayKeepsEverythingWhileTheBlocksHaveRoom(t *testing.T) {
	r := NewReplay(2, 200, -1)
	for i := range 4 {
		r.AddFrame(0x0A, storeFrame(43, byte(i)))
	}
	// 43+57 = 100 bytes a frame, so two frames fill a 200-byte block, and the
	// four frames fit in the two blocks: nothing is ever dropped.
	if got := r.Retained(); got != 4 {
		t.Fatalf("Retained = %d, want 4", got)
	}
	if got := r.Total(); got != 4 {
		t.Fatalf("Total = %d, want 4", got)
	}
	if r.Base() != 0 {
		t.Fatalf("Base = %d, want 0: the store still holds the first frame of the match", r.Base())
	}
	if got := len(r.GetFrames(0, 3)); got != 4 {
		t.Fatalf("GetFrames(0,3) = %d frames, want 4", got)
	}
}

func TestReplayDropsTheOldestBlockWhenTheyAreAllFull(t *testing.T) {
	r := NewReplay(2, 200, -1)
	for i := range 5 {
		r.AddFrame(0x0A, storeFrame(43, byte(i)))
	}
	// Five frames at 100 bytes need three blocks and the cap is two, so the first
	// one went and the window starts at frame 2.
	if got := r.Retained(); got != 3 {
		t.Fatalf("Retained = %d, want 3", got)
	}
	if got := r.Total(); got != 5 {
		t.Fatalf("Total = %d, want 5 (the store counts every frame it ever took)", got)
	}
	if r.Base() != 2 {
		t.Fatalf("Base = %d, want 2", r.Base())
	}
	if got := len(r.GetFrames(2, 4)); got != 3 {
		t.Fatalf("GetFrames(2,4) = %d frames, want 3", got)
	}
	// A range reaching below the window is refused, not clamped: the fetch returns
	// nothing and both callers log that the observer cannot be updated.
	if got := r.GetFrames(0, 0); got != nil {
		t.Fatalf("GetFrames(0,0) = %v, want nil (evicted)", got)
	}
	if got := r.GetFrames(1, 3); got != nil {
		t.Fatalf("GetFrames(1,3) = %v, want nil (reaches below the window)", got)
	}
}

func TestReplayFrameCapDropsWholeBlocks(t *testing.T) {
	// maxFrames 3 with two frames per block: the cap is applied when a block is
	// opened and it drops whole blocks, so the window lands at 2, not at 3.
	r := NewReplay(8, 200, 3)
	for i := range 6 {
		r.AddFrame(0x0A, storeFrame(43, byte(i)))
	}
	if r.Base() == 0 {
		t.Fatal("Base = 0: a frame cap of 3 over 6 frames cannot leave the whole match")
	}
	if got := r.Total() - r.Base(); got > 3 {
		t.Fatalf("retained %d frames, want at most the cap of 3", got)
	}
}

func TestReplayGetFramesEdges(t *testing.T) {
	r := NewReplay(4, 65536, -1)
	if got := r.GetFrames(0, 3); got != nil {
		t.Errorf("empty store should return nil, got %v", got)
	}
	if r.Total() != 0 {
		t.Errorf("Total = %d, want 0", r.Total())
	}

	for i := range 3 {
		r.AddFrame(0x0A, storeFrame(43, byte(i)))
	}
	if r.Total() != 3 {
		t.Fatalf("Total = %d, want 3", r.Total())
	}

	if got := r.GetFrames(-1, 2); got != nil {
		t.Errorf("negative start should return nil, got %v", got)
	}
	if got := r.GetFrames(3, 5); got != nil {
		t.Errorf("last at or past the frame count should return nil, got %v", got)
	}
	if got := r.GetFrames(2, 1); got != nil {
		t.Errorf("end before start should return nil, got %v", got)
	}

	all := r.GetFrames(0, 2)
	if len(all) != 3 {
		t.Fatalf("GetFrames(0,2) = %d frames, want 3", len(all))
	}
	for i, f := range all {
		if f.Type != 0x0A {
			t.Errorf("frame %d kept type 0x%02X, want 0x0A: the store holds whole messages", i, f.Type)
		}
	}
}

func TestReplayCapacityFloor(t *testing.T) {
	r := NewReplay(0, 0, 0)
	if r == nil {
		t.Fatal("NewReplay(0) returned nil")
	}
	r.AddFrame(0x0A, []byte{0x01})
	if r.Total() != 1 {
		t.Errorf("Total = %d, want 1", r.Total())
	}
}

func TestReplayCopiesTheBody(t *testing.T) {
	r := NewReplay(4, 65536, -1)
	body := []byte{0x01, 0x02, 0x03}
	r.AddFrame(0x0A, body)
	body[0] = 0xFF
	frames := r.GetFrames(0, 0)
	if len(frames) != 1 || frames[0].Body[0] != 0x01 {
		t.Fatalf("stored frame = %v, want the bytes as they were at AddFrame time", frames)
	}
}

// --- session-level frame store ---

func TestSessionRecordFrameEvictsWholeBlocks(t *testing.T) {
	s := NewSession(1)
	s.maxPagesPerMatch = 3
	s.pageSize = 200

	// 3 blocks, 2 frames each. The seventh frame needs a fourth block, the cap is
	// 3, so block 0 is recycled and the window is [2..6].
	for i := range 7 {
		s.RecordFrame(0x0A, storeFrame(43, byte(i)))
	}
	if got := s.ReplayPageCount(); got != 5 {
		t.Fatalf("ReplayPageCount = %d, want 5", got)
	}
	if s.PageCount != 7 {
		t.Fatalf("PageCount = %d, want 7", s.PageCount)
	}
	if //goland:noinspection GoMaybeNil
	s.replay == nil || s.replay.Base() != 2 {
		t.Fatalf("Base = %d, want 2", s.replay.Base())
	}
}

// TestSessionRecordFrameDefaultHoldsAWholeMatch is the sizing that the second
// observe attempt of capture 2026-09-29_012528_03cb67c7cf19 depends on.
//
// The store is ceil(peers * f) blocks of pageSize bytes, 8 x 65536 = 512 KiB at
// the settings that run used. The original's whole 3m38s match is 259651 bytes of
// ack over 1570 frames, it fits in four blocks, and Base() therefore never moved:
// which is why the original's second observer, joining 171 s in, was still served
// from frame 0 and got 800 messages where this server used to give it 70.
func TestSessionRecordFrameDefaultHoldsAWholeMatch(t *testing.T) {
	s := NewSession(1)
	s.maxPagesPerMatch = 0
	s.pageSize = 0

	const matchFrames = 1570
	const matchBytes = 259651
	for range matchFrames {
		s.RecordFrame(0x0A, make([]byte, matchBytes/matchFrames-frameOverhead))
	}
	if got := s.replay.MaxPages(); got != 8*defaultPageSize {
		t.Fatalf("store window = %d bytes, want %d (8 blocks of the default page size)", got, 8*defaultPageSize)
	}
	if got := s.ReplayPageCount(); got != matchFrames {
		t.Fatalf("Retained = %d, want all %d: a match that fits must not be forgotten", got, matchFrames)
	}
	if s.replay.Base() != 0 {
		t.Fatalf("Base = %d, want 0 after a whole match that fits in the window", s.replay.Base())
	}
	if s.PageCount != matchFrames {
		t.Fatalf("PageCount = %d, want %d", s.PageCount, matchFrames)
	}
	if !s.Recording {
		t.Error("Recording = false: the store exists, so Recording is set")
	}
}

// --- Concurrent session (match) limit tests ---

func TestManagerMaxSessionsDeniesNew(t *testing.T) {
	mgr := NewManager(0, 2)
	s := mgr.CreateSession()
	if s == nil {
		t.Fatal("first CreateSession should succeed")
	}
	s2 := mgr.CreateSession()
	if s2 == nil {
		t.Fatal("second CreateSession should succeed")
	}
	if got := mgr.SessionCount(); got != 2 {
		t.Fatalf("SessionCount = %d, want 2", got)
	}
	// Next CreateSession should be denied.
	if s3 := mgr.CreateSession(); s3 != nil {
		t.Fatal("CreateSession should return nil when max session limit reached")
	}
	// A session already tracked can still be fetched.
	if got := mgr.GetSession(SessionKey{SessionID: s.ID64()}); got != s {
		t.Fatal("existing session should still be reachable")
	}
}

func TestManagerGetOrCreateSessionDeniedForNew(t *testing.T) {
	mgr := NewManager(0, 1)
	s := mgr.CreateSession()
	// Existing session should still be returned.
	if got := mgr.GetOrCreateSession(s.Key()); got == nil {
		t.Fatal("GetOrCreateSession should return existing session")
	}
	// New session ID should be denied.
	if got := mgr.GetOrCreateSession(SessionKey{SessionID: 999}); got != nil {
		t.Fatal("GetOrCreateSession should return nil for new session when max session limit reached")
	}
}

func TestManagerRemoveSessionFreesSlot(t *testing.T) {
	mgr := NewManager(0, 1)
	s := mgr.CreateSession()
	if s == nil {
		t.Fatal("first session should succeed")
	}
	mgr.RemoveSession(s.Key())
	// A new session should now succeed.
	if s2 := mgr.CreateSession(); s2 == nil {
		t.Fatal("CreateSession should succeed after removing the session")
	}
}

func TestManagerMaxSessionsAcrossSessions(t *testing.T) {
	mgr := NewManager(0, 2)
	s1 := mgr.CreateSession()
	s2 := mgr.CreateSession()
	if s1 == nil || s2 == nil {
		t.Fatal("first two sessions should succeed")
	}
	// Third session denied.
	if s3 := mgr.CreateSession(); s3 != nil {
		t.Fatal("third session should be denied")
	}
	// Remove one session, freeing a slot.
	mgr.RemoveSession(s1.Key())
	s3 := mgr.CreateSession()
	if s3 == nil {
		t.Fatal("CreateSession should succeed after removing a session")
	}
}

// TestManagerPerMatchBlockCap covers the same eviction through the Manager,
// which is where the operator's -maxPagesToStorePerSession reaches a session.
// The flag is the BLOCK count (ceil(peers * value/8), with peers floored at 8),
// so 3 gives three blocks, and the page size has to be small for seven frames to
// fill them.
func TestManagerPerMatchBlockCap(t *testing.T) {
	mgr := NewManagerWithPageSize(3, 0, 200)
	s := mgr.CreateSession()
	for i := range 7 {
		s.RecordFrame(0x0A, storeFrame(43, byte(i)))
	}
	if got := s.ReplayPageCount(); got != 5 {
		t.Fatalf("ReplayPageCount = %d, want 5", got)
	}
	if //goland:noinspection GoMaybeNil
	s.replay == nil || s.replay.Base() != 2 {
		t.Fatalf("Base = %d, want 2", s.replay.Base())
	}
}

func TestManagerCreateSession(t *testing.T) {
	mgr := NewManager(0, 0)
	s := mgr.CreateSession()
	if s == nil {
		t.Fatal("CreateSession returned nil")
	}
	if s.ID64() != 1 {
		t.Errorf("first session ID = %d, want 1", s.ID64())
	}
	if mgr.SessionCount() != 1 {
		t.Errorf("SessionCount = %d, want 1", mgr.SessionCount())
	}
}

// Sessions are keyed on the pair (titleID, sessionID), the way the original's
// ordered tree is. The same sessionID under two titles is
// two sessions, which matters now that one process can host several titles.
func TestManagerKeysSessionsByTitleAndID(t *testing.T) {
	mgr := NewManager(0, 0)

	a := mgr.GetOrCreateSession(SessionKey{TitleID: 2, SessionID: 55})
	b := mgr.GetOrCreateSession(SessionKey{TitleID: 3, SessionID: 55})
	if a == nil || b == nil {
		t.Fatal("GetOrCreateSession returned nil")
	}
	if a == b {
		t.Fatal("the same sessionID under two titles must be two sessions")
	}
	if mgr.SessionCount() != 2 {
		t.Errorf("SessionCount = %d, want 2", mgr.SessionCount())
	}

	// The same pair returns the same session.
	if got := mgr.GetOrCreateSession(SessionKey{TitleID: 2, SessionID: 55}); got != a {
		t.Error("the same (titleID, sessionID) must return the same session")
	}
	if got := mgr.GetSession(SessionKey{TitleID: 3, SessionID: 55}); got != b {
		t.Error("lookup by the wrong titleID must not find the session")
	}
	if got := mgr.GetSession(SessionKey{TitleID: 0, SessionID: 55}); got != nil {
		t.Error("a third titleID with the same id must not resolve")
	}

	// The 0xFFFF marker the legacy StartObserve form carries is -1 read as a
	// signed 16-bit value, and no session is keyed that way.
	if got := mgr.GetSession(SessionKey{TitleID: -1, SessionID: 55}); got != nil {
		t.Error("the legacy 0xFFFF titleID marker must not resolve to a session")
	}

	// Removing one leaves the other.
	mgr.RemoveSession(a.Key())
	if mgr.GetSession(a.Key()) != nil {
		t.Error("removed session still resolves")
	}
	if mgr.GetSession(b.Key()) != b {
		t.Error("removing one session removed the other")
	}
}

func TestManagerGetSessions(t *testing.T) {
	mgr := NewManager(0, 0)
	if got := mgr.GetSessions(); len(got) != 0 {
		t.Errorf("expected no sessions, got %d", len(got))
	}

	mgr.CreateSession()
	mgr.CreateSession()
	mgr.CreateSession()

	got := mgr.GetSessions()
	if len(got) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(got))
	}
}

func TestManagerGetSessionNotFound(t *testing.T) {
	mgr := NewManager(0, 0)
	if s := mgr.GetSession(SessionKey{SessionID: 42}); s != nil {
		t.Errorf("expected nil for unknown session, got %v", s)
	}
}

func TestManagerRemoveSessionMissing(t *testing.T) {
	mgr := NewManager(0, 0)
	mgr.CreateSession()
	mgr.RemoveSession(SessionKey{SessionID: 999})

	if mgr.SessionCount() != 1 {
		t.Errorf("SessionCount = %d, want 1", mgr.SessionCount())
	}
}

func TestManagerRemoveSessionCloses(t *testing.T) {
	mgr := NewManager(0, 0)
	s := mgr.CreateSession()
	conn := &mockConn{}
	s.AddPeer(NewPeer(1, conn))

	mgr.RemoveSession(s.Key())

	if !s.IsClosed() {
		t.Error("removed session should be closed")
	}
	if !conn.closed {
		t.Error("peer connection should be closed")
	}
}

func TestManagerFindSessionByPeerNotFound(t *testing.T) {
	mgr := NewManager(0, 0)
	mgr.CreateSession()
	if s := mgr.FindSessionByPeer(&mockConn{}); s != nil {
		t.Errorf("expected nil, got %v", s)
	}
}

func TestManagerHandleLeaveSimulationNotFound(t *testing.T) {
	mgr := NewManager(0, 0)
	mgr.CreateSession()

	mgr.HandleLeaveSimulation(&mockConn{})

	if mgr.SessionCount() != 1 {
		t.Errorf("SessionCount = %d, want 1", mgr.SessionCount())
	}
}

// TestManagerHandleLeaveSimulationKeepsThePeer pins that a LeaveSimulation does
// NOT take the peer out of the session. Its handler is three lines
// and forwards to the leave routine, which empties the leaver's two queues, clears
// its Participating mark and can end the recording: it never touches the session's
// peer vector. The only thing that removes a peer is the disconnect path.
//
// It used to assert the opposite, from the message's name.
func TestManagerHandleLeaveSimulationKeepsThePeer(t *testing.T) {
	mgr := NewManager(0, 0)
	s := mgr.CreateSession()
	conn1 := &mockConn{}
	conn2 := &mockConn{}
	mgr.AddPeer(s, NewPeer(1, conn1))
	mgr.AddPeer(s, NewPeer(2, conn2))
	// The frame store is what marks every peer as being in the started game
	// and the LeaveSimulation only has a mark to clear
	// because of it.
	s.RecordFrame(0x0A, []byte{0xfe, 0x00, 0x01, 0x01, 0x00, 0x00})

	mgr.HandleLeaveSimulation(conn1)

	if mgr.SessionCount() != 1 {
		t.Errorf("SessionCount = %d, want 1", mgr.SessionCount())
	}
	if s.PeerCount() != 2 {
		t.Errorf("PeerCount = %d, want 2: a LeaveSimulation does not remove the peer", s.PeerCount())
	}
	if p := s.GetPeer(1); p == nil {
		t.Error("peer 1 should still be a member of the session after a LeaveSimulation")
	} else if p.Participating {
		t.Error("peer 1 should not be marked as participating after its LeaveSimulation")
	}
	if p := s.GetPeer(2); p == nil || !p.Participating {
		t.Error("peer 2 is still in the match and must keep its Participating mark")
	}
}

// BenchmarkReplayGetFrames measures the observer catch-up fetch. It
// returns a sub-slice view with no allocation, which is safe because a stored
// frame is never modified once in.
func BenchmarkReplayGetFrames(b *testing.B) {
	for _, frames := range []int{1000, 5000} {
		for _, span := range []int{1, 16, 200} {
			b.Run(fmt.Sprintf("frames=%d span=%d", frames, span), func(b *testing.B) {
				r := NewReplay(frames, 1<<20, -1)
				body := make([]byte, 256)
				for range frames {
					r.AddFrame(0x0A, body)
				}
				start := frames / 2
				end := start + span - 1
				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					_ = r.GetFrames(start, end)
				}
			})
		}
	}
}

// BenchmarkManagerFindSessionByPeer measures the relay session-route lookup
// (FindSessionByPeer on every relayed message). Before P3 it scanned every
// session and its peers (O(sessions x peers)); after it is an O(1) index lookup.
// The target is a peer of the LAST session, the worst case for the old linear
// scan.
func BenchmarkManagerFindSessionByPeer(b *testing.B) {
	for _, sessions := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("sessions=%d", sessions), func(b *testing.B) {
			mgr := NewManager(0, 0)
			var last *Session
			for range sessions {
				sess := mgr.CreateSession()
				mgr.AddPeer(sess, NewPeer(1, &mockConn{}))
				mgr.AddPeer(sess, NewPeer(2, &mockConn{}))
				last = sess
			}
			conn := last.GetPeer(2).Conn
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if mgr.FindSessionByPeer(conn) == nil {
					b.Fatal("not found")
				}
			}
		})
	}
}
