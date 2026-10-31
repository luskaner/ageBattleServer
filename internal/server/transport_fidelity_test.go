package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/lxzan/gws"

	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// This file pins the transport contract of the two game ports: how a message is
// delimited on each, that a stalled or slow peer never desynchronises or
// silently loses part of the stream, that a server-side close still delivers
// what was queued before it, and when a silent connection is closed.

// wsMessage is a game message as the game sends it on webSocketPort: the bare
// [type][body]. The WebSocket frame already delimits the message, so nothing
// precedes the type. The [u16 BE length] prefix belongs to bsPort (wireFrame).
func wsMessage(msgType byte, body []byte) []byte {
	frame := make([]byte, 0, 1+len(body))
	frame = append(frame, msgType)
	return append(frame, body...)
}

// tcpPair returns both ends of one loopback TCP connection: peer is the remote
// side a game client would hold, server is the side a Client wraps.
func tcpPair(t *testing.T) (peer, server *net.TCPConn) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()
	accepted := make(chan net.Conn, 1)
	go func() {
		c, aerr := ln.Accept()
		if aerr != nil {
			accepted <- nil
			return
		}
		accepted <- c
	}()
	dialed, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	srvSide := <-accepted
	if srvSide == nil {
		t.Fatal("accept failed")
	}
	t.Cleanup(func() {
		_ = dialed.Close()
		_ = srvSide.Close()
	})
	return dialed.(*net.TCPConn), srvSide.(*net.TCPConn)
}

// dispatched is one message as the transport handed it to the dispatcher.
type dispatched struct {
	msgType byte
	body    []byte
}

// recordingHost is a ClientHost that records what a transport dispatches, so a
// test observes the transport's parse without any handler behind it. A client
// that signals work is processed at once.
type recordingHost struct {
	mu      sync.Mutex
	got     []dispatched
	removed int
	clients []ClientConn
}

func (h *recordingHost) AddClient(c ClientConn) {
	h.mu.Lock()
	h.clients = append(h.clients, c)
	h.mu.Unlock()
}

func (h *recordingHost) RemoveClient(ClientConn) {
	h.mu.Lock()
	h.removed++
	h.mu.Unlock()
}

func (h *recordingHost) NotifyWork(c ClientConn) { c.Process(1024, time.Now()) }

func (h *recordingHost) dispatchMessage(_ messageClient, msgType byte, body []byte) {
	h.mu.Lock()
	h.got = append(h.got, dispatched{msgType: msgType, body: bytes.Clone(body)})
	h.mu.Unlock()
}

func (h *recordingHost) removedCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.removed
}

