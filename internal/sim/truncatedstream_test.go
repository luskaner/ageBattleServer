package sim

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// WHAT THIS FILE IS
//
// A replay whose server->client stream was not collected whole cannot be
// compared, and this is the half of that rule unfinishedReplay did not cover.
//
// unfinishedReplay (393d528) handles the half where the harness could not DRIVE
// the connection: the dial failed, the WebSocket upgrade never completed. Those
// leave the replay as an error, so the run says "the replay did not complete"
// instead of reporting the empty side as a divergence of the server.
//
// The other half is what happened on the way back, and it used to be three notes
// on a connection that was then compared as if it were complete:
//
//   - the reader was still working when the 2 s wait after the close ran out, so
//     bytes the server had already written may never have been read;
//   - the collected bytes could not be framed at all;
//   - the last frame arrived in halves, and scanBytes reported the remainder as
//     "trailing bytes".
//
// All three are the same fact, and it is the fact this harness exists to refuse:
// what the harness got is not what the server sent, and the two look identical in
// the connection table. Here they are an error, through the same unfinishedReplay
// that already carries the driving failures.
//
// The fixture is a socket and not the server: it is the HARNESS's collection that
// is under test, so a fake that answers with half a frame is the honest way to ask
// what the harness does with half a frame. Nothing here needs a capture.

// partialFrameServer answers one connection with n whole RLink frames and then a
// frame that stops in the middle of its body, and closes when the harness closes.
// It returns the port it is listening on.
func partialFrameServer(t *testing.T, whole int) uint {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		// Wait for the harness's first write, so the frames below land inside the
		// replay window instead of racing the connection.
		buf := make([]byte, 512)
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if _, err := conn.Read(buf); err != nil {
			return
		}
		for i := range whole {
			body := []byte{byte(i), 0xAA, 0xBB}
			frame := []byte{0x00, byte(3 + len(body)), protocol.MsgEndOfGameData}
			if _, err := conn.Write(append(frame, body...)); err != nil {
				return
			}
		}
		// A frame that promises four body bytes and delivers two.
		_, _ = conn.Write([]byte{0x00, 0x06, protocol.MsgEndOfGameData, 0x01, 0x02})
		// Then read until the harness closes, which is when it stops collecting.
		_, _ = conn.Read(buf)
	}()
	return uint(ln.Addr().(*net.TCPAddr).Port)
}

// driveFragment drives one replayConn against the fake and returns what came back.
func driveFragment(t *testing.T, port uint, end time.Duration) (*Conn, error) {
	t.Helper()
	up := &Msg{Type: protocol.MsgLobbyConnect}
	truth := &Conn{ClientPort: 51001, ServerPort: int(port), Transport: FramingRLink, Up: []*Msg{up}}
	return replayConn(truth, port, []scheduled{{at: 0, msg: up}}, time.Now(), end, false)
}

// TestAStreamThatStopsMidFrameIsNotCompared is the whole claim: the harness got
// half of the last message, so the connection has no verdict to give, and the run
// has to say so with the reason.
//
// Before this, replayConn returned that connection with nil error and a note, and
// the comparison went ahead over the four messages it did get: the report would
// have said "the original answered N and we answered N-1" about a server that had
// answered all of them.
func TestAStreamThatStopsMidFrameIsNotCompared(t *testing.T) {
	port := partialFrameServer(t, 2)
	got, err := driveFragment(t, port, 300*time.Millisecond)
	if err == nil {
		t.Fatalf("replayConn accepted a stream whose last frame stopped mid-body and returned %d message(s) and no "+
			"error, as if they were the whole answer. The fake sent two whole frames and then four bytes of a third, "+
			"and the harness knows it", len(got.Down))
	}
	for _, want := range []string{"did not collect the whole", "trailing bytes", "incomplete frame"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error does not mention %q, and it is the only thing a reader gets: %v", want, err)
		}
	}
	// And it must not read as a verdict about the server, the same way
	// unfinishedReplay's error does not.
	for _, unwanted := range []string{"must-match", "divergence", "original"} {
		if strings.Contains(err.Error(), unwanted) {
			t.Errorf("the error says %q, which is a conclusion about the server or the capture: %v", unwanted, err)
		}
	}
	t.Logf("refused, as it has to be: %v", err)
	for _, n := range got.Notes {
		t.Logf("note on the connection: %s", n)
	}
}

// TestAStreamThatEndsCleanlyIsCompared is the other half, and it is what keeps the
// rule above from being "always fail": a fake that answers with whole frames and
// nothing else has to come back with no error at all, which is the case every one
// of the committed captures is in.
func TestAStreamThatEndsCleanlyIsCompared(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		buf := make([]byte, 512)
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if _, err := conn.Read(buf); err != nil {
			return
		}
		for i := range 3 {
			// [u16 BE length = type + 2 body bytes][type][body]
			frame := []byte{0x00, 0x03, protocol.MsgEndOfGameData, byte(i), 0xAA}
			if _, err := conn.Write(frame); err != nil {
				return
			}
		}
		_, _ = conn.Read(buf)
	}()
	port := uint(ln.Addr().(*net.TCPAddr).Port)
	got, err := driveFragment(t, port, 300*time.Millisecond)
	if err != nil {
		t.Fatalf("a stream of whole frames was refused: %v. The rule is about what is MISSING, not about the end of "+
			"a connection", err)
	}
	if len(got.Down) != 3 {
		t.Errorf("collected %d message(s), want the 3 whole frames the fake sent", len(got.Down))
	}
	for _, n := range got.Notes {
		if strings.Contains(n, "trailing bytes") || strings.Contains(n, "did not finish") {
			t.Errorf("a clean stream carries the truncation note %q", n)
		}
	}
}