package server

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/lxzan/gws"

	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/session"
)

// wsTestServer starts a full server with a WebSocket listener and returns it
// plus the port, ready for dialWS. The listener binds ":0" and the port it got
// is read back, so the number is never one this test chose and released.
func wsTestServer(t *testing.T) (*Server, int) {
	t.Helper()
	cfg := &Config{
		EphemeralListenPorts: EphemeralWS,
		LogLevel:             "none",
	}
	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	if err := srv.Start(ctx); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	t.Cleanup(func() { _ = srv.Stop() })
	port := srv.wsListener.GetPort()
	if port == 0 {
		t.Fatal("webSocketPort 0 after Start: the listener is not listening on a port")
	}
	return srv, port
}

// serverWsClient returns the server-side WsClient for a dialed connection.
func serverWsClient(t *testing.T, srv *Server) *WsClient {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		srv.mu.Lock()
		for _, c := range srv.clients {
			if wc, ok := c.(*WsClient); ok {
				srv.mu.Unlock()
				return wc
			}
		}
		srv.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("no server-side WsClient registered")
	return nil
}

// TestParseWsMessage pins the webSocketPort message format: one binary message
// is [type][payload] and nothing else. The first two bytes are never read as a
// length prefix, even when their big-endian value equals the message length
// minus two.
func TestParseWsMessage(t *testing.T) {
	if _, _, ok := parseWsMessage(nil); ok {
		t.Error("nil message should be rejected")
	}
	if _, _, ok := parseWsMessage([]byte{}); ok {
		t.Error("empty message should be rejected")
	}

	// 00 03 equals the length minus two: still type 0x00, the rest is payload.
	typ, body, ok := parseWsMessage([]byte{0x00, 0x03, 0x09, 0xAA, 0xBB})
	if !ok || typ != 0x00 || !bytes.Equal(body, []byte{0x03, 0x09, 0xAA, 0xBB}) {
		t.Errorf("look-alike prefix parse: type=%x body=%x ok=%v, want type 00 body 0309aabb", typ, body, ok)
	}

	// 00 00 equals the length minus two as well, and has no third byte.
	typ, body, ok = parseWsMessage([]byte{0x00, 0x00})
	if !ok || typ != 0x00 || !bytes.Equal(body, []byte{0x00}) {
		t.Errorf("00 00 parse: type=%x body=%x ok=%v, want type 00 body 00", typ, body, ok)
	}

	typ, body, ok = parseWsMessage([]byte{0x09, 0xAA, 0xBB})
	if !ok || typ != 0x09 || len(body) != 2 || body[0] != 0xAA {
		t.Errorf("bare parse failed: type=%x body=%x ok=%v", typ, body, ok)
	}

	// Short bare frames (1-2 bytes, e.g. ActivateObserver 0x1a or a LobbyPing
	// with a one-byte body) are valid bare messages with no length prefix.
	typ, body, ok = parseWsMessage([]byte{0x1a})
	if !ok || typ != 0x1a || len(body) != 0 {
		t.Errorf("1-byte bare parse failed: type=%x body=%x ok=%v", typ, body, ok)
	}
	typ, body, ok = parseWsMessage([]byte{0x14, 0x01})
	if !ok || typ != 0x14 || len(body) != 1 || body[0] != 0x01 {
		t.Errorf("2-byte bare parse failed: type=%x body=%x ok=%v", typ, body, ok)
	}
}

func TestWsClientEnqueueProcess(t *testing.T) {
	srv, port := wsTestServer(t)
	client, handler := dialWS(t, port)
	defer func() { _ = client.WriteClose(1000, nil) }()
	wc := serverWsClient(t, srv)

	wc.enqueue(wsMessage(protocol.MsgLobbyPing, nil))
	wc.Process(100, time.Now())

	waitForFrame(t, handler, protocol.MsgLobbyPing, 2*time.Second)
}

