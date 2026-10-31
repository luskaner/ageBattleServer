package sim

import (
	"net"
	"testing"
	"time"
)

// TestReplayDialsWhenTheClientFirstTalks checks that a connection is opened at
// its first message's offset, not at the replay's start. The proxy logs a
// client's upgrade in the same second as its first message, so a real client
// connects when it is about to talk. A connection dialled at the start instead
// sits silent from its accept, and the server's silent timeout closes a late
// observer before it says a word.
func TestReplayDialsWhenTheClientFirstTalks(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	accepted := make(chan time.Time, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		accepted <- time.Now()
		buf := make([]byte, 64)
		for {
			if _, err := c.Read(buf); err != nil {
				_ = c.Close()
				return
			}
		}
	}()

	const firstMessageAt = 400 * time.Millisecond
	port := uint(ln.Addr().(*net.TCPAddr).Port)
	capConn := &Conn{ClientPort: 1, ServerPort: int(port), Transport: FramingRLink}
	sched := []scheduled{{at: firstMessageAt, msg: &Msg{Type: 0x14, Body: []byte{0x01}, Framing: FramingRLink}}}
	start := time.Now()
	if _, err := replayConn(capConn, port, sched, start, firstMessageAt+100*time.Millisecond, false); err != nil {
		t.Fatalf("replayConn: %v", err)
	}

	select {
	case at := <-accepted:
		if lead := at.Sub(start); lead < firstMessageAt-100*time.Millisecond {
			t.Fatalf("dialled %v after the start, want about %v (the first message's offset)", lead, firstMessageAt)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the replay never connected")
	}
}
