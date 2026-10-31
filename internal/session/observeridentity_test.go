package session

import "testing"

// THE OBSERVER'S IDENTIFIER: the list is ordered by the peerIdentifier, not by
// the connection.
//
// The field is the third u64 of the StartObserveRequest Vx, at offset 9 of the body
// (written by the request parser and
// named by the original's own log). The connection keeps
// it at its first field, and in the whole original that field has two writers and
// two readers: the connection's constructor
// sets it to 0, the arm that registers an observer writes the request's value into
// it, the activate walks the list comparing the caller's field against each
// entry's, and the stop does the same and stops at the first match.
//
// The comparator is an equality over a u64, inside the session of the connection
// itself (the activate and the stop both take the session from
// the connection), and the FIRST entry that matches wins. There is no peerID, no
// address, no second key. These tests are the counter-proof of the five questions
// that comparator raises, and each one looks at state with data on both sides: two
// entries, two activations, two different Progress values.

// identSession is a session with a store of 600 frames, one observer slot each
// identifier's worth of room, and the update interval at 1 so every Tick serves
// every entry. 600 frames of 43 bytes is 25 KB: bounded here in the code and far
// below the budget, and enough that a 200-frame visit is not
// the whole store.
func identSession(maxObservers, frames int) *Session {
	s := NewSession(1)
	s.SetObserverConfig(maxObservers, 1)
	s.SetReconnectable(true)
	recordAcks(s, frames)
	return s
}

const (
	// identA and identB are the peerIdentifiers of the two observers of
	// 2026-09-29_012528_03cb67c7cf19, read at offset 9 of their two 0x1C bodies.
	identA uint64 = 0xC5C0428C2971CE8B
	identB uint64 = 0xA9AE099C85ED50C4
)

// 1. TWO CONNECTIONS, DISTINCT IDENTIFIERS: each one finds ITS entry.
//
// This is the case the four real captures already cover (a different identifier
// per observer), and it is the one that must not break when the key changes. The
// assertion looks at BOTH sides: the two 0x1E go to whoever asked for them, and
// each entry ends up active with its own progress.
func TestDistinctIdentifiersFindTheirOwnEntry(t *testing.T) {
	s := identSession(2, 600)

	a, b := &mockConn{}, &mockConn{}
	if status, registered := s.StartObserve(a, identA, 0); status != 0 || !registered {
		t.Fatalf("StartObserve(a) = %d/%v, want 0/true", status, registered)
	}
	if status, registered := s.StartObserve(b, identB, 0); status != 0 || !registered {
		t.Fatalf("StartObserve(b) = %d/%v, want 0/true", status, registered)
	}
	if s.ObserverCount() != 2 {
		t.Fatalf("ObserverCount = %d, want 2: two identifiers are two entries", s.ObserverCount())
	}
	if a.PeerIdentifier() != identA || b.PeerIdentifier() != identB {
		t.Fatalf("identifiers on the connections = %#x/%#x, want %#x/%#x",
			a.PeerIdentifier(), b.PeerIdentifier(), identA, identB)
	}
	if !s.ActivateObserver(a, 8) || !s.ActivateObserver(b, 8) {
		t.Fatal("both activations should find their entry")
	}
	// Each 0x1E went to the connection that asked for it, and to no one else.
	if countMsgs(a.msgs, 0x1E) != 1 || countMsgs(b.msgs, 0x1E) != 1 {
		t.Fatalf("0x1E counts = %d/%d, want 1/1", countMsgs(a.msgs, 0x1E), countMsgs(b.msgs, 0x1E))
	}
	if !s.observers[0].Active || !s.observers[1].Active {
		t.Fatalf("both entries should be active, got %v/%v", s.observers[0].Active, s.observers[1].Active)
	}

	// And the two entries are served independently: the first 200 frames each,
	// from the same store, to their own connection.
	s.Tick()
	if got := countMsgs(a.msgs, ackFrameType); got != 200 {
		t.Errorf("a got %d frames on the first visit, want 200", got)
	}
	if got := countMsgs(b.msgs, ackFrameType); got != 200 {
		t.Errorf("b got %d frames on the first visit, want 200", got)
	}
	if s.observers[0].Progress != 200 || s.observers[1].Progress != 200 {
		t.Errorf("progresses = %d/%d, want 200/200", s.observers[0].Progress, s.observers[1].Progress)
	}
}