// wsListenerOnlyClient starts a listener-only server (its mainLoop is not
// running, so client queues are not drained automatically) on a free port and
// returns the server-side WsClient for a dialed connection.
func wsListenerOnlyClient(t *testing.T) *WsClient {
	t.Helper()
	srv, _, port := wsListenerServer(t)
	client, _ := dialWS(t, port)
	t.Cleanup(func() { _ = client.WriteClose(1000, nil) })
	return serverWsClient(t, srv)
}

// wsListenerServer starts a listener-only server on an OS-assigned port and
// returns the server, the listener, and the port it bound.
func wsListenerServer(t *testing.T) (*Server, *WsSocketListener, int) {
	t.Helper()
	srv := createTestServer(t)
	lnr, err := NewWsSocketListener(srv, srv.metricsCollector, srv.observerSocketSkipTicks(), 0, "", "")
	if err != nil {
		t.Fatalf("NewWsSocketListener error: %v", err)
	}
	if err := lnr.Start(t.Context()); err != nil {
		t.Fatalf("listener Start error: %v", err)
	}
	t.Cleanup(func() { lnr.Stop() })
	port := lnr.GetPort()
	if port == 0 {
		t.Fatal("webSocketPort 0 after Start: the listener is not listening on a port")
	}
	return srv, lnr, port
}

// TestObserverSendIsWebSocketFramed pins the observer framing. An observer is
// not a raw stream: the original routes every outbound send through the
// connection's own socket manager, which frames for that connection's transport,
// and the per-peer observer sends go through it like every other send. This
// server used to bypass the WebSocket layer for observers and prepend a
// [u16 BE length]; the payload the observer reads must now be the bare
// [type][body] inside a normal WebSocket binary frame.
func TestObserverSendIsWebSocketFramed(t *testing.T) {
	srv, _, port := wsListenerServer(t)

	client, handler := dialWS(t, port)
	t.Cleanup(func() { _ = client.WriteClose(1000, nil) })

	wc := serverWsClient(t, srv)
	wc.SetObserver(true)

	if err := wc.SendMessage(protocol.MsgEndOfGameData, []byte{0xFE, 0x01}); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}

	body := waitForFrame(t, handler, protocol.MsgEndOfGameData, 2*time.Second)
	// Bare [type][body]: the first two bytes are the 0xFE chunk tag and the
	// count, not a big-endian length.
	if len(body) != 2 || body[0] != 0xFE || body[1] != 0x01 {
		t.Errorf("observer payload = %x, want the bare body FE 01 with no length prefix", body)
	}
}

