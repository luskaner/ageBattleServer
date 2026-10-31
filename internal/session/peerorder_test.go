package session

import (
	"testing"
)

// WHAT THIS FILE IS
//
// Session.Peers is a slot table and every walk of the peers goes through
// peerOrder, the order AddPeer was called in. This file pins the two facts that
// makes of a walk, because both of them are observable and neither of them is a
// matter of taste:
//
//   - The walk is JOIN order, and it is not slot order. The original's vector of
//     game-data queue records is pushed in join order
//     and everything that reads the peers reads that
//     vector: the 0x0A's pending list and the fallback
//     secret scan (which steps the vector eight bytes at a
//     time and takes the FIRST match). PendingPeersAlreadyFollowsPeerAddOrder
//     below says the pending list does not read the walk, so the two cannot
//     disagree; this file says what the walk itself is.
//
//   - The walk is NOT what a client sees. Broadcast and the ack send loop hand
//     every peer the SAME body, so the order in which the peers are visited can
//     only decide which of two connections sees a byte first, and each
//     connection's own stream is the same whatever the order is. That is the
//     claim that let the peer table become the original's [256] slot array
//     (SESSION-JOURNAL.md 39.7 left it as a map because nothing had shown it).
//     TestEveryPeerGetsTheSameStreamWhateverTheWalkOrder is the falsifier: with
//     the peers walked in an order that is neither the slot order nor the join
//     order, every connection still has to carry the same bytes in the same
//     order.
//
// The falsification of the order claim itself is the defect put back by hand:
// with the walk reading the 256 slots ascending, PendingWalksInJoinOrder fails
// on its first assertion (2 before 1), and with the walk ranging a map it fails
// on half the runs, which is what "map order" means.

// pendingWalksInJoinOrder is the shared fixture: peers whose ids were handed out
// in an order that is NOT ascending, so join order and slot order disagree.
func pendingWalksInJoinOrder(t *testing.T) *Session {
	t.Helper()
	s := joinedIn(t, 1, 2, 1, 3)
	if got := s.PeerCount(); got != 3 {
		t.Fatalf("a session with three peers reports %d", got)
	}
	return s
}

// PendingWalksInJoinOrder is the order every walk of the peers has to have.
func TestWalksFollowPeerAddOrderNotSlotOrder(t *testing.T) {
	s := pendingWalksInJoinOrder(t)

	got := make([]byte, 0, 3)
	for _, p := range s.GetPeers() {
		got = append(got, p.ID)
	}
	if want := []byte{2, 1, 3}; !equalBytes(got, want) {
		t.Errorf("GetPeers walked %v, want the join order %v. The peers joined as 2, 1, 3 and the original's queue "+
			"vector is in join order, not slot order", got, want)
	}

	var each []byte
	s.ForEachPeer(func(p *Peer) { each = append(each, p.ID) })
	if !equalBytes(each, got) {
		t.Errorf("ForEachPeer walked %v and GetPeers walked %v. There is one walk (peerOrder) and these are two views "+
			"of it", each, got)
	}
}

// The slot table is what answers "is there a peer in this slot", and peerOrder is
// the membership. A slot that was emptied has to read empty and has to leave the
// walk; a peer that comes back to the same slot goes to the end of the walk,
// because it joined again.
func TestSlotTableAndWalkAreTheSameMembership(t *testing.T) {
	s := pendingWalksInJoinOrder(t)
	if p := s.Peers[2]; p == nil || p.ID != 2 {
		t.Fatal("slot 2 does not hold peer 2: the table is indexed by player slot, as the original's is")
	}
	if p := s.Peers[0]; p != nil {
		t.Errorf("slot 0 holds peer %d: peer ids are 1-based and 0 is reserved", p.ID)
	}
	if p := s.Peers[200]; p != nil {
		t.Errorf("slot 200 holds peer %d: only the peers of this session are in the table", p.ID)
	}

	s.RemovePeer(1)
	if s.Peers[1] != nil {
		t.Error("slot 1 still holds a peer after RemovePeer")
	}
	if got := s.PeerCount(); got != 2 {
		t.Errorf("PeerCount = %d after removing one of three peers, want 2. The count is len(peerOrder), and a slot "+
			"table with 255 empty slots cannot answer it", got)
	}
	var after []byte
	s.ForEachPeer(func(p *Peer) { after = append(after, p.ID) })
	if want := []byte{2, 3}; !equalBytes(after, want) {
		t.Errorf("after removing peer 1 the walk is %v, want %v", after, want)
	}

	// Back into the same slot: the record is pushed again, so it is last.
	s.AddPeer(NewPeer(1, &mockConn{}))
	s.ForEachPeer(func(p *Peer) { after = append(after, p.ID) })
	if want := []byte{2, 3, 1}; !equalBytes(after[len(after)-3:], want) {
		t.Errorf("after peer 1 rejoined the walk is %v, want %v at the end", after, want)
	}
}