// 2. THE SAME IDENTIFIER TWICE: the original pairs the entry it finds first and
// sends the 0x1E to whoever asked.
//
// The activate handler arms the entry that matches and answers the argument, that
// is the connection that asked. When both match, the entry armed
// is A's and the answer goes to B: two different facts, which is why both are
// checked.
func TestSameIdentifierArmsTheFirstEntryAndAnswersTheCaller(t *testing.T) {
	s := identSession(2, 600)

	a, b := &mockConn{}, &mockConn{}
	if status, _ := s.StartObserve(a, identA, 0); status != 0 {
		t.Fatalf("StartObserve(a) = %d, want 0", status)
	}
	if status, _ := s.StartObserve(b, identA, 0); status != 0 {
		t.Fatalf("StartObserve(b) = %d, want 0", status)
	}
	if s.ObserverCount() != 2 {
		t.Fatalf("ObserverCount = %d, want 2: the original appends without asking", s.ObserverCount())
	}

	// B asks to be activated. The FIRST entry (A's) is the one that matches.
	if !s.ActivateObserver(b, 8) {
		t.Fatal("ActivateObserver from the second connection should still find a match")
	}
	if !s.observers[0].Active {
		t.Error("the first entry should be the armed one (the activate walks from the head)")
	}
	if s.observers[1].Active {
		t.Error("the second entry must stay inactive: its connection is the caller, not a separate match")
	}
	// And the 0x1E went to B, the caller, not to A.
	if countMsgs(b.msgs, 0x1E) != 1 {
		t.Errorf("the caller got %d 0x1E, want 1", countMsgs(b.msgs, 0x1E))
	}
	if countMsgs(a.msgs, 0x1E) != 0 {
		t.Errorf("the other connection got %d 0x1E, want 0: the answer goes to the caller", countMsgs(a.msgs, 0x1E))
	}
	// A is served and B is not, because A's entry is the active one.
	s.Tick()
	if got := countMsgs(a.msgs, ackFrameType); got != 200 {
		t.Errorf("the armed entry's connection got %d frames, want 200", got)
	}
	if got := countMsgs(b.msgs, ackFrameType); got != 0 {
		t.Errorf("the inactive entry's connection got %d frames, want 0", got)
	}
}

// 2b. THE SAME IDENTIFIER AND StopObserve: the FIRST entry is erased.
//
// The stop uses the SAME comparator as the activate, and the removal
// takes out the entry that matched, not the one of the caller. The connection the
// original closes afterwards is the one it was given, not
// the entry's: which is why RemoveObserver returns and leaves the close to the
// transport.
func TestStopObserveRemovesTheFirstEntryOfASharedIdentifier(t *testing.T) {
	s := identSession(2, 600)

	a, b := &mockConn{}, &mockConn{}
	if status, _ := s.StartObserve(a, identA, 0); status != 0 {
		t.Fatalf("StartObserve(a) = %d, want 0", status)
	}
	if status, _ := s.StartObserve(b, identA, 0); status != 0 {
		t.Fatalf("StartObserve(b) = %d, want 0", status)
	}

	if !s.RemoveObserver(b) {
		t.Fatal("StopObserve should find a match to remove")
	}
	if s.ObserverCount() != 1 {
		t.Fatalf("ObserverCount = %d, want 1", s.ObserverCount())
	}
	// The entry that SURVIVED is B's, not A's: B registered second, so removing
	// the first match removed A's entry. Both connections carry the same
	// identifier, so the survivor is told apart by its position in the list.
	if s.observers[0].Client != PeerConnection(b) {
		t.Error("the wrong entry survived: the first match (A's) is the one the original removes")
	}
	// And the removal resolves like the lookup: B is still known, A is not.
	if !s.HasObserver(b) {
		t.Error("the surviving entry should still be found by the shared identifier")
	}
	if s.ActivateObserver(b, 8) == false {
		t.Error("the survivor should still activate")
	}
	if s.ActivateObserver(&mockConn{peerID: identA + 1}, 8) {
		t.Error("a connection with another identifier must not match")
	}
}