// waitDispatched returns the dispatched messages once there are at least n.
func (h *recordingHost) waitDispatched(t *testing.T, n int) []dispatched {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		h.mu.Lock()
		got := append([]dispatched(nil), h.got...)
		h.mu.Unlock()
		if len(got) >= n {
			return got
		}
		if time.Now().After(deadline) {
			t.Fatalf("dispatched %d message(s) within 2s, want %d: %v", len(got), n, got)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

type nopCounter struct{}

func (nopCounter) AddToCounter(string, float64) {}

// OverrideSilentTimeout completes ClientConn for the handler tests' fake, which is
// never processed against a clock.
func (m *mockMsgClient) OverrideSilentTimeout(time.Duration) {}

// --- webSocketPort: one binary message is [type][payload], nothing else ---

// TestWsMessageIsTypeAndPayload sends messages over a real WebSocket connection
// and checks what reaches the dispatcher. Each binary message is the type byte
// followed by the payload: no length prefix is stripped, whatever the first two
// bytes happen to be. The 4101-byte 0x10 starts 10 03, and 0x1003 is its own
// length minus two, so a parser that guesses at a prefix sees type 0x07. The
// two-byte 00 00 is type 0x00 with a one-byte payload. The empty message is
// ignored.
func TestWsMessageIsTypeAndPayload(t *testing.T) {
	host := &recordingHost{}
	lnr, err := NewWsSocketListener(host, nopCounter{}, 0, 0, "", "")
	if err != nil {
		t.Fatalf("NewWsSocketListener: %v", err)
	}
	if err := lnr.Start(t.Context()); err != nil {
		t.Fatalf("listener Start: %v", err)
	}
	t.Cleanup(lnr.Stop)
	client, _ := dialWS(t, lnr.GetPort())
	t.Cleanup(func() { _ = client.WriteClose(1000, nil) })

	long := make([]byte, 4101)
	for i := range long {
		long[i] = byte(i)
	}
	long[0], long[1], long[2] = protocol.MsgProxy, 0x03, 0x07

	for _, m := range [][]byte{{}, long, {0x00, 0x00}, wsMessage(protocol.MsgLobbyPing, []byte{0x01})} {
		if err := client.WriteMessage(gws.OpcodeBinary, m); err != nil {
			t.Fatalf("write %d bytes: %v", len(m), err)
		}
	}

	got := host.waitDispatched(t, 3)
	want := []dispatched{
		{msgType: protocol.MsgProxy, body: long[1:]},
		{msgType: 0x00, body: []byte{0x00}},
		{msgType: protocol.MsgLobbyPing, body: []byte{0x01}},
	}
	if len(got) != len(want) {
		t.Fatalf("dispatched %d messages, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].msgType != want[i].msgType || !bytes.Equal(got[i].body, want[i].body) {
			t.Errorf("message %d dispatched as type 0x%02x with %d payload bytes, want type 0x%02x with %d",
				i, got[i].msgType, len(got[i].body), want[i].msgType, len(want[i].body))
		}
	}
}

// TestWsZeroTwoByteMessageLeavesTheServerUp sends the two bytes 00 00 as one
// binary message to a running server. The server must survive it: the same
// connection is still answered, a new WebSocket connection still gets its 101,
// and bsPort still echoes a 0x14.
func TestWsZeroTwoByteMessageLeavesTheServerUp(t *testing.T) {
	// Started without startServer's goroutine-leak check: that profile is
	// process-wide, and this test is about the parser, not about what earlier
	// tests left behind.
	srv, err := New(&Config{EphemeralListenPorts: EphemeralBoth, LogLevel: "none"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := srv.Start(t.Context()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = srv.Stop() })

	first, firstH := dialWS(t, srv.wsListener.GetPort())
	t.Cleanup(func() { _ = first.WriteClose(1000, nil) })
	if err := first.WriteMessage(gws.OpcodeBinary, []byte{0x00, 0x00}); err != nil {
		t.Fatalf("write 00 00: %v", err)
	}
	// Messages of one connection are handled in order, so this echo proves the
	// 00 00 before it was handled.
	if err := first.WriteMessage(gws.OpcodeBinary, wsMessage(protocol.MsgLobbyPing, []byte{0xAB})); err != nil {
		t.Fatalf("write ping: %v", err)
	}
	if body := waitForFrame(t, firstH, protocol.MsgLobbyPing, 2*time.Second); !bytes.Equal(body, []byte{0xAB}) {
		t.Fatalf("ping echo on the same connection = %x, want AB", body)
	}

	second, secondH := dialWS(t, srv.wsListener.GetPort())
	t.Cleanup(func() { _ = second.WriteClose(1000, nil) })
	if err := second.WriteMessage(gws.OpcodeBinary, wsMessage(protocol.MsgLobbyPing, []byte{0xCD})); err != nil {
		t.Fatalf("write ping on a new connection: %v", err)
	}
	if body := waitForFrame(t, secondH, protocol.MsgLobbyPing, 2*time.Second); !bytes.Equal(body, []byte{0xCD}) {
		t.Fatalf("ping echo on a new connection = %x, want CD", body)
	}

	bs := dialPort(t, srv.bsListener.GetPort())
	if _, err := bs.Write(wireFrame(protocol.MsgLobbyPing, []byte{0xEF})); err != nil {
		t.Fatalf("bsPort write: %v", err)
	}
	if typ, body := readFrameFrom(t, bs); typ != protocol.MsgLobbyPing || !bytes.Equal(body, []byte{0xEF}) {
		t.Fatalf("bsPort answered type 0x%02x body %x, want the 0x14 echo EF", typ, body)
	}
}

// --- bsPort: a stall in the middle of a frame keeps the stream in step ---

// TestBsMidFrameStallKeepsTheStreamInStep splits frames at arbitrary points
// and stalls between the parts for many read deadlines. The bytes read before a
// stall belong to the frame being assembled, so both pings are echoed whole and
// in order: the first is split after five bytes (inside the body), the second
// after one byte (inside the length).
func TestBsMidFrameStallKeepsTheStreamInStep(t *testing.T) {
	peer, conn := tcpPair(t)
	client := NewClient(conn, createTestServer(t))
	client.readPoll = 20 * time.Millisecond
	_ = client.Start(t.Context())
	t.Cleanup(func() { _ = client.Close() })

	first := wireFrame(protocol.MsgLobbyPing, []byte{1, 2, 3, 4, 5, 6, 7, 8})
	second := wireFrame(protocol.MsgLobbyPing, []byte{9, 10, 11, 12, 13, 14, 15, 16})
	stall := 10 * client.readPoll

	parts := [][]byte{first[:5], append(append([]byte(nil), first[5:]...), second[:1]...), second[1:]}
	for i, p := range parts {
		if i > 0 {
			time.Sleep(stall)
		}
		if _, err := peer.Write(p); err != nil {
			t.Fatalf("write part %d: %v", i, err)
		}
	}

	for i, want := range [][]byte{first[3:], second[3:]} {
		typ, body := readFrameFrom(t, peer)
		if typ != protocol.MsgLobbyPing || !bytes.Equal(body, want) {
			t.Fatalf("echo %d = type 0x%02x body %x, want 0x14 %x", i, typ, body, want)
		}
	}
}

// --- bsPort: a server-side close delivers what was queued before it ---

// TestBsServerCloseDeliversQueuedFrames queues one reply and closes the
// connection from the server side, the way an observer kick or a denied join
// does. The peer must read the complete frame before the end of the stream. It
// runs 50 times because the loss is a race between the writer and the close.
func TestBsServerCloseDeliversQueuedFrames(t *testing.T) {
	host := &recordingHost{}
	want := wireFrame(protocol.MsgStartObserveResp, []byte{4})
	for run := range 50 {
		peer, conn := tcpPair(t)
		client := NewClient(conn, host)
		_ = client.Start(t.Context())
		if err := client.SendMessage(protocol.MsgStartObserveResp, []byte{4}); err != nil {
			t.Fatalf("run %d: SendMessage: %v", run, err)
		}
		_ = client.Close()

		_ = peer.SetReadDeadline(time.Now().Add(2 * time.Second))
		got, err := io.ReadAll(peer)
		if err != nil {
			t.Fatalf("run %d: read until the end of the stream: %v (got %x)", run, err, got)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("run %d: the peer read %x before the end of the stream, want %x", run, got, want)
		}
	}
	if n := host.removedCount(); n != 50 {
		t.Errorf("RemoveClient ran %d times for 50 closes, want 50", n)
	}
}

// TestWsServerCloseDeliversSentFrames is the same check on webSocketPort: a
// reply sent before a server-side close reaches the peer before the close.
func TestWsServerCloseDeliversSentFrames(t *testing.T) {
	srv, _, port := wsListenerServer(t)
	for run := range 20 {
		c, tap, done := dialLeakWS(t, port)
		wc := serverWsClient(t, srv)
		if err := wc.SendMessage(protocol.MsgStartObserveResp, []byte{4}); err != nil {
			t.Fatalf("run %d: SendMessage: %v", run, err)
		}
		_ = wc.Close()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("run %d: the server did not close the connection", run)
		}
		_ = c.NetConn().Close()
		select {
		case f := <-tap.frames:
			if !bytes.Equal(f, []byte{protocol.MsgStartObserveResp, 4}) {
				t.Fatalf("run %d: the peer read %x before the close, want 19 04", run, f)
			}
		default:
			t.Fatalf("run %d: the peer read nothing before the close, want 19 04", run)
		}
	}
}

// --- bsPort: back-pressure never loses a frame on an open connection ---

// TestBsPausedReaderGetsEveryFrameInOrder fills small socket buffers while the
// peer does not read, then lets the peer drain. Every frame arrives whole and
// in order, and the connection stays open: a slow reader is not a reason to
// drop frames. The pause is longer than any per-send wait could be.
func TestBsPausedReaderGetsEveryFrameInOrder(t *testing.T) {
	peer, conn := tcpPair(t)
	_ = conn.SetWriteBuffer(4096)
	_ = peer.SetReadBuffer(4096)
	host := &recordingHost{}
	client := NewClient(conn, host)
	_ = client.Start(t.Context())
	t.Cleanup(func() { _ = client.Close() })

	const frames = 600
	const bodyLen = 1000
	sent := make(chan error, 1)
	go func() {
		body := make([]byte, bodyLen)
		for i := range frames {
			binary.BigEndian.PutUint32(body, uint32(i))
			if err := client.SendMessage(protocol.MsgEndOfGameData, body); err != nil {
				sent <- fmt.Errorf("frame %d: %w", i, err)
				return
			}
		}
		sent <- nil
	}()

	time.Sleep(1500 * time.Millisecond)
	// The pause is over: a wide receive window lets the drain run at loopback
	// speed instead of one small window per round trip.
	_ = peer.SetReadBuffer(1 << 20)

	for i := range frames {
		typ, body := readFrameFrom(t, peer)
		if typ != protocol.MsgEndOfGameData || len(body) != bodyLen {
			t.Fatalf("frame %d: type 0x%02x with %d body bytes, want 0x0a with %d", i, typ, len(body), bodyLen)
		}
		if seq := binary.BigEndian.Uint32(body); seq != uint32(i) {
			t.Fatalf("frame %d carries sequence %d: %d frame(s) lost", i, seq, int(seq)-i)
		}
	}
	if err := <-sent; err != nil {
		t.Fatalf("SendMessage refused a frame on an open connection: %v", err)
	}
	if n := host.removedCount(); n != 0 {
		t.Errorf("the connection was closed (%d removal) although the peer drained it", n)
	}
}

// failOnceConn is a net.Conn whose first Write fails with a timeout, which is
// what an expired write deadline returns, and whose later Writes succeed. Read
// blocks until Close.
type failOnceConn struct {
	mu       sync.Mutex
	attempts int
	written  []byte
	done     chan struct{}
	once     sync.Once
}

func newFailOnceConn() *failOnceConn { return &failOnceConn{done: make(chan struct{})} }

func (c *failOnceConn) Read([]byte) (int, error) {
	<-c.done
	return 0, net.ErrClosed
}

func (c *failOnceConn) Write(b []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.attempts++
	if c.attempts == 1 {
		return 0, os.ErrDeadlineExceeded
	}
	c.written = append(c.written, b...)
	return len(b), nil
}

func (c *failOnceConn) Close() error {
	c.once.Do(func() { close(c.done) })
	return nil
}

func (c *failOnceConn) isClosed() bool {
	select {
	case <-c.done:
		return true
	default:
		return false
	}
}

func (c *failOnceConn) state() (attempts int, written []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.attempts, append([]byte(nil), c.written...)
}

func (c *failOnceConn) LocalAddr() net.Addr              { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)} }
func (c *failOnceConn) RemoteAddr() net.Addr             { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)} }
func (c *failOnceConn) SetDeadline(time.Time) error      { return nil }
func (c *failOnceConn) SetReadDeadline(time.Time) error  { return nil }
func (c *failOnceConn) SetWriteDeadline(time.Time) error { return nil }