// TestWsSocketListenerRemovesClosed covers the conn-set cleanup: a connection
// that disconnects by itself must be removed from the listener's conn set by
// OnClose, so GetConnectionCount does not over-count over the server's life.
func TestWsSocketListenerRemovesClosed(t *testing.T) {
	srv, lnr, port := wsListenerServer(t)

	client, _ := dialWS(t, port)
	serverWsClient(t, srv) // wait until the server has registered the connection

	if n := lnr.GetConnectionCount(); n != 1 {
		t.Fatalf("GetConnectionCount after dial = %d, want 1", n)
	}

	_ = client.WriteClose(1000, nil)
	_ = client.NetConn().Close()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if lnr.GetConnectionCount() == 0 {
			return // closed conn removed by OnClose
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("GetConnectionCount never dropped after the client closed")
}

func TestWsClientProcessMaxPackets(t *testing.T) {
	// Listener-only server: no mainLoop draining between Process calls.
	wc := wsListenerOnlyClient(t)

	for range 5 {
		wc.enqueue(wsMessage(protocol.MsgLobbyPing, nil))
	}
	wc.Process(2, time.Now())

	wc.mu.Lock()
	remaining := len(wc.incoming)
	wc.mu.Unlock()
	if remaining != 3 {
		t.Errorf("incoming after Process(2) = %d, want 3", remaining)
	}
}

func TestWsClientProcessMalformed(t *testing.T) {
	srv, port := wsTestServer(t)
	client, _ := dialWS(t, port)
	defer func() { _ = client.WriteClose(1000, nil) }()
	wc := serverWsClient(t, srv)

	wc.enqueue([]byte{0x01, 0x02})
	wc.Process(100, time.Now())
}

func TestWsClientProcessObserverSkip(t *testing.T) {
	// Listener-only server: no mainLoop, so only the test's own Process calls
	// advance observerSkip and drain the queue.
	wc := wsListenerOnlyClient(t)

	wc.SetObserver(true)
	wc.mu.Lock()
	wc.skipMS = 3
	wc.mu.Unlock()

	// Pending data bypasses the skip: the message is processed on the first
	// Process call even though observerSkip (0) is below skipMS.
	wc.enqueue(wsMessage(protocol.MsgLobbyPing, nil))
	wc.Process(100, time.Now())

	wc.mu.Lock()
	remaining := len(wc.incoming)
	wc.mu.Unlock()
	if remaining != 0 {
		t.Errorf("incoming with pending data after 1 call = %d, want 0", remaining)
	}

	// Idle: the skip advances the counter and returns early; after skipMS calls
	// the counter resets and Process completes normally.
	wc.Process(100, time.Now())
	wc.Process(100, time.Now())

	wc.mu.Lock()
	skip := wc.observerSkip
	wc.mu.Unlock()
	if skip != 2 {
		t.Errorf("observerSkip after 2 idle calls = %d, want 2", skip)
	}

	wc.Process(100, time.Now())

	wc.mu.Lock()
	skip = wc.observerSkip
	wc.mu.Unlock()
	if skip != 0 {
		t.Errorf("observerSkip after 3rd idle call = %d, want 0", skip)
	}
}

func TestWsClientEnqueueOverflow(t *testing.T) {
	// Listener-only server: no mainLoop, so Process never drains the queue.
	wc := wsListenerOnlyClient(t)

	for range maxIncomingFrames {
		wc.enqueue([]byte{0x00, 0x01, 0x09})
	}
	wc.enqueue([]byte{0x00, 0x01, 0x09})

	wc.mu.Lock()
	n := len(wc.incoming)
	wc.mu.Unlock()
	if n != maxIncomingFrames {
		t.Errorf("incoming = %d, want %d", n, maxIncomingFrames)
	}
}

func TestWsClientProcessTimeout(t *testing.T) {
	srv, port := wsTestServer(t)
	client, _ := dialWS(t, port)
	defer func() { _ = client.WriteClose(1000, nil) }()
	wc := serverWsClient(t, srv)

	assertIdleTimeout(t,
		wc.Process,
		func(tm time.Time) {
			wc.mu.Lock()
			wc.lastSeen = tm
			wc.mu.Unlock()
		},
		func() bool {
			wc.mu.Lock()
			defer wc.mu.Unlock()
			return wc.closed
		})
}

func TestWsClientSendAndSendMessage(t *testing.T) {
	srv, port := wsTestServer(t)
	client, handler := dialWS(t, port)
	defer func() { _ = client.WriteClose(1000, nil) }()
	wc := serverWsClient(t, srv)

	if err := wc.Send([]byte{0xAA, 0xBB}); err != nil {
		t.Fatalf("Send error: %v", err)
	}
	waitForFrame(t, handler, 0xAA, 2*time.Second)

	if err := wc.SendMessage(protocol.MsgLobbyPing, []byte{0x01}); err != nil {
		t.Fatalf("SendMessage error: %v", err)
	}
	waitForFrame(t, handler, protocol.MsgLobbyPing, 2*time.Second)
}

func TestWsClientCloseAndSendClosed(t *testing.T) {
	srv, port := wsTestServer(t)
	client, _ := dialWS(t, port)
	defer func() { _ = client.WriteClose(1000, nil) }()
	wc := serverWsClient(t, srv)

	if err := wc.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	if err := wc.Close(); err != nil {
		t.Fatalf("double Close error: %v", err)
	}
	if err := wc.Send([]byte{0x01}); err == nil {
		t.Error("Send after Close should error")
	}
}

func TestWsClientLobbyPingTick(t *testing.T) {
	srv, port := wsTestServer(t)
	client, handler := dialWS(t, port)
	defer func() { _ = client.WriteClose(1000, nil) }()
	wc := serverWsClient(t, srv)

	now := time.Now().Add(11 * time.Second)
	wc.lobbyPingTick(now, 100)
	wc.lobbyPingTick(now.Add(5*time.Second), 101)

	body := waitForFrame(t, handler, protocol.MsgLobbyPing, 2*time.Second)
	if len(body) != 8 {
		t.Fatalf("ping body = %d bytes, want 8", len(body))
	}
	if body[0] != 100 {
		t.Errorf("ping counter = %d, want 100", body[0])
	}
}

func TestWsClientEndOfGameDataTick(t *testing.T) {
	srv, port := wsTestServer(t)
	client, _ := dialWS(t, port)
	defer func() { _ = client.WriteClose(1000, nil) }()
	wc := serverWsClient(t, srv)

	// The 0x0A tick is session-wide: queue a data sub for the host's slot
	// (peer 1) and pop one body from the session's ack state.
	testSessionWithPeer(t, srv, wc)
	sess := srv.sessionManager.GetSession(session.SessionKey{SessionID: 1})
	if sess == nil {
		t.Fatal("no session after join")
	}
	sess.QueueGameData(1, []byte{0x02, 0x00, 0xEE, 0xFF})
	body, _ := sess.PopAckTick()
	if len(body) < 7 || body[0] != 0xFE || body[1] != 0x01 {
		t.Errorf("unexpected ack body: %x", body)
	}
	if body, _ := sess.PopAckTick(); body == nil {
		t.Error("heartbeat expected after flush")
	}
}

func TestWsClientFlags(t *testing.T) {
	srv, port := wsTestServer(t)
	client, _ := dialWS(t, port)
	defer func() { _ = client.WriteClose(1000, nil) }()
	wc := serverWsClient(t, srv)

	if wc.pingStarted() {
		t.Error("new client should not be ping-started")
	}
	wc.setPingStarted()
	if !wc.pingStarted() {
		t.Error("client should be ping-started after setPingStarted")
	}

	wc.setTimeoutOverride(0xCAFEBABE)
	wc.mu.Lock()
	v := wc.timeoutOverride
	wc.mu.Unlock()
	if v != 0xCAFEBABE {
		t.Errorf("timeoutOverride = %#x, want 0xCAFEBABE", v)
	}
}

func TestWsClientSetObserver(t *testing.T) {
	srv, port := wsTestServer(t)
	client, _ := dialWS(t, port)
	defer func() { _ = client.WriteClose(1000, nil) }()
	wc := serverWsClient(t, srv)

	if wc.IsObserver() {
		t.Error("new client should not be an observer")
	}
	wc.SetObserver(true)
	if !wc.IsObserver() {
		t.Error("client should be an observer after SetObserver(true)")
	}
	wc.SetObserver(false)
	if wc.IsObserver() {
		t.Error("client should not be an observer after SetObserver(false)")
	}
}

func TestWsClientOnPingOnPong(t *testing.T) {
	srv, port := wsTestServer(t)
	client, _ := dialWS(t, port)
	defer func() { _ = client.WriteClose(1000, nil) }()
	wc := serverWsClient(t, srv)

	// Client ping -> server OnPing -> pong.
	if err := client.WriteMessage(gws.OpcodePing, []byte("ping-me")); err != nil {
		t.Fatalf("client ping: %v", err)
	}

	// Server ping -> client auto-pong -> server OnPong.
	if err := wc.conn.WritePing([]byte("pong-me")); err != nil {
		t.Fatalf("server ping: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
}