// 3. StopObserve WITH DISTINCT IDENTIFIERS: each one removes its own.
//
// This is the normal session's case, and it is the one a "lookup by identifier
// but erased by identity" would break silently: both entries stay and the list
// never empties. It is checked with the state of both entries, not only with the
// return value.
func TestStopObserveResolvesLikeTheLookup(t *testing.T) {
	s := identSession(2, 600)

	a, b := &mockConn{}, &mockConn{}
	if status, _ := s.StartObserve(a, identA, 0); status != 0 {
		t.Fatalf("StartObserve(a) = %d, want 0", status)
	}
	if status, _ := s.StartObserve(b, identB, 0); status != 0 {
		t.Fatalf("StartObserve(b) = %d, want 0", status)
	}
	if !s.RemoveObserver(b) {
		t.Fatal("StopObserve from b should remove b's entry")
	}
	if s.ObserverCount() != 1 {
		t.Fatalf("ObserverCount = %d, want 1", s.ObserverCount())
	}
	if s.observers[0].Client != PeerConnection(a) {
		t.Error("b's StopObserve removed the wrong entry")
	}
	if !s.HasObserver(a) || s.HasObserver(b) {
		t.Error("after b stops, only a may still be an observer")
	}
	// A keeps working, B no longer does: that is the difference between the two.
	if !s.ActivateObserver(a, 8) {
		t.Error("a should still activate after b stopped")
	}
	if s.ActivateObserver(b, 8) {
		t.Error("b must not activate once its entry is gone (the activate returns silently)")
	}
	// And a second StopObserve from b removes nothing.
	if s.RemoveObserver(b) {
		t.Error("a repeated StopObserve should report nothing removed")
	}
	if s.ObserverCount() != 1 {
		t.Errorf("ObserverCount = %d after the repeated StopObserve, want 1", s.ObserverCount())
	}
}

// 4. ZERO IDENTIFIER: it is an identity, not its absence.
//
// The connection's identifier field starts at 0 (the constructor) and the 24-byte
// legacy format sets it to 0 explicitly, so a 0 is a real value. With
// one entry at 0, ANY connection of the session that has not sent a StartObserve
// Vx is also 0, and therefore the player's first 0x1A arms the observer's entry
// and the first 0x1B erases it.
//
// It is not a guess about what "should" happen: it is what the comparator does,
// and what is checked here is that ours does the same and not something more
// sensible.
func TestZeroIdentifierIsAnIdentity(t *testing.T) {
	s := identSession(2, 600)

	// The legacy observer: its request has no peerIdentifier, so the original
	// writes 0 into the connection.
	legacy := &mockConn{}
	if status, registered := s.StartObserve(legacy, 0, 0); status != 0 || !registered {
		t.Fatalf("StartObserve(legacy) = %d/%v, want 0/true", status, registered)
	}
	if legacy.PeerIdentifier() != 0 {
		t.Fatalf("the legacy observer's identifier = %#x, want 0", legacy.PeerIdentifier())
	}

	// A player of the same session never sent a StartObserve, so its identifier
	// field is
	// the constructor's 0. It is a different object, and the original still cannot
	// tell it apart from the legacy observer: that is the point of this test, and it
	// is why the answer here is "yes, it is reported as an observer".
	player := &mockConn{}
	if player.PeerIdentifier() != 0 {
		t.Fatalf("a fresh connection's identifier = %#x, want 0 (the constructor sets it)", player.PeerIdentifier())
	}
	if !s.HasObserver(player) {
		t.Fatal("a zero identifier matches a zero-identified entry: 0 == 0, and the " +
			"comparator has nothing else to go on")
	}

	// Its 0x1A does reach the zero-identified entry, and the 0x1E comes back to
	// the player, not to the observer.
	if !s.ActivateObserver(player, 8) {
		t.Fatal("a zero identifier must match a zero-identified entry")
	}
	if !s.observers[0].Active {
		t.Error("the observer's entry should be the one armed")
	}
	if countMsgs(player.msgs, 0x1E) != 1 {
		t.Errorf("the player got %d 0x1E, want 1: the answer goes to the caller", countMsgs(player.msgs, 0x1E))
	}
	if countMsgs(legacy.msgs, 0x1E) != 0 {
		t.Errorf("the observer got %d 0x1E, want 0: it did not ask", countMsgs(legacy.msgs, 0x1E))
	}
	// Armed or not, it is the entry that gets served, and the frames go to the
	// observer's own connection.
	s.Tick()
	if got := countMsgs(legacy.msgs, ackFrameType); got != 200 {
		t.Errorf("the observer's connection got %d frames, want 200", got)
	}

	// And the player's 0x1B removes that entry. It does not remove "its own",
	// because it has none: the comparator cannot tell them apart.
	if !s.RemoveObserver(player) {
		t.Fatal("StopObserve from a zero-identified connection should remove the zero entry")
	}
	if s.ObserverCount() != 0 {
		t.Errorf("ObserverCount = %d, want 0", s.ObserverCount())
	}
}