// TestBsFailedWriteNeverLeavesAnOpenBlackHole makes one socket write fail. A
// failed write leaves the stream at an unknown offset, so the frames after it
// cannot be delivered in order: the connection must close. What may never
// happen is the connection staying open while frames vanish.
func TestBsFailedWriteNeverLeavesAnOpenBlackHole(t *testing.T) {
	conn := newFailOnceConn()
	host := &recordingHost{}
	client := NewClient(conn, host)
	_ = client.Start(t.Context())
	t.Cleanup(func() { _ = client.Close() })

	_ = client.SendMessage(protocol.MsgEndOfGameData, []byte{1})
	deadline := time.Now().Add(2 * time.Second)
	for {
		if n, _ := conn.state(); n > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the first frame was never written")
		}
		time.Sleep(time.Millisecond)
	}
	_ = client.SendMessage(protocol.MsgEndOfGameData, []byte{2})

	everything := append(wireFrame(protocol.MsgEndOfGameData, []byte{1}), wireFrame(protocol.MsgEndOfGameData, []byte{2})...)
	for {
		_, written := conn.state()
		if conn.isClosed() {
			if n := host.removedCount(); n != 1 {
				t.Errorf("the connection closed but RemoveClient ran %d times, want 1", n)
			}
			return
		}
		if bytes.Equal(written, everything) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("after a failed write the connection is still open and the socket got %x, want it closed", written)
		}
		time.Sleep(time.Millisecond)
	}
}

