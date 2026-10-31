package server

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/luskaner/ageBattleServer/internal/logger"
	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/session"
)

// silenceLogger drops the default logger to LevelNone so the traffic tests and
// benchmarks measure protocol work, not log formatting.
func silenceLogger() {
	logger.SetGlobalLevel(logger.LevelNone)
}

// startBlackBoxServer boots a real server on a real bsPort listener, speaking
// only the wire protocol from the client side. A short simulation period makes
// the per-tick acks arrive quickly.
//
// The port is not chosen: the listener binds ":0" (Config.EphemeralListenPorts)
// and the port it got is read back from the server, so there is no window in
// which the number belongs to nobody. This helper used to bind ":0", read the
// number and close it, which loses that window on Windows, where a listener bind
// is exclusive: whatever takes the port in between keeps it and Start fails. Only
// the bsPort is asked for, so the server under test has the same one listener it
// had before.
func startBlackBoxServer(t testing.TB) (*Server, uint) {
	t.Helper()
	silenceLogger()
	if tt, ok := t.(*testing.T); ok {
		assertNoGoroutineLeaks(tt)
	}
	cfg := &Config{
		EphemeralListenPorts: EphemeralBS,
		SimulationPeriod:    10,
		LogLevel:            "none",
	}
	srv := startServer(t, cfg)
	port := uint(srv.bsListener.GetPort())
	if port == 0 {
		t.Fatalf("bsPort 0 after Start: the listener is not listening on a port")
	}
	return srv, port
}

// trafficClient is a black-box wire client: a real TCP connection speaking the
// real [u16 length][type][body] framing, with byte counters for the load tests.
type trafficClient struct {
	t    testing.TB
	conn net.Conn
	rd   *bufio.Reader
	sent int64
	recv int64
}

func newTrafficClient(t testing.TB, port uint) *trafficClient {
	t.Helper()
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("traffic dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return &trafficClient{t: t, conn: conn, rd: bufio.NewReader(conn)}
}

func (c *trafficClient) send(msgType byte, body []byte) {
	c.t.Helper()
	frame := wireFrame(msgType, body)
	if _, err := c.conn.Write(frame); err != nil {
		c.t.Fatalf("traffic write: %v", err)
	}
	c.sent += int64(len(frame))
}

// readFrame reads one framed message; fails the test on timeout or EOF.
func (c *trafficClient) readFrame(timeout time.Duration) (byte, []byte) {
	c.t.Helper()
	typ, body, ok := c.tryReadFrame(timeout)
	if !ok {
		c.t.Fatalf("timed out reading frame")
	}
	return typ, body
}

// tryReadFrame reads one framed message, returning ok=false on timeout.
func (c *trafficClient) tryReadFrame(timeout time.Duration) (byte, []byte, bool) {
	c.t.Helper()
	_ = c.conn.SetReadDeadline(time.Now().Add(timeout))
	var lenBuf [2]byte
	if _, err := io.ReadFull(c.rd, lenBuf[:]); err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return 0, nil, false
		}
		c.t.Fatalf("traffic read length: %v", err)
	}
	length := int(binary.BigEndian.Uint16(lenBuf[:]))
	if length < 1 {
		c.t.Fatalf("traffic frame length %d", length)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(c.rd, payload); err != nil {
		c.t.Fatalf("traffic read payload: %v", err)
	}
	c.recv += int64(2 + length)
	return payload[0], payload[1:], true
}

// expectFrame waits for a frame of the wanted type, skipping any session ack
// heartbeats that may interleave (once game data has started the server emits
// one 0x0A heartbeat per simulation tick), then verifies the body.
func (c *trafficClient) expectFrame(timeout time.Duration, want byte, body []byte) {
	c.t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		typ, got, ok := c.tryReadFrame(150 * time.Millisecond)
		if !ok {
			continue
		}
		if typ != want {
			continue
		}
		if body != nil && !bytes.Equal(got, body) {
			c.t.Fatalf("frame type=0x%02x body=%x, want body=%x", typ, got, body)
		}
		return
	}
	c.t.Fatalf("timed out waiting for frame type=0x%02x", want)
}

