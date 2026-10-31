package session

import (
	"bytes"
	"testing"
	"time"
)

// The pending-peer list of the 0x0A, pinned against the drain that builds it.
//
// The whole list is one walk over the session's
// game-data queue records, and every question about it is decided
// there: which records exist, what an "empty queue" is, and which peer each
// record names. The three tests below are the three conditions that walk applies
// per record, plus the order it emits them in.
//
// What it is NOT, contrary to what the field audit said: there is no cut. The
// audit read the skip's jump target as the end of the list and wrote that the walk
// "stops at the first record whose peer is missing or not participating". That
// target is in fact the bottom of the loop, which advances to the next record and
// tests the count. A record that
// fails the test is skipped and the walk carries on.

// joinedIn builds a session whose peers joined in the given order, each with its
// own connection, and returns it with the connections.
func joinedIn(t *testing.T, id int64, ids ...byte) *Session {
	t.Helper()
	s := NewSession(id)
	for _, pid := range ids {
		s.AddPeer(NewPeer(pid, &mockConn{}))
	}
	return s
}

// started forces the state the drain reads: game data has arrived, so the store
// exists and every peer present carries the Participating mark.
func started(s *Session, from byte) {
	s.QueueGameData(from, []byte{0x00, 0x00})
}

// The list is the queue records in the vector's order, and the vector is in
// PEER-ADD order (a record is pushed when the peer joins
// and erased when it disconnects). With the peer ids handed
// out in join order the two orders coincide and this is indistinguishable from
// sorting by slot; with them the other way round it is not.
func TestPendingListFollowsPeerAddOrderNotSlotOrder(t *testing.T) {
	s := joinedIn(t, 1, 2, 1)
	started(s, 2)
	// Peer 2 uploaded, so the first ack flushes it and lists the other record.
	if body, _ := s.PopAckTick(); !bytes.Equal(body, []byte{0xfe, 0x01, 0x02, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00}) {
		t.Fatalf("the flush must name slot 2 first (the record is found by id, the walk is by "+
			"position): got %x", body)
	}

	// Nothing queued: both records are pending, in the order they joined.
	body, _ := s.PopAckTick()
	if !bytes.Equal(body, []byte{0xfe, 0x00, 0x02, 0x02, 0x01, 0x00, 0x00}) {
		t.Fatalf("pending list = %x, want the records in join order (2 then 1)", body)
	}
}

// A record whose peer is not participating is skipped, and the walk CONTINUES:
// the records behind it are still walked and their peers still listed. This is one
// record skipped, not the list cut. The audit read
// the skip's jump target as the end of the loop, and it is the bottom
// of it.
//
// A LeaveSimulation is what puts a record in the middle of the vector with its
// peer clear: the leave writes Participating back to 0 and empties the two
// queues, and it removes nothing, so the record keeps its slot and its position.
func TestPendingListKeepsGoingPastANonParticipatingPeer(t *testing.T) {
	s := joinedIn(t, 1, 1, 2, 3)
	started(s, 1)
	if body, _ := s.PopAckTick(); body == nil {
		t.Fatal("no ack built")
	}
	if !s.LeaveSimulation(2) {
		t.Fatal("LeaveSimulation did not find peer 2")
	}
	if s.GetPeer(2) == nil {
		t.Fatal("the record's owner must stay in the session: the leave removes nothing")
	}
	if s.GetPeer(2).Participating {
		t.Fatal("Participating must be clear for the leaver")
	}

	// Peer 1 and peer 3 are participating and silent, peer 2 is skipped.
	body, _ := s.PopAckTick()
	if !bytes.Equal(body, []byte{0xfe, 0x00, 0x02, 0x01, 0x03, 0x00, 0x00}) {
		t.Fatalf("pending list = %x, want [1 3]: the walk skips peer 2's record and carries on with peer 3's", body)
	}
}

// The other way a record's peer comes to be clear: the mark is set once, when the
// store appears, and only for the peers already there, so a peer that joins a
// running match never gets
// it and its record is skipped too.
func TestPendingListSkipsAPeerThatJoinedAfterTheStoreExists(t *testing.T) {
	s := joinedIn(t, 1, 1, 2)
	started(s, 1)
	if body, _ := s.PopAckTick(); body == nil {
		t.Fatal("no ack built")
	}
	s.AddPeer(NewPeer(3, &mockConn{}))
	if s.GetPeer(3).Participating {
		t.Fatal("a peer that joins after the store exists must not be marked Participating")
	}

	body, _ := s.PopAckTick()
	if !bytes.Equal(body, []byte{0xfe, 0x00, 0x02, 0x01, 0x02, 0x00, 0x00}) {
		t.Fatalf("pending list = %x, want [1 2]: the peer that joined late has no mark", body)
	}
}

// The same record, carrying data. A record that fails the test is skipped whether
// it is empty or not, and the record behind it still flushes.
func TestFlushSkipsANonParticipatingPeerAndKeepsGoing(t *testing.T) {
	s := joinedIn(t, 1, 1)
	started(s, 1)
	if body, _ := s.PopAckTick(); body == nil {
		t.Fatal("no ack built")
	}
	s.AddPeer(NewPeer(2, &mockConn{}))
	s.QueueGameData(1, []byte{0x02, 0x00, 0x11, 0x22})
	s.QueueGameData(2, []byte{0x02, 0x00, 0x33, 0x44})

	// Peer 2's upload is not even queued: the original drops it before the queue
	// when the peer is not participating, so its record has
	// nothing and its peer is skipped anyway. What this pins is that peer 1's
	// record, which sits BEFORE peer 2's, still flushes.
	body, _ := s.PopAckTick()
	if !bytes.Contains(body, []byte{0x11, 0x22}) {
		t.Fatalf("the participating peer's data is missing from the ack %x", body)
	}
}

