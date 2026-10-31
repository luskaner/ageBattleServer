package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/lxzan/gws"

	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// benchWSHandler receives WS frames on a channel so the benchmark waits on a
// blocking receive instead of polling.
type benchWSHandler struct {
	frames chan []byte
}

func (h *benchWSHandler) OnOpen(*gws.Conn)         {}
func (h *benchWSHandler) OnClose(*gws.Conn, error) {}
func (h *benchWSHandler) OnPing(*gws.Conn, []byte) {}
func (h *benchWSHandler) OnPong(*gws.Conn, []byte) {}
func (h *benchWSHandler) OnMessage(_ *gws.Conn, m *gws.Message) {
	h.frames <- append([]byte(nil), m.Bytes()...)
}

// startWSServerBench boots a real server with a WebSocket listener.
func startWSServerBench(b *testing.B) (*Server, int) {
	b.Helper()
	silenceLogger()
	cfg := &Config{
		EphemeralListenPorts: EphemeralWS,
		LogLevel:             "none",
	}
	srv := startServer(b, cfg)
	port := srv.wsListener.GetPort()
	if port == 0 {
		b.Fatal("webSocketPort 0 after Start: the listener is not listening on a port")
	}
	return srv, port
}

// dialWSBench opens a WS client connection to the listener.
func dialWSBench(b *testing.B, port int, h *benchWSHandler) *gws.Conn {
	b.Helper()
	addr := fmt.Sprintf("ws://127.0.0.1:%d/", port)
	deadline := time.Now().Add(2 * time.Second)
	for {
		conn, _, err := gws.NewClient(h, &gws.ClientOption{Addr: addr})
		if err == nil {
			go conn.ReadLoop()
			return conn
		}
		if time.Now().After(deadline) {
			b.Fatalf("WS dial: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// BenchmarkWSPingRoundTrip measures one full WS client->server->client ping
// round trip. The reply is only dispatched inside WsClient.Process, which is
// the mainLoop-driven path P1 replaces: with the 1ms ticker each round trip
// includes the wait for the next tick; with event-driven processing it should
// drop to the raw network/dispatch latency.
func BenchmarkWSPingRoundTrip(b *testing.B) {
	_, port := startWSServerBench(b)
	h := &benchWSHandler{frames: make(chan []byte, 64)}
	conn := dialWSBench(b, port, h)
	defer func() { _ = conn.WriteClose(1000, nil) }()

	frame := wsMessage(protocol.MsgPing, nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := conn.WriteMessage(gws.OpcodeBinary, frame); err != nil {
			b.Fatalf("write: %v", err)
		}
		select {
		case got := <-h.frames:
			if typ, _ := frameType(got); typ != protocol.MsgPing {
				b.Fatalf("unexpected reply type 0x%02x", typ)
			}
		case <-time.After(2 * time.Second):
			b.Fatal("timeout waiting for WS ping reply")
		}
	}
	b.StopTimer()
}

// BenchmarkWSPingBurst measures the cost of draining a burst of N pings: it
// sends the whole batch first, then waits for all replies. This shows how the
// 1ms tick batches replies (poll) versus how event-driven processing drains
// them as they arrive.
func BenchmarkWSPingBurst(b *testing.B) {
	_, port := startWSServerBench(b)
	h := &benchWSHandler{frames: make(chan []byte, 1024)}
	conn := dialWSBench(b, port, h)
	defer func() { _ = conn.WriteClose(1000, nil) }()

	frame := wsMessage(protocol.MsgPing, nil)
	const batch = 64

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for range batch {
			if err := conn.WriteMessage(gws.OpcodeBinary, frame); err != nil {
				b.Fatalf("write: %v", err)
			}
		}
		for range batch {
			select {
			case got := <-h.frames:
				if typ, _ := frameType(got); typ != protocol.MsgPing {
					b.Fatalf("unexpected reply type 0x%02x", typ)
				}
			case <-time.After(2 * time.Second):
				b.Fatal("timeout waiting for WS ping reply")
			}
		}
	}
	b.StopTimer()
}