// PeerBySecret is the reconnect lookup, and two peers carrying the same secret
// is the only case where the order of the walk decides the answer. The original
// decides it by vector position: the lookup walks eight bytes at a time
// and returns the first peer whose secret is the one asked for, so the peer that
// joined first is the one that gets the socket swapped.
func TestPeerBySecretTakesTheFirstPeerThatJoined(t *testing.T) {
	const secret = 0xFEEDFACECAFEBEEF
	s := joinedIn(t, 1, 2, 1, 3)
	for _, id := range []byte{2, 1, 3} {
		if !s.SetPeerSecret(id, secret) {
			t.Fatalf("SetPeerSecret(%d) refused", id)
		}
	}
	p := s.PeerBySecret(secret)
	if p == nil {
		t.Fatal("PeerBySecret found nobody: three peers carry the secret")
	}
	if p.ID != 2 {
		t.Errorf("PeerBySecret returned peer %d, want peer 2. The original's scan takes the first peer of its vector "+
			"and the vector is in join order, so this is peer 2 and not a per-run coin toss", p.ID)
	}
}

// PeerByKey is the other search by a peer field. The original's table is keyed by
// the identity itself, so two peers with
// the same key cannot exist there; here they can, because the key is a field. When
// they do, the walk has to answer the same way every run.
func TestPeerByKeyTakesTheFirstPeerThatJoined(t *testing.T) {
	const key = 0x1122334455667788
	s := joinedIn(t, 1, 3, 1)
	for _, id := range []byte{3, 1} {
		if !s.SetPeerKey(id, key) {
			t.Fatalf("SetPeerKey(%d) refused", id)
		}
	}
	p := s.PeerByKey(key)
	if p == nil {
		t.Fatal("PeerByKey found nobody: two peers carry the key")
	}
	if p.ID != 3 {
		t.Errorf("PeerByKey returned peer %d, want peer 3, the one that joined first", p.ID)
	}
}

// TestEveryPeerGetsTheSameStreamWhateverTheWalkOrder is the claim that makes the
// walk order unobservable, and it is stated as output rather than as argument.
//
// Broadcast and the ack send loop hand the peers the same body. Two peers on two
// connections therefore see the same bytes in the same per-connection order, and
// the order in which the walk visits them decides nothing a client can read: it
// decides which connection sees a byte first, and each connection's stream is
// compared on its own.
//
// The fixture adds the peers in an order that is neither ascending nor the walk
// order, so a walk that visited the slots, or one that visited them in reverse,
// would produce a different visiting order here than the code under test does.
// What it may NOT produce is a different stream.
func TestEveryPeerGetsTheSameStreamWhateverTheWalkOrder(t *testing.T) {
	conns := map[byte]*mockConn{}
	s := NewSession(1)
	for _, id := range []byte{3, 1, 2} {
		c := &mockConn{}
		conns[id] = c
		s.AddPeer(NewPeer(id, c))
	}

	bodies := [][]byte{{0x01, 0x02}, {0x03, 0x04, 0x05}}
	for _, body := range bodies {
		s.Broadcast(0x09, body, 2)
	}

	for id, c := range conns {
		var want []mockMsg
		for _, body := range bodies {
			if id == 2 {
				continue // the excluded peer
			}
			want = append(want, mockMsg{typ: 0x09, body: body})
		}
		if len(c.msgs) != len(want) {
			t.Fatalf("peer %d received %d message(s), want %d", id, len(c.msgs), len(want))
		}
		for i := range want {
			if c.msgs[i].typ != want[i].typ || !equalBytes(c.msgs[i].body, want[i].body) {
				t.Errorf("peer %d message #%d is %v, want %v. Every peer is sent the same body, so what the order of "+
					"the walk decides is which connection sees it first and not what any of them receives", id, i,
					c.msgs[i], want[i])
			}
		}
	}
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}