// --- the silent timeout ---

// clientClosed reports whether the bsPort client has been closed.
func clientClosed(c *Client) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

// wsClientClosed reports whether the WebSocket client has been closed.
func wsClientClosed(c *WsClient) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

// silentTimeoutCase is one connection with or without a timeout override, its
// last complete message at base, and the expected state at each probe after
// base.
type silentTimeoutCase struct {
	name     string
	override time.Duration // 0: no override, the 60 s default applies
	probes   []silentProbe
}

type silentProbe struct {
	after  time.Duration
	closed bool
}

var silentTimeoutCases = []silentTimeoutCase{
	{name: "default 60s", probes: []silentProbe{{59 * time.Second, false}, {60 * time.Second, false}, {61 * time.Second, true}}},
	{name: "override 5s", override: 5 * time.Second, probes: []silentProbe{{5 * time.Second, false}, {6 * time.Second, true}}},
	{name: "override 90s", override: 90 * time.Second, probes: []silentProbe{{61 * time.Second, false}, {91 * time.Second, true}}},
}

// TestClientSilentTimeout checks the bsPort client against an injected clock:
// a connection closes once more than its silent timeout has passed since its
// last complete message. The timeout is 60 s unless an override replaces it.
func TestClientSilentTimeout(t *testing.T) {
	for _, tc := range silentTimeoutCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient(&mockNetConn{}, &recordingHost{})
			base := time.Now()
			client.mu.Lock()
			client.lastSeen = base
			client.mu.Unlock()
			if tc.override > 0 {
				client.OverrideSilentTimeout(tc.override)
			}
			for _, p := range tc.probes {
				client.Process(100, base.Add(p.after))
				if got := clientClosed(client); got != p.closed {
					t.Fatalf("closed at %v = %v, want %v", p.after, got, p.closed)
				}
			}
		})
	}
}

