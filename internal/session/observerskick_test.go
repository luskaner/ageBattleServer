package session

import (
	"bytes"
	"testing"
	"time"
)

// WHAT THIS FILE IS
//
// Session.Tick() kicks an observer out of a session that is neither recording nor
// reconnectable, and the original closes that observer's connection while it
// does, logging that the observer session is unavailable.
//
// Closing it is not a leaf call, and that is the whole subject of this file.
// WsClient.Close() calls RemoveClient synchronously, and RemoveClient reaches
// Session.removeObserverConn through Manager.HandleDisconnect, and that takes
// s.mu -- the lock the tick is holding. Client.Close() on the bsPort transport
// does the same from its own close path. A sync.RWMutex is not
// reentrant, so the tick deadlocks against itself, and the goroutine that runs it
// is the server's main loop: one kicked observer
// would stop every session's pages, not just its own.
//
// Two tests, because the two properties fail differently:
//
//   - TestObserverKickClosesTheConnection calls Tick() on the test goroutine. If
//     the close ever moves back under the lock while the mock connection does not
//     re-enter the session, this fails immediately and says so. It cannot hang:
//     its Close() is a plain field write.
//   - TestObserverKickDoesNotReenterTheSessionLock is the one that reproduces the
//     cycle: its connection's Close() calls back into the session exactly the way
//     the transports do. Tick() runs on its own goroutine and the test requires it
//     to come back inside observerKickReturnBound, so the defect is a failure with
//     a number in it rather than a suite that hangs until the go test timeout.
//
// The bound is a constant because "wait forever" is how a test of this shape
// takes the whole run down with it.

// observerKickReturnBound is how long the lock test waits for Tick to return. It
// is a ceiling and not a cost: Tick on a session with one observer is
// microseconds of work, and in a healthy run the wait ends on the first poll.
// The number only decides how long a REGRESSION takes to report itself.
const observerKickReturnBound = 2 * time.Second

// observerKickPoll is the gap between two checks of the tick's completion.
const observerKickPoll = 2 * time.Millisecond

// kickConn is a PeerConnection that closes the way the two transports close:
// straight back into the session, through the chain
//
//	Close() -> Server.RemoveClient(c) -> Manager.HandleDisconnect -> Session.removeObserverConn(c)
//
// (WsClient.Close, Client.Close and Server.RemoveClient). RemoveClient also drops the connection from
// the server's registries, but that half touches no session lock, so this fake
// stops at the part the deadlock is made of. Everything it records is what the
// tick handed the connection, so the assertions have data on both sides: the
// kick has to have produced a message AND a close, not one or the other.
type kickConn struct {
	msgs    []mockMsg
	closes  int
	sess    *Session
	peerID  uint64
	msgVer  uint16
	started bool
}

func (c *kickConn) SendMessage(msgType byte, body []byte) error {
	c.msgs = append(c.msgs, mockMsg{typ: msgType, body: append([]byte(nil), body...)})
	return nil
}

func (c *kickConn) Close() error {
	c.closes++
	c.started = true
	if c.sess != nil {
		// The same call Server.RemoveClient makes for this connection.
		c.sess.RemoveObserver(c)
	}
	return nil
}

func (c *kickConn) RemoteAddr() string { return "kick-conn" }

func (c *kickConn) SetPeerIdentifier(v uint64) { c.peerID = v }

func (c *kickConn) PeerIdentifier() uint64 { return c.peerID }

func (c *kickConn) ClientMessageVersion() uint16 { return c.msgVer }

// kickableSession is a session in the state the kick needs: not recording (so the
// tick does not leave through the serving arm first) and not reconnectable (so it
// enters the observer loop at all), holding one registered and activated observer.
func kickableSession(t *testing.T, id int64, conn PeerConnection) *Session {
	t.Helper()
	s := NewSession(id)
	s.SetObserverConfig(1, 1)
	s.SetReconnectable(true)
	s.RecordFrame(ackFrameType, ackFrame(1))
	if status, registered := s.StartObserve(conn, 7, 0); status != 0 || !registered {
		t.Fatalf("StartObserve = %d, registered %v; want 0, true", status, registered)
	}
	if !s.ActivateObserver(conn, 8) {
		t.Fatal("ActivateObserver should have found the entry it just registered")
	}
	if n := s.ObserverCount(); n != 1 {
		t.Fatalf("observer count = %d, want 1: the kick below would have nothing to kick", n)
	}
	// The store exists, so Recording is true; the kick needs it false.
	s.Recording = false
	s.SetReconnectable(false)
	return s
}

// TestObserverKickClosesTheConnection pins the original's kick arm on its own,
// with a connection whose Close cannot re-enter the session so the tick cannot
// deadlock even if the fix regresses: the failure is an assertion, not a hang.
//
// mockConn is the right connection here precisely because it is the opposite of
// the real one. It records the close and returns.
func TestObserverKickClosesTheConnection(t *testing.T) {
	conn := &mockConn{}
	s := kickableSession(t, 1, conn)
	conn.msgs = nil // drop the 0x1E the activation answered with

	s.Tick()

	if n := countMsgs(conn.msgs, 0x19); n != 1 {
		t.Fatalf("the kick sent %d StartObserveResponses, want 1: %+v", n, conn.msgs)
	}
	if !bytes.Equal(conn.msgs[0].body, []byte{4}) {
		t.Errorf("0x19 body = %x, want the single byte 4", conn.msgs[0].body)
	}
	// The original schedules the close with a zero delay, so the
	// socket is due on the next pass of its loop. Closing it here is that close.
	if !conn.closed {
		t.Error("the kicked observer's connection is still open: the kick closes it, " +
			"and a client left on a session with no data is the state that arm exists to end")
	}
	if n := s.ObserverCount(); n != 0 {
		t.Errorf("observer count after the kick = %d, want 0", n)
	}
}