// The record whose peer is late is closed and NOT listed: same branch, same pass.
// The walk schedules the close instead of appending the id.
func TestPendingListOmitsThePeerBeingReapedForSyncTimeout(t *testing.T) {
	s := joinedIn(t, 1, 1, 2)
	s.SetPeerTimeoutOverride(1, 1) // 1 ms deadline
	started(s, 1)
	// The first tick flushes peer 1, which leaves its queue empty and stamps its
	// last sync; peer 2 is listed.
	if body, _ := s.PopAckTick(); !bytes.Equal(body, []byte{0xfe, 0x01, 0x01, 0x00, 0x00, 0x01, 0x02, 0x00, 0x00}) {
		t.Fatalf("first ack = %x, want the flush with peer 2 pending", body)
	}
	time.Sleep(20 * time.Millisecond)

	if got := s.SyncTimedOutPeers(time.Now()); len(got) != 1 || got[0] != 1 {
		t.Fatalf("SyncTimedOutPeers = %v, want [1]", got)
	}
	// The next tick has nothing to flush, and peer 1 is past its deadline: the
	// list is peer 2 alone.
	body, _ := s.PopAckTick()
	if !bytes.Equal(body, []byte{0xfe, 0x00, 0x01, 0x02, 0x00, 0x00}) {
		t.Fatalf("pending list = %x, want [2]: the peer past its deadline is closed, not listed", body)
	}
}

// The store and the participating marks go up BEFORE the ack of that tick is
// built, so the FIRST ack of a match lists
// every peer. Building the body first and marking afterwards left this one ack
// with an empty pending list, because nothing was participating yet.
func TestFirstAckAfterTheFirstUploadListsEveryPeer(t *testing.T) {
	s := joinedIn(t, 1, 1, 2)
	started(s, 1)
	body, _ := s.PopAckTick()
	if !bytes.Equal(body, []byte{0xfe, 0x01, 0x01, 0x00, 0x00, 0x01, 0x02, 0x00, 0x00}) {
		t.Fatalf("first ack = %x, want peer 1 flushed and peer 2 pending", body)
	}
	if !s.Recording {
		t.Error("the store should exist by the time the first ack goes out")
	}
	for _, id := range []byte{1, 2} {
		if p := s.GetPeer(id); p == nil || !p.Participating {
			t.Errorf("peer %d should be Participating before the first ack is built", id)
		}
	}
}

// The flushed chunks come out of the same walk, so they are in record order too:
// every pass takes one chunk from each record in that order, and each record's
// own chunks come out in arrival order.
func TestFlushSubMessagesFollowPeerAddOrder(t *testing.T) {
	s := joinedIn(t, 1, 2, 1)
	started(s, 2)
	// Peer 1 uploads twice and peer 2 once, all before the tick. Each argument is
	// the client's own [len u16 LE][chunk].
	s.QueueGameData(1, []byte{0x02, 0x00, 0x11, 0x11})
	s.QueueGameData(1, []byte{0x02, 0x00, 0x12, 0x12})
	s.QueueGameData(2, []byte{0x02, 0x00, 0x22, 0x22})

	body, _ := s.PopAckTick()
	want := []byte{0xfe, 0x04}
	want = append(want, 0x02, 0x00, 0x00)             // pass 1: peer 2's record comes first, its own chunk
	want = append(want, 0x01, 0x02, 0x00, 0x11, 0x11) // pass 1: then peer 1's first
	want = append(want, 0x02, 0x02, 0x00, 0x22, 0x22) // pass 2: peer 2's second
	want = append(want, 0x01, 0x02, 0x00, 0x12, 0x12) // pass 2: peer 1's second
	want = append(want, 0x00, 0x00, 0x00)             // empty pending list, empty block
	if !bytes.Equal(body, want) {
		t.Fatalf("flush = %x,\nwant     %x", body, want)
	}
}

// A peer's queued game data dies with the peer: the disconnect erases its record
// and the chunks in it are freed with it, so what it uploaded and nobody flushed is
// never sent to the
// peers that are left.
func TestRemovingAPeerDropsItsQueuedGameData(t *testing.T) {
	s := joinedIn(t, 1, 1, 2)
	started(s, 2)
	if body, _ := s.PopAckTick(); body == nil {
		t.Fatal("no ack built")
	}
	s.QueueGameData(1, []byte{0x02, 0x00, 0x11, 0x22})
	s.QueueGameData(2, []byte{0x02, 0x00, 0x33, 0x44})
	s.RemovePeer(2)

	body, _ := s.PopAckTick()
	if bytes.Contains(body, []byte{0x33, 0x44}) {
		t.Fatalf("the ack carries data from a peer that disconnected: %x", body)
	}
	if !bytes.Contains(body, []byte{0x11, 0x22}) {
		t.Fatalf("the remaining peer's data is missing from the ack %x", body)
	}
}