// TestWsClientSilentTimeout is the same check on the WebSocket client.
func TestWsClientSilentTimeout(t *testing.T) {
	for _, tc := range silentTimeoutCases {
		t.Run(tc.name, func(t *testing.T) {
			wc := wsListenerOnlyClient(t)
			base := time.Now()
			wc.mu.Lock()
			wc.lastSeen = base
			wc.mu.Unlock()
			if tc.override > 0 {
				wc.OverrideSilentTimeout(tc.override)
			}
			for _, p := range tc.probes {
				wc.Process(100, base.Add(p.after))
				if got := wsClientClosed(wc); got != p.closed {
					t.Fatalf("closed at %v = %v, want %v", p.after, got, p.closed)
				}
			}
		})
	}
}

// TestClientSilentTimeoutMessagePostpones accepts a bsPort connection 50 s in
// the past and arms 60 s; a message arriving now restarts the count, so 61 s
// after the accept the connection is still open, and 61 s after the message it
// is closed.
func TestClientSilentTimeoutMessagePostpones(t *testing.T) {
	peer, conn := tcpPair(t)
	client := NewClient(conn, createTestServer(t))
	accepted := time.Now().Add(-50 * time.Second)
	client.mu.Lock()
	client.lastSeen = accepted
	client.mu.Unlock()
	client.OverrideSilentTimeout(60 * time.Second)
	_ = client.Start(t.Context())
	t.Cleanup(func() { _ = client.Close() })

	if _, err := peer.Write(wireFrame(protocol.MsgLobbyPing, []byte{0x01})); err != nil {
		t.Fatalf("write: %v", err)
	}
	readFrameFrom(t, peer) // the echo: the message has been read

	client.Process(100, accepted.Add(61*time.Second))
	if clientClosed(client) {
		t.Fatal("closed 61s after the accept although a message arrived 50s after it")
	}
	client.Process(100, time.Now().Add(61*time.Second))
	if !clientClosed(client) {
		t.Fatal("still open 61s after the last message")
	}
}

