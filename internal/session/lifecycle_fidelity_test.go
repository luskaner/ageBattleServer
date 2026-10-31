package session

import (
	"testing"
)

// WHAT THIS FILE IS
//
// The session side of a connection's lifecycle: a reconnect moves the relay route
// with the peer, a close takes the peer out and can end the match, and a session
// is erased only once it has no member and no observer left.

// The close ends the match the same way the last LeaveSimulation does: a
// recording session whose last active member is removed is over, so the tick
// stops building acks.
func TestRemovingTheLastActiveMemberEndsTheMatch(t *testing.T) {
	s, _ := recordingSession(t, 1, 1, 2)

	s.RemovePeer(2)
	if !s.Recording || s.GameStarted {
		t.Fatal("peer 1 is still in the game: the match must go on")
	}

	s.RemovePeer(1)
	if s.Recording {
		t.Error("Recording is still set after the last active member left")
	}
	if !s.GameStarted {
		t.Error("GameStarted is not set after the last active member left")
	}
	if body, _ := s.PopAckTick(); body != nil {
		t.Errorf("an ended match still builds acks: %x", body)
	}
}

// The test is "no ACTIVE member remains", not "no member remains": a peer that
// joined after the frame store was created is not in the started game and does
// not keep the match alive.
func TestALateJoinerDoesNotKeepTheMatchAlive(t *testing.T) {
	s, _ := recordingSession(t, 1, 1)
	s.AddPeer(NewPeer(2, &mockConn{}))

	s.RemovePeer(1)
	if !s.GameStarted {
		t.Error("GameStarted is not set: the remaining peer joined after the store and is not in the game")
	}
}

// A session that never recorded has no match to end.
func TestRemovingAPeerOfALobbyLeavesTheMatchFlagsAlone(t *testing.T) {
	s := joinedIn(t, 1, 1)
	s.RemovePeer(1)
	if s.Recording || s.GameStarted {
		t.Errorf("Recording=%v GameStarted=%v, want both false for a session that never recorded",
			s.Recording, s.GameStarted)
	}
}

// A reconnect moves the relay route with the peer: the new connection resolves to
// the peer's session and the replaced one to none.
func TestManagerReconnectPeerMovesTheRoute(t *testing.T) {
	m := NewManager(0, 0)
	s := m.GetOrCreateSession(SessionKey{SessionID: 5})
	old, fresh := &mockConn{}, &mockConn{}
	if err := m.AddPeer(s, NewPeer(1, old)); err != nil {
		t.Fatalf("AddPeer: %v", err)
	}

	replaced, ok := m.ReconnectPeer(s, 1, fresh, 0x77)
	if !ok || replaced != old {
		t.Fatalf("ReconnectPeer = (%v, %v), want the old connection and true", replaced, ok)
	}
	if got := m.FindSessionByPeer(fresh); got != s {
		t.Errorf("the new connection resolves to %p, want the peer's session %p", got, s)
	}
	if got := m.FindSessionByPeer(old); got != nil {
		t.Errorf("the replaced connection resolves to %p, want none", got)
	}
	if got := s.PeerIDForConn(fresh); got != 1 {
		t.Errorf("the new connection holds slot %d, want 1", got)
	}
}

// The close path: the peer leaves at once, the session stays while an observer
// watches it, and it is erased with the last observer.
func TestHandleDisconnectErasesTheSessionWithItsLastWatcher(t *testing.T) {
	m := NewManager(0, 0)
	key := SessionKey{SessionID: 6}
	s := m.GetOrCreateSession(key)
	s.SetObserverConfig(1, 8)
	player, watcher := &mockConn{}, &mockConn{}
	if err := m.AddPeer(s, NewPeer(1, player)); err != nil {
		t.Fatalf("AddPeer: %v", err)
	}
	if _, ok := s.StartObserve(watcher, 0x42, 0); !ok {
		t.Fatal("the observer should be registered")
	}

	m.HandleDisconnect(player, nil)
	if got := s.PeerCount(); got != 0 {
		t.Errorf("PeerCount() = %d after the player's close, want 0", got)
	}
	if got := m.FindSessionByPeer(player); got != nil {
		t.Errorf("the closed connection still resolves to %p", got)
	}
	if m.GetSession(key) != s {
		t.Fatal("the session was erased while an observer still watches it")
	}

	m.HandleDisconnect(watcher, s)
	if got := s.ObserverCount(); got != 0 {
		t.Errorf("ObserverCount() = %d after the observer's close, want 0", got)
	}
	if m.GetSession(key) != nil {
		t.Error("the session survived its last member and its last observer")
	}
}

// The close removes the closing connection's own observer entry. An entry is
// matched by its connection here, not by the identifier the StopObserve lookup
// compares, so of two observers sharing an identifier the one that closes is the
// one that goes.
func TestHandleDisconnectRemovesTheClosingObserversOwnEntry(t *testing.T) {
	m := NewManager(0, 0)
	s := m.GetOrCreateSession(SessionKey{SessionID: 7})
	s.SetObserverConfig(2, 8)
	if err := m.AddPeer(s, NewPeer(1, &mockConn{})); err != nil {
		t.Fatalf("AddPeer: %v", err)
	}
	first, second := &mockConn{}, &mockConn{}
	for _, c := range []*mockConn{first, second} {
		if _, ok := s.StartObserve(c, 0x42, 0); !ok {
			t.Fatal("both observers should be registered")
		}
	}

	m.HandleDisconnect(second, s)

	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.observers) != 1 || s.observers[0].Client != first {
		t.Errorf("the observer list after the second one closed holds %d entries, want only the first's", len(s.observers))
	}
}