// TestObserverKickDoesNotReenterTheSessionLock is the falsifiable half: the
// connection closes back into the session, so a close taken while Tick holds
// s.mu never returns and this test reports that instead of hanging.
//
// A second observer is registered on purpose. The defect closes every observer,
// so with two of them the first close is enough to wedge the tick and the test
// fails; with one it would also fail, and two makes the claim "the whole tick
// stops, not just one observer" something the test actually exercises.
func TestObserverKickDoesNotReenterTheSessionLock(t *testing.T) {
	s := NewSession(2)
	// sess is handed to each connection up front: it is what makes their Close()
	// re-enter this session, and a connection built without it would close like a
	// mock and the test would pass over the defect it exists to catch.
	first := &kickConn{sess: s}
	second := &kickConn{sess: s}
	s.SetObserverConfig(2, 1)
	s.SetReconnectable(true)
	s.RecordFrame(ackFrameType, ackFrame(1))
	for i, c := range []*kickConn{first, second} {
		if status, registered := s.StartObserve(c, uint64(i+1), 0); status != 0 || !registered {
			t.Fatalf("StartObserve(%d) = %d, registered %v; want 0, true", i, status, registered)
		}
		if !s.ActivateObserver(c, 8) {
			t.Fatalf("ActivateObserver(%d) did not find the entry", i)
		}
		c.msgs = nil
	}
	s.Recording = false
	s.SetReconnectable(false)

	done := make(chan struct{})
	go func() {
		defer close(done)
		s.Tick()
	}()

	select {
	case <-done:
	case <-time.After(observerKickReturnBound):
		// Nothing below may touch the session: the tick is holding s.mu and will
		// not let go of it, so any read here would block the test instead of
		// failing it.
		t.Fatalf("Session.Tick did not return within %s with %d observers in the list. "+
			"The kick closes each observer's connection from inside the section that holds s.mu "+
			"(internal/session/session.go), and Close() re-enters the session through "+
			"RemoveObserver, so the tick is waiting on a lock it already owns. The goroutine "+
			"that runs it is the server's main loop, so this is a whole-server hang, not one "+
			"observer's send. The connections close after the lock is released, the way "+
			"closeExpiredPeers does it.",
			observerKickReturnBound, 2)
	}

	// Tick returned, so everything below is ordered after it and the session is
	// free. Both halves of the kick have to be there: the message went out and
	// the connection went down.
	for i, c := range []*kickConn{first, second} {
		if n := countMsgs(c.msgs, 0x19); n != 1 {
			t.Errorf("observer %d got %d StartObserveResponses, want 1: %+v", i, n, c.msgs)
		}
		if !bytes.Equal(c.msgs[0].body, []byte{4}) {
			t.Errorf("observer %d got 0x19 body %x, want the single byte 4", i, c.msgs[0].body)
		}
		if c.closes != 1 {
			t.Errorf("observer %d had Close called %d times, want exactly 1: the kick "+
				"schedules the close once per observer", i, c.closes)
		}
	}
	if n := s.ObserverCount(); n != 0 {
		t.Errorf("observer count after the kick = %d, want 0", n)
	}
	// The re-entrant RemoveObserver ran on a list this tick had already emptied,
	// so it must have found nothing to remove. If it ever removes something, the
	// list is being emptied somewhere other than where it should be emptied.
	if s.observers != nil && len(s.observers) != 0 {
		t.Errorf("observer list = %d entries, want 0", len(s.observers))
	}
}

// TestObserverKickTickReturnsWithoutObservers is the guard on the other side of
// the split: the tick that finds nothing to do must return an empty kick list and
// must not touch a connection. It exists so a change that made Tick close
// something outside the kick arm would fail here rather than only in the two
// tests above.
func TestObserverKickTickReturnsWithoutObservers(t *testing.T) {
	s := NewSession(3)
	s.SetObserverConfig(1, 1)
	if got := s.tickLocked(); len(got) != 0 {
		t.Errorf("tickLocked on a session with no observers returned %d connection(s), want 0", len(got))
	}
	conn := &mockConn{}
	if status, _ := s.StartObserve(conn, 9, 0); status != 0 {
		t.Fatalf("StartObserve = %d, want 0", status)
	}
	// Still reconnectable and recording, which is the serving arm and not the
	// kick: nothing may be closed.
	if got := s.tickLocked(); len(got) != 0 {
		t.Errorf("tickLocked on a serving session returned %d connection(s), want 0", len(got))
	}
	if conn.closed {
		t.Error("tickLocked closed a connection on a session that is still serving it")
	}
}