// TestWsClientSilentTimeoutMessagePostpones is the same check on the WebSocket
// client.
func TestWsClientSilentTimeoutMessagePostpones(t *testing.T) {
	srv, _, port := wsListenerServer(t)
	client, handler := dialWS(t, port)
	t.Cleanup(func() { _ = client.WriteClose(1000, nil) })
	wc := serverWsClient(t, srv)
	accepted := time.Now().Add(-50 * time.Second)
	wc.mu.Lock()
	wc.lastSeen = accepted
	wc.mu.Unlock()
	wc.OverrideSilentTimeout(60 * time.Second)

	if err := client.WriteMessage(gws.OpcodeBinary, wsMessage(protocol.MsgLobbyPing, []byte{0x01})); err != nil {
		t.Fatalf("write: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		wc.mu.Lock()
		queued := len(wc.incoming)
		wc.mu.Unlock()
		if queued > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the message never reached the client's queue")
		}
		time.Sleep(time.Millisecond)
	}

	wc.Process(100, accepted.Add(61*time.Second))
	if wsClientClosed(wc) {
		t.Fatal("closed 61s after the accept although a message arrived 50s after it")
	}
	waitForFrame(t, handler, protocol.MsgLobbyPing, 2*time.Second)
	wc.Process(100, time.Now().Add(61*time.Second))
	if !wsClientClosed(wc) {
		t.Fatal("still open 61s after the last message")
	}
}

// lobbyConnectWithOverride is a LobbyConnect in reconnect state state whose
// timeoutOverride is ms.
func lobbyConnectWithOverride(state byte, ms uint32) []byte {
	body := []byte{
		0x66, 0x55, 0x44, 0x33, 0x22, 0x11, // sessionID
		0xBB, 0xAA, // titleID
		0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA, 0x99, 0x88, // secret
		0x00,       // reconnectState, filled in below
		0x05, 0x00, // clientMessageVersion
		0, 0, 0, 0, // timeoutOverride, filled in below
		0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, // profileID
	}
	body[16] = state
	binary.LittleEndian.PutUint32(body[19:23], ms)
	return body
}

