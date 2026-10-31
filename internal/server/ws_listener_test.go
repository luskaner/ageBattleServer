package server

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/lxzan/gws"

	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// wsListenerTest starts a standalone WsSocketListener (no server mainLoop).
//
// The listener is given port 0, so it binds ":0" and the port it got is read back
// from it. It used to be given a port this helper chose by binding ":0" and
// letting go, which leaves the number unowned until Start binds it and loses that
// race on Windows.
func wsListenerTest(t *testing.T) (*WsSocketListener, *Server, int) {
	t.Helper()

	srv := createTestServer(t)
	lnr, err := NewWsSocketListener(srv, srv.metricsCollector, srv.observerSocketSkipTicks(), 0, "", "")
	if err != nil {
		t.Fatalf("NewWsSocketListener error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	if err := lnr.Start(ctx); err != nil {
		t.Fatalf("listener Start error: %v", err)
	}
	t.Cleanup(func() { lnr.Stop() })
	port := lnr.GetPort()
	if port == 0 {
		t.Fatal("webSocketPort 0 after Start: the listener is not listening on a port")
	}
	return lnr, srv, port
}

func waitConns(t *testing.T, lnr *WsSocketListener, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for lnr.GetConnectionCount() != want && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if lnr.GetConnectionCount() != want {
		t.Fatalf("connection count = %d, want %d", lnr.GetConnectionCount(), want)
	}
}

func TestWsSocketListenerConnections(t *testing.T) {
	lnr, _, port := wsListenerTest(t)

	c1, _ := dialWS(t, port)
	defer func() { _ = c1.WriteClose(1000, nil) }()
	c2, _ := dialWS(t, port)
	defer func() { _ = c2.WriteClose(1000, nil) }()

	waitConns(t, lnr, 2)
	if len(lnr.GetConnections()) != 2 {
		t.Errorf("GetConnections = %d, want 2", len(lnr.GetConnections()))
	}
}

func TestWsSocketListenerBroadcast(t *testing.T) {
	lnr, _, port := wsListenerTest(t)

	c1, h1 := dialWS(t, port)
	defer func() { _ = c1.WriteClose(1000, nil) }()
	c2, h2 := dialWS(t, port)
	defer func() { _ = c2.WriteClose(1000, nil) }()

	waitConns(t, lnr, 2)

	lnr.Broadcast(gws.OpcodeBinary, []byte{0xAA, 0xBB})
	waitForFrame(t, h1, 0xAA, 2*time.Second)
	waitForFrame(t, h2, 0xAA, 2*time.Second)
}

func TestWsSocketListenerCloseConnection(t *testing.T) {
	lnr, _, port := wsListenerTest(t)

	c1, _ := dialWS(t, port)
	defer func() { _ = c1.WriteClose(1000, nil) }()
	c2, h2 := dialWS(t, port)
	defer func() { _ = c2.WriteClose(1000, nil) }()

	waitConns(t, lnr, 2)

	conns := lnr.GetConnections()
	lnr.CloseConnection(conns[0])
	if lnr.GetConnectionCount() != 1 {
		t.Errorf("connection count after CloseConnection = %d, want 1", lnr.GetConnectionCount())
	}

	lnr.CloseConnection(&gws.Conn{}) // unknown connection is a no-op
	if lnr.GetConnectionCount() != 1 {
		t.Errorf("connection count after unknown CloseConnection = %d, want 1", lnr.GetConnectionCount())
	}

	// The surviving connection still receives broadcasts.
	lnr.Broadcast(gws.OpcodeBinary, []byte{0xBB})
	waitForFrame(t, h2, 0xBB, 2*time.Second)
}

func TestWsSocketListenerGetLocalAddr(t *testing.T) {
	lnr, _, port := wsListenerTest(t)
	if got := lnr.GetLocalAddr(); got != ":"+strconv.Itoa(port) {
		t.Errorf("GetLocalAddr = %q, want %q", got, ":"+strconv.Itoa(port))
	}
	if lnr.IsSecure() {
		t.Error("listener should be plain")
	}
	if lnr.UnderlyingConn() != nil {
		t.Error("UnderlyingConn should be nil")
	}
}

func TestWsSocketListenerOnPingPong(t *testing.T) {
	lnr, _, port := wsListenerTest(t)

	client, _ := dialWS(t, port)
	defer func() { _ = client.WriteClose(1000, nil) }()

	waitConns(t, lnr, 1)

	// Client ping -> server OnPing -> pong (gws handles pong internally).
	if err := client.WriteMessage(gws.OpcodePing, []byte("hi")); err != nil {
		t.Fatalf("client ping: %v", err)
	}

	// Server ping -> client auto-pong -> server OnPong.
	lnr.Broadcast(gws.OpcodePing, []byte("yo"))
	time.Sleep(200 * time.Millisecond)
}

func TestWsSocketListenerOnOpenClose(t *testing.T) {
	lnr, srv, port := wsListenerTest(t)

	client, _ := dialWS(t, port)
	defer func() { _ = client.WriteClose(1000, nil) }()

	waitConns(t, lnr, 1)

	// gws calls OnOpen from ReadLoop (a goroutine), so the WsClient may lag
	// behind the connection registration; wait for it.
	var found bool
	deadline := time.Now().Add(2 * time.Second)
	for !found && time.Now().Before(deadline) {
		conns := lnr.GetConnections()
		srv.mu.Lock()
		for _, c := range srv.clients {
			if wc, ok := c.(*WsClient); ok {
				if len(conns) > 0 && wc.conn == conns[0] {
					found = true
				}
			}
		}
		srv.mu.Unlock()
		if !found {
			time.Sleep(10 * time.Millisecond)
		}
	}
	if !found {
		t.Error("server-side WsClient not registered for the socket")
	}

	// Client disconnect -> OnClose -> WsClient removed and closed.
	_ = client.WriteClose(1000, nil)
	gone := time.Now().Add(2 * time.Second)
	for {
		srv.mu.Lock()
		n := len(srv.clients)
		srv.mu.Unlock()
		if n == 0 || time.Now().After(gone) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	srv.mu.Lock()
	n := len(srv.clients)
	srv.mu.Unlock()
	if n != 0 {
		t.Errorf("clients after close = %d, want 0", n)
	}
}

func TestWsSocketListenerMessageFlow(t *testing.T) {
	lnr, _, port := wsListenerTest(t)

	client, _ := dialWS(t, port)
	defer func() { _ = client.WriteClose(1000, nil) }()

	waitConns(t, lnr, 1)

	// A ping message is enqueued without a server mainLoop draining it;
	// nothing crashes and the queue holds it.
	if err := client.WriteMessage(gws.OpcodeBinary, wsMessage(protocol.MsgPing, nil)); err != nil {
		t.Fatalf("client message: %v", err)
	}

	// Stop with an open connection sends close frames and shuts down cleanly.
	lnr.Stop()
	time.Sleep(100 * time.Millisecond)
	if err := lnr.Start(context.Background()); err != nil {
		t.Fatalf("restart after Stop error: %v", err)
	}
	lnr.Stop()
}

// TestWsSocketListenerInvalidUpgrade tests the handler path for invalid WS upgrade.
func TestWsSocketListenerInvalidUpgrade(t *testing.T) {
	lnr, _, port := wsListenerTest(t)

	// Send an HTTP request that looks like WS upgrade request but fails validation.
	conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Malformed HTTP request - no proper headers.
	_, _ = conn.Write([]byte("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	time.Sleep(100 * time.Millisecond)

	// The connection should not be registered since upgrade failed.
	waitConns(t, lnr, 0)
}

// TestWsSocketListenerEmptyMessage tests handling of empty messages.
func TestWsSocketListenerEmptyMessage(t *testing.T) {
	lnr, _, port := wsListenerTest(t)

	client, _ := dialWS(t, port)
	defer func() { _ = client.WriteClose(1000, nil) }()

	waitConns(t, lnr, 1)

	// Send an empty message - the handler should return early (lines 67-69).
	if err := client.WriteMessage(gws.OpcodeBinary, []byte{}); err != nil {
		t.Fatalf("client empty message: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}