// 5. A SECOND StartObserve ON THE SAME CONNECTION: the entry is renamed.
//
// The original scans the list nowhere before appending, so
// a second 0x1C arriving at the 0 arm adds a second entry AND overwrites the
// connection's identifier. Because both entries point at the same connection and
// the identity is read LIVE through the pointer, both come to hold
// the new identifier. This is exactly what the "a repeated StartObserve is
// idempotent" guard this code had got wrong: it swallowed the new identifier and
// never registered it.
func TestStartObserveOnTheSameConnectionTwice(t *testing.T) {
	s := identSession(4, 600)

	c := &mockConn{}
	if status, _ := s.StartObserve(c, identA, 0); status != 0 {
		t.Fatalf("first request = %d, want 0", status)
	}
	if status, registered := s.StartObserve(c, identB, 0); status != 0 || !registered {
		t.Fatalf("second request = %d/%v, want 0/true: the original appends unconditionally", status, registered)
	}
	if s.ObserverCount() != 2 {
		t.Fatalf("ObserverCount = %d, want 2", s.ObserverCount())
	}
	if c.PeerIdentifier() != identB {
		t.Fatalf("the connection's identifier = %#x, want the second request's %#x", c.PeerIdentifier(), identB)
	}

	// Both entries now answer to identB, because the identity is read live.
	probe := &mockConn{peerID: identA}
	if s.ActivateObserver(probe, 8) {
		t.Error("the old identifier must no longer match: the entry follows the connection")
	}
	if !s.ActivateObserver(c, 8) {
		t.Error("the new identifier must match both entries")
	}
	// One activation walks the list once and stops at the first match, so exactly
	// one of the two entries is armed even though both now answer to identB.
	if !s.observers[0].Active {
		t.Error("the first entry should be the armed one (the activate stops at the match)")
	}
	if s.observers[1].Active {
		t.Error("the second entry must stay inactive: the walk returned at the first match")
	}
}

// 6. THE REPLACED CONNECTION: the entry does not migrate, and neither does the
// original's.
//
// The original's reconnection path looks the peer up by secret,
// and an observer is NOT in the session's peer table: peers
// are created by REQUESTJOIN + ADDPEER (the only path that writes a peer slot)
// and an observer sends neither. That is, the
// reconnection by secret cannot reach an observer, and what happens when its
// socket goes is the usual thing: the entry stays (nothing removes
// it, see RemoveObserver) and the game opens a new connection and sends a new
// 0x1C.
//
// What is checked here is the consequence that DOES depend on the key: a NEW
// connection presenting the OLD identifier resolves the OLD entry, because the
// comparator is the u64 and not the object. With the cap at 1, the entry keeps
// occupying it and the new 0x1C receives the cap's 2, not a 0.
func TestAReplacedConnectionDoesNotMigrateTheEntry(t *testing.T) {
	s := identSession(1, 600)

	old := &mockConn{}
	if status, _ := s.StartObserve(old, identA, 0); status != 0 {
		t.Fatalf("first request = %d, want 0", status)
	}
	if !s.ActivateObserver(old, 8) {
		t.Fatal("activation should find the entry")
	}
	s.Tick()
	progressBefore := s.observers[0].Progress
	if progressBefore != 200 {
		t.Fatalf("progress = %d after the first visit, want 200", progressBefore)
	}

	// The socket is gone. The new connection object's identifier is the
	// constructor's 0 and only its own 0x1C will give it an identity.
	fresh := &mockConn{}
	if fresh.PeerIdentifier() != 0 {
		t.Fatalf("a fresh connection's identifier = %#x, want 0", fresh.PeerIdentifier())
	}
	// It presents the identifier the OLD connection had, and that resolves the old
	// entry: same u64, first match.
	if !s.ActivateObserver(&mockConn{peerID: identA}, 8) {
		t.Error("a connection presenting the old identifier must resolve the old entry")
	}
	// But nothing was migrated, so the cap still holds the old entry and the new
	// 0x1C is rejected with 2, the observer-list-full answer.
	if status, registered := s.StartObserve(fresh, identA, 0); status != 2 || registered {
		t.Fatalf("the new request = %d/%v, want 2/false: the entry is not migrated", status, registered)
	}
	if s.ObserverCount() != 1 {
		t.Errorf("ObserverCount = %d, want 1", s.ObserverCount())
	}
	// The progress is the old entry's, not a fresh window: the session pointer in
	// the entry never changed.
	if s.observers[0].Progress != progressBefore {
		t.Errorf("progress = %d, want the old entry's %d", s.observers[0].Progress, progressBefore)
	}
}
