package session

import (
	"errors"
	"testing"
	"time"
)

// TestAddPeerRefusesAnErasedSession covers the window between a connect's
// GetOrCreateSession and its AddPeer: the session's last member closes in
// between and the session is erased. AddPeer must refuse the erased session, so
// the connect can fetch the live one, instead of routing the joiner into a
// session no later connect will find.
func TestAddPeerRefusesAnErasedSession(t *testing.T) {
	m := NewManager(0, 0)
	k := SessionKey{SessionID: 9}
	s := m.GetOrCreateSession(k)
	last := &mockConn{}
	if err := m.AddPeer(s, NewPeer(1, last)); err != nil {
		t.Fatal(err)
	}

	handed := m.GetOrCreateSession(k)
	m.HandleDisconnect(last, nil)
	joiner := &mockConn{}
	err := m.AddPeer(handed, NewPeer(handed.NextPeerID(), joiner))
	if !errors.Is(err, ErrSessionErased) {
		t.Fatalf("AddPeer on an erased session = %v, want ErrSessionErased", err)
	}
	if m.FindSessionByPeer(joiner) != nil {
		t.Fatal("the joiner was routed to the erased session")
	}
}

// reentrantConn is a connection whose Close reaches back into the manager, as a
// WebSocket connection's does (Close -> RemoveClient -> HandleDisconnect).
type reentrantConn struct {
	mockConn
	onClose func()
}

func (c *reentrantConn) Close() error {
	c.closed = true
	if c.onClose != nil {
		c.onClose()
	}
	return nil
}

// TestRemoveSessionClosesOutsideTheManagerLock checks that RemoveSession does not
// close the session's connections while it holds the manager lock: a close that
// calls HandleDisconnect would wait on that lock for good.
func TestRemoveSessionClosesOutsideTheManagerLock(t *testing.T) {
	m := NewManager(0, 0)
	s := m.CreateSession()
	conn := &reentrantConn{}
	conn.onClose = func() { m.HandleDisconnect(conn, nil) }
	if err := m.AddPeer(s, NewPeer(1, conn)); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		m.RemoveSession(s.Key())
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("RemoveSession did not return within 3 s: a close re-entered the manager lock")
	}
	if !conn.closed {
		t.Error("the peer's connection was not closed")
	}
	if m.GetSession(s.Key()) != nil {
		t.Error("the session is still in the map")
	}
}