// TestLobbyConnectTimeoutOverrideSetsTheSilentTimeout checks the wiring: with
// -allowTimeoutOverride, a LobbyConnect's timeoutOverride (ms) replaces the 60 s
// silent timeout of its connection; an override of 0, or the option off, keeps
// the default.
func TestLobbyConnectTimeoutOverrideSetsTheSilentTimeout(t *testing.T) {
	cases := []struct {
		name     string
		allow    bool
		override uint32
		probes   []silentProbe
	}{
		{name: "allowed 5000ms", allow: true, override: 5000, probes: []silentProbe{{5 * time.Second, false}, {6 * time.Second, true}}},
		{name: "allowed 0ms", allow: true, override: 0, probes: []silentProbe{{6 * time.Second, false}, {61 * time.Second, true}}},
		{name: "not allowed", allow: false, override: 5000, probes: []silentProbe{{6 * time.Second, false}, {61 * time.Second, true}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := createTestServer(t)
			srv.config.AllowTimeoutOverride = tc.allow
			client := NewClient(&mockNetConn{}, srv)
			client.handleMessage(protocol.MsgLobbyConnect, lobbyConnectWithOverride(0, tc.override))
			base := time.Now()
			client.mu.Lock()
			client.lastSeen = base
			client.mu.Unlock()
			for _, p := range tc.probes {
				client.Process(100, base.Add(p.after))
				if got := clientClosed(client); got != p.closed {
					t.Fatalf("closed at %v = %v, want %v", p.after, got, p.closed)
				}
			}
		})
	}
}

// TestReconnectKeepsThePeersTimeoutOverride checks that a reconnect gives the new
// connection the silent timeout of the peer it resumes. The original moves only
// the socket into the old peer, which keeps the override from its first
// LobbyConnect, so the reconnect's own timeoutOverride field is not read.
func TestReconnectKeepsThePeersTimeoutOverride(t *testing.T) {
	srv := createTestServer(t)
	srv.config.AllowTimeoutOverride = true
	first := NewClient(&mockNetConn{}, srv)
	first.handleMessage(protocol.MsgLobbyConnect, lobbyConnectWithOverride(0, 5000))

	again := NewClient(&mockNetConn{}, srv)
	again.handleMessage(protocol.MsgLobbyConnect, lobbyConnectWithOverride(1, 0))
	if clientClosed(again) {
		t.Fatal("the reconnect was refused; the test needs it resumed")
	}
	base := time.Now()
	again.mu.Lock()
	again.lastSeen = base
	again.mu.Unlock()
	again.Process(100, base.Add(6*time.Second))
	if !clientClosed(again) {
		t.Fatal("open 6s after the last message, want the resumed peer's 5000ms override")
	}
}

// TestSendQueueOverflowUnderASessionLockDoesNotWedge overflows a session peer's
// send queue from the 0x0A broadcast, which sends while it holds the session's
// read lock. The overflow closes the connection; its removal from the session
// takes that lock too, so it must not run on the sending goroutine. The peer still
// leaves the session once the broadcast has released the lock.
func TestSendQueueOverflowUnderASessionLockDoesNotWedge(t *testing.T) {
	srv := createTestServer(t)
	_, srvSide := tcpPair(t)
	c := NewClient(srvSide, srv)
	c.maxQueuedBytes = 8 // the first 0x0A crosses it; the client is not started, so nothing drains
	sess := testSessionWithPeer(t, srv, c)
	sess.QueueGameData(sess.PeerIDForConn(c), []byte{0x02, 0x00, 0xAA, 0xBB})

	done := make(chan struct{})
	go func() {
		srv.sendEndOfGameDataTicks()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("sendEndOfGameDataTicks did not return within 3 s: the main loop is wedged")
	}
	if !clientClosed(c) {
		t.Fatal("the overflowing connection is still open")
	}
	deadline := time.Now().Add(3 * time.Second)
	for sess.PeerCount() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("PeerCount() = %d 3 s after the overflow, want 0", sess.PeerCount())
		}
		time.Sleep(time.Millisecond)
	}
}