// lobbyConnectBody builds a valid LobbyConnect body (accepted by
// protocol.ParseLobbyConnect) for the given 48-bit session ID.
func lobbyConnectBody(sessionID uint64) []byte {
	body := make([]byte, 0, 31)
	var b8 [8]byte
	binary.LittleEndian.PutUint64(b8[:], sessionID)
	body = append(body, b8[:]...)
	binary.LittleEndian.PutUint64(b8[:], 0x1234_5678_9abc_def0)
	body = append(body, b8[:]...)
	body = append(body, 0) // reconnectState = 0
	binary.LittleEndian.PutUint16(b8[:2], 5)
	body = append(body, b8[:2]...) // clientMessageVersion = 5
	binary.LittleEndian.PutUint32(b8[:4], 0)
	body = append(body, b8[:4]...) // timeoutOverride
	binary.LittleEndian.PutUint64(b8[:], 0x0102_0304_0506_0708)
	body = append(body, b8[:]...) // profileID
	return body
}

// gameDataPage serializes a replay-data GameData page (peer ID 0xFE), the same
// stream a real host uploads and which generates the session-wide acks.
func gameDataPage(payload []byte) []byte {
	return (&protocol.GameDataMessage{InnerType: 0xFE, Payload: payload}).Serialize()
}

// TestBlackBoxSessionTraffic simulates a real two-player game session entirely
// over the wire, as an external client: LobbyConnect handshake, keepalive
// echoes, GameData pages with session-wide acks, and the LobbyRelay peer-init
// and relay-routing paths. No internals are touched.
func TestBlackBoxSessionTraffic(t *testing.T) {
	_, port := startBlackBoxServer(t)

	host := newTrafficClient(t, port)
	join := newTrafficClient(t, port)

	// Both players join the same session; the server echoes the connect body.
	connect := lobbyConnectBody(1)
	host.send(protocol.MsgLobbyConnect, connect)
	host.expectFrame(2*time.Second, protocol.MsgLobbyConnect, connect)
	join.send(protocol.MsgLobbyConnect, connect)
	join.expectFrame(2*time.Second, protocol.MsgLobbyConnect, connect)

	// Keepalive echoes.
	host.send(protocol.MsgLobbyPing, nil)
	host.expectFrame(2*time.Second, protocol.MsgLobbyPing, nil)
	host.send(protocol.MsgLobbyPing, []byte{0x01})
	host.expectFrame(2*time.Second, protocol.MsgLobbyPing, []byte{0x01})

	// The host streams two replay pages; every player in the session receives
	// the per-tick ack flush (0xFE <count> ...) covering both sub-messages.
	page := gameDataPage([]byte{0xAA, 0xBB, 0xCC})
	host.send(protocol.MsgGameData, page)
	host.send(protocol.MsgGameData, page)
	for _, c := range []*trafficClient{host, join} {
		deadline := time.Now().Add(2 * time.Second)
		acked := 0
		for acked < 2 {
			if time.Now().After(deadline) {
				t.Fatalf("client got %d/2 game-data acks", acked)
			}
			typ, body := c.readFrame(250 * time.Millisecond)
			if typ != protocol.MsgEndOfGameData || len(body) < 2 || body[0] != 0xfe {
				continue
			}
			acked += int(body[1])
		}
	}

	// The host's first relay carries its compressed RequestJoin for the session:
	// the server consumes it and delivers the peer message it carried as a
	// MsgPeer to every session member (host peer ID is 1, the joiner is 2).
	initRelay := relayBody(0x01, 0x01, compressedPeerMessage(t, requestJoinInner(session.SessionKey{SessionID: 1}, 0x10)))
	host.send(protocol.MsgLobbyRelay, initRelay)
	host.expectFrame(2*time.Second, protocol.MsgPeer, initRelay[3:])
	join.expectFrame(2*time.Second, protocol.MsgPeer, initRelay[3:])

	// Later relays are routed to the addressed peer (body[0] = target peer ID),
	// not echoed to the sender.
	relay := make([]byte, 64)
	relay[0], relay[1], relay[2] = 0x02, 0x01, 0x0f // target = joiner (peer 2)
	for i := 3; i < len(relay); i++ {
		relay[i] = byte(0x20 + i%0x70)
	}
	host.send(protocol.MsgLobbyRelay, relay)
	join.expectFrame(2*time.Second, protocol.MsgLobbyRelay, relay)

	// The host must not receive that routed relay (only session acks flow to it).
	deadline := time.Now().Add(150 * time.Millisecond)
	for time.Now().Before(deadline) {
		typ, _, ok := host.tryReadFrame(150 * time.Millisecond)
		if !ok {
			continue
		}
		if typ == protocol.MsgLobbyRelay {
			t.Fatal("host received a relay it did not address")
		}
	}
}
