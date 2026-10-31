package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/session"
	"github.com/luskaner/ageBattleServer/internal/testutil"
)

func TestNewClient(t *testing.T) {
	srv := createTestServer(t)
	conn := &mockNetConn{}

	client := NewClient(conn, srv)
	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.RemoteAddr() != "192.168.1.100:12345" {
		t.Errorf("expected remote addr '192.168.1.100:12345', got '%s'", client.RemoteAddr())
	}
}

func TestClientSendMessage(t *testing.T) {
	srv := createTestServer(t)
	conn := &mockNetConn{}

	client := NewClient(conn, srv)

	err := client.SendMessage(protocol.MsgPing, []byte{0x01, 0x02})
	if err != nil {
		t.Errorf("SendMessage error: %v", err)
	}

	err = client.Send([]byte{0x01, 0x02, 0x03})
	if err != nil {
		t.Errorf("Send error: %v", err)
	}
}

func TestClientClose(t *testing.T) {
	srv := createTestServer(t)
	conn := &mockNetConn{}

	client := NewClient(conn, srv)
	err := client.Close()
	if err != nil {
		t.Errorf("Close error: %v", err)
	}

	if !conn.closed {
		t.Error("connection should be closed")
	}
}

func TestClientDoubleClose(t *testing.T) {
	srv := createTestServer(t)
	conn := &mockNetConn{}

	client := NewClient(conn, srv)
	_ = client.Close()
	err := client.Close()
	if err != nil {
		t.Errorf("double close should not error: %v", err)
	}
}

func TestClientHandleMessagePing(t *testing.T) {
	srv := createTestServer(t)
	conn := &mockNetConn{}

	client := NewClient(conn, srv)
	client.handleMessage(protocol.MsgPing, nil)
}

func TestClientHandleMessageGameData(t *testing.T) {
	srv := createTestServer(t)
	conn := &mockNetConn{}

	client := NewClient(conn, srv)

	gdMsg := &protocol.GameDataMessage{
		InnerType: 0x01,
		Payload:   []byte{0x01, 0x02},
	}
	body := gdMsg.Serialize()
	client.handleMessage(protocol.MsgGameData, body)
}

func TestClientHandleMessageUnknown(t *testing.T) {
	srv := createTestServer(t)
	conn := &mockNetConn{}

	client := NewClient(conn, srv)
	client.handleMessage(0xFF, nil)
}

func TestBsProtocolHandshake(t *testing.T) {
	srv := createTestServer(t)
	clientConn, serverConn := net.Pipe()
	client := NewClient(clientConn, srv)
	_ = client.Start(context.Background())
	defer func() { _ = client.Close() }()

	// 1. lobby connect is echoed verbatim.
	connectBody := []byte{0x1d, 0x4d, 0x79, 0x3c, 0x00, 0x00, 0x18, 0x00, 0xcc, 0xd9, 0x48,
		0x38, 0xec, 0x7f, 0x6c, 0x55, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x2f, 0x0a, 0x77, 0x6a, 0x00, 0x00, 0x00, 0x00}
	if _, err := serverConn.Write(wireFrame(protocol.MsgLobbyConnect, connectBody)); err != nil {
		t.Fatalf("write connect: %v", err)
	}
	msgType, body := readFrameFrom(t, serverConn)
	if msgType != protocol.MsgLobbyConnect {
		t.Fatalf("expected connect echo type 0x15, got 0x%02x", msgType)
	}
	if !bytes.Equal(body, connectBody) {
		t.Fatalf("connect echo mismatch: %x", body)
	}

	// 2. the first relay carries the player's compressed RequestJoin for the
	// session it connected to: it is consumed and comes back as the peer (0x0f)
	// message it carried, the 3-byte relay envelope stripped.
	joinKey := session.SessionKey{TitleID: 24, SessionID: 0x3c794d1d}
	relayBody := relayBody(0x01, 0x01, compressedPeerMessage(t, requestJoinInner(joinKey, 0xab)))
	if _, err := serverConn.Write(wireFrame(protocol.MsgLobbyRelay, relayBody)); err != nil {
		t.Fatalf("write relay: %v", err)
	}
	msgType, body = readFrameFrom(t, serverConn)
	if msgType != protocol.MsgPeer {
		t.Fatalf("expected peer type 0x0f, got 0x%02x", msgType)
	}
	if !bytes.Equal(body, relayBody[3:]) {
		t.Fatalf("peer body mismatch: %x", body)
	}

	// 3. subsequent relays are echoed verbatim.
	relay2 := make([]byte, 411)
	relay2[0], relay2[1], relay2[2] = 0x01, 0x01, 0x0f
	copy(relay2[100:], []byte{0x63, 0x63, 0x60})
	if _, err := serverConn.Write(wireFrame(protocol.MsgLobbyRelay, relay2)); err != nil {
		t.Fatalf("write relay 2: %v", err)
	}
	msgType, body = readFrameFrom(t, serverConn)
	if msgType != protocol.MsgLobbyRelay {
		t.Fatalf("expected relay echo type 0x13, got 0x%02x", msgType)
	}
	if !bytes.Equal(body, relay2) {
		t.Fatalf("relay echo mismatch: %x", body)
	}

	// 4. small lobby pings are echoed verbatim.
	pingBody := []byte{0x01, 0x01, 0x0f, 0x00, 0x00, 0x00}
	pingBody = append(pingBody, make([]byte, 20)...)
	if _, err := serverConn.Write(wireFrame(protocol.MsgLobbyRelay, pingBody)); err != nil {
		t.Fatalf("write ping: %v", err)
	}
	msgType, body = readFrameFrom(t, serverConn)
	if msgType != protocol.MsgLobbyRelay || !bytes.Equal(body, pingBody) {
		t.Fatalf("ping echo mismatch: type=0x%02x body=%x", msgType, body)
	}
}

func TestBsProtocolGameDataAcks(t *testing.T) {
	srv := createTestServer(t)
	ctx := t.Context()
	if err := srv.Start(ctx); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer func() { _ = srv.Stop() }()

	clientConn, serverConn := net.Pipe()
	client := NewClient(clientConn, srv)
	_ = client.Start(context.Background())
	defer func() { _ = client.Close() }()

	// Acks are session-scoped: the client must join a session first.
	testSessionWithPeer(t, srv, client)

	// The responses are tick-driven (~125ms), with idle heartbeats in between
	// (solo session, capture _24 format); skip them and return the next
	// substantive response.
	readResponse := func() (byte, []byte) {
		for {
			msgType, body := readFrameFrom(t, serverConn)
			if msgType == protocol.MsgEndOfGameData && bytes.Equal(body, []byte{0xfe, 0x00, 0x01, 0x01, 0x00, 0x00}) {
				continue
			}
			return msgType, body
		}
	}

	// 1. gamedata is queued and flushed at the next simulation tick.
	gd := make([]byte, 40)
	gd[0] = 0xfe
	copy(gd[1:], []byte{0x25, 0x00, 0x01, 0x78, 0x01})
	if _, err := serverConn.Write(wireFrame(protocol.MsgGameData, gd)); err != nil {
		t.Fatalf("write gamedata: %v", err)
	}
	msgType, ack := readResponse()
	if msgType != protocol.MsgEndOfGameData {
		t.Fatalf("expected ack type 0x0a, got 0x%02x", msgType)
	}
	want := append([]byte{0xfe, 0x01, 0x01}, gd[1:]...)
	// Single-sub flush for the only player: empty pending list (00 00 00).
	want = append(want, 0x00, 0x00, 0x00)
	if !bytes.Equal(ack, want) {
		t.Fatalf("gamedata flush mismatch: %x", ack)
	}

	// 2. checksum message with a nested 09 fe chunk: the chunk is flushed at
	// the next tick and the entries acked standalone the tick after. Entries
	// are [crc u32 LE][page u32 LE] and deduplicated by page (both players
	// upload the same page checksums), so the two entries use distinct pages.
	// The chunk is [u16 LE length][payload], and the length must be the
	// payload's: the nested 0x09 is consumed exactly like a top-level one.
	chunk := append([]byte{58, 0x00}, bytes.Repeat([]byte{0xcd}, 58)...)
	cs := make([]byte, 1+16+2+len(chunk))
	cs[0] = 2
	binary.LittleEndian.PutUint32(cs[5:9], 1)
	binary.LittleEndian.PutUint32(cs[13:17], 2)
	cs[17], cs[18] = 0x09, 0xfe
	copy(cs[19:], chunk)
	if _, err := serverConn.Write(wireFrame(protocol.MsgGameDataChecksum, cs)); err != nil {
		t.Fatalf("write checksum: %v", err)
	}
	msgType, ack = readResponse()
	if msgType != protocol.MsgEndOfGameData {
		t.Fatalf("expected checksum flush type 0x0a, got 0x%02x", msgType)
	}
	want = append([]byte{0xfe, 0x01, 0x01}, chunk...)
	// Single-sub flush for the only player: empty pending list (00), then an
	// empty entries tail (00 00). The reconciled entries do NOT ride this frame:
	// the original sends the 0x0A to every peer and only then reconciles, so
	// anything reconciliation merges can only reach the peers on a later frame.
	want = append(want, 0x00, 0x00, 0x00)
	if !bytes.Equal(ack, want) {
		t.Fatalf("checksum flush mismatch: %x", ack)
	}

	// 3. checksum message without a chunk. The entries the reconciliation merged
	// after the previous tick go out on this frame on their own, with no chunk to
	// go with them.
	cs2 := make([]byte, 20)
	cs2[0] = 1
	if _, err := serverConn.Write(wireFrame(protocol.MsgGameDataChecksum, cs2)); err != nil {
		t.Fatalf("write checksum 2: %v", err)
	}
	msgType, ack = readResponse()
	if msgType != protocol.MsgEndOfGameData {
		t.Fatalf("expected standalone ack type 0x0a, got 0x%02x", msgType)
	}
	want = append([]byte{0xfe, 0x00, 0x01, 0x01, 0x13, 0x00, 0x02, 0x02, 0x00}, cs[1:17]...)
	if !bytes.Equal(ack, want) {
		t.Fatalf("standalone ack mismatch: %x", ack)
	}
}

func TestClientLobbyPingTick(t *testing.T) {
	srv := createTestServer(t)
	conn := &mockNetConn{}
	client := NewClient(conn, srv)

	now := time.Now().Add(11 * time.Second)
	client.lobbyPingTick(now, 0x02540000)
	client.lobbyPingTick(now.Add(5*time.Second), 0x02540001)

	queued := takeQueued(client)
	if len(queued) != 1 {
		t.Fatalf("queued %d lobby pings, want exactly one", len(queued))
	}
	d := queued[0].data
	if len(d) != 11 || d[2] != protocol.MsgLobbyPing {
		t.Fatalf("bad ping frame: %x", d)
	}
	if got := binary.LittleEndian.Uint32(d[3:7]); got != 0x02540000 {
		t.Fatalf("bad ping nonce: %08x", got)
	}
	if got := d[7:11]; !bytes.Equal(got, []byte{0, 0, 0, 0}) {
		t.Fatalf("bad ping tail: %x", got)
	}

	client.lobbyPingTick(now.Add(11*time.Second), 0x02540002)
	queued = takeQueued(client)
	if len(queued) != 1 {
		t.Fatalf("queued %d second lobby pings, want exactly one", len(queued))
	}
	if got := binary.LittleEndian.Uint32(queued[0].data[3:7]); got != 0x02540002 {
		t.Fatalf("bad second ping nonce: %08x", got)
	}
}

// takeQueued removes and returns the frames queued on a client whose writer is
// not running.
func takeQueued(c *Client) []outboundMsg {
	c.mu.Lock()
	defer c.mu.Unlock()
	q := c.sendQueue
	c.sendQueue = nil
	c.queuedBytes = 0
	return q
}

func TestClientProcess(t *testing.T) {
	srv := createTestServer(t)
	conn := &mockNetConn{}

	client := NewClient(conn, srv)
	client.Process(100, time.Now())
}

// TestClientProcessTimeout covers a connection with no timeout override: the
// default silent timeout closes it (TestClientSilentTimeout has the boundary
// cases).
func TestClientProcessTimeout(t *testing.T) {
	srv := createTestServer(t)
	conn := &mockNetConn{}

	client := NewClient(conn, srv)
	assertIdleTimeout(t,
		client.Process,
		func(tm time.Time) {
			client.mu.Lock()
			client.lastSeen = tm
			client.mu.Unlock()
		},
		func() bool {
			client.mu.Lock()
			defer client.mu.Unlock()
			return client.closed
		})
}

func TestClientSendMessageTooLarge(t *testing.T) {
	srv := createTestServer(t)
	client := NewClient(&mockNetConn{}, srv)

	if err := client.SendMessage(protocol.MsgPing, make([]byte, 65535)); err == nil {
		t.Error("SendMessage with 65535-byte body should error")
	}
	if err := client.SendMessage(protocol.MsgPing, make([]byte, 65534)); err != nil {
		t.Errorf("SendMessage with 65534-byte body should succeed: %v", err)
	}
}

func TestClientSendAfterClose(t *testing.T) {
	srv := createTestServer(t)
	client := NewClient(&mockNetConn{}, srv)
	if err := client.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	if err := client.Send([]byte{0x01}); err == nil {
		t.Error("Send after Close should error")
	}
	if err := client.SendMessage(protocol.MsgPing, nil); err == nil {
		t.Error("SendMessage after Close should error")
	}
}

func TestClientSetObserver(t *testing.T) {
	srv := createTestServer(t)
	client := NewClient(&mockNetConn{}, srv)

	if client.IsObserver() {
		t.Error("TCP client should never be an observer")
	}
	client.SetObserver(true) // no-op
	if client.IsObserver() {
		t.Error("TCP client should still not be an observer")
	}
}

func TestClientDeadlines(t *testing.T) {
	srv := createTestServer(t)
	client := NewClient(&mockNetConn{}, srv)

	testutil.CheckDeadlines(t, client.SetDeadline, client.SetReadDeadline, client.SetWriteDeadline)
}

func TestClientReadMessageEdges(t *testing.T) {
	pc := &readStubConn{}
	client := NewClient(pc, createTestServer(t))
	defer func() { _ = client.Close() }()

	// Zero declared length is an error.
	pc.push([]byte{0x00, 0x00})
	if _, _, err := client.readMessage(); err == nil {
		t.Error("zero-length message should error")
	}

	// Truncated body.
	pc.push([]byte{0x00, 0x05, 0x09, 0x01})
	if _, _, err := client.readMessage(); err == nil {
		t.Error("truncated body should error")
	}

	// Valid message.
	pc.push(wireFrame(protocol.MsgPing, []byte{0xAA}))
	typ, body, err := client.readMessage()
	if err != nil {
		t.Fatalf("readMessage error: %v", err)
	}
	if typ != protocol.MsgPing || !bytes.Equal(body, []byte{0xAA}) {
		t.Errorf("readMessage = %#x %x", typ, body)
	}
}

type readStubConn struct {
	buf []byte
}

func (s *readStubConn) push(data []byte) { s.buf = append(s.buf, data...) }

func (s *readStubConn) Read(b []byte) (int, error) {
	if len(s.buf) == 0 {
		return 0, io.EOF
	}
	n := copy(b, s.buf)
	s.buf = s.buf[n:]
	return n, nil
}

func (s *readStubConn) Write(b []byte) (int, error)   { return len(b), nil }
func (s *readStubConn) Close() error                  { return nil }
func (s *readStubConn) LocalAddr() net.Addr           { return nil }
func (s *readStubConn) RemoteAddr() net.Addr          { return nil }
func (s *readStubConn) SetDeadline(_ time.Time) error { return nil }
func (s *readStubConn) SetReadDeadline(_ time.Time) error {
	return nil
}
func (s *readStubConn) SetWriteDeadline(_ time.Time) error { return nil }

func TestClientReadLoopCancel(t *testing.T) {
	srv := createTestServer(t)
	clientConn, serverConn := net.Pipe()
	client := NewClient(clientConn, srv)

	ctx, cancel := context.WithCancel(context.Background())
	_ = client.Start(ctx)
	cancel()

	deadline := time.Now().Add(2 * time.Second)
	for {
		client.mu.Lock()
		closed := client.closed
		client.mu.Unlock()
		if closed || time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	client.mu.Lock()
	closed := client.closed
	client.mu.Unlock()
	if !closed {
		t.Error("client should be closed after context cancel")
	}
	_ = serverConn.Close()
}

// TestClientSendNeverWaitsBelowTheCap queues far more frames than a writer has
// taken (none is running here). Every Send below the byte cap returns at once
// and keeps its frame: a long queue is never a reason to drop one.
func TestClientSendNeverWaitsBelowTheCap(t *testing.T) {
	srv := createTestServer(t)
	client := NewClient(&mockNetConn{}, srv)

	start := time.Now()
	for i := range 2000 {
		if err := client.Send([]byte{byte(i)}); err != nil {
			t.Fatalf("Send %d on an open connection: %v", i, err)
		}
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Errorf("2000 sends took %v: a send waited", elapsed)
	}
	if clientClosed(client) {
		t.Error("the connection closed although the queue is far below the cap")
	}
}

// TestClientSendQueueOverflowClosesTheConnection fills the queue past its byte
// cap. The send that crosses it is refused AND the connection closes: a frame
// is never refused while the connection stays open.
func TestClientSendQueueOverflowClosesTheConnection(t *testing.T) {
	srv := createTestServer(t)
	conn := &mockNetConn{}
	client := NewClient(conn, srv)
	client.maxQueuedBytes = 1024

	body := make([]byte, 100) // a 103-byte frame
	for i := range 9 {
		if err := client.SendMessage(protocol.MsgEndOfGameData, body); err != nil {
			t.Fatalf("frame %d below the cap: %v", i, err)
		}
	}
	if clientClosed(client) {
		t.Fatal("closed below the cap")
	}
	if err := client.SendMessage(protocol.MsgEndOfGameData, body); err == nil {
		t.Error("the frame that crosses the cap was accepted")
	}
	if !clientClosed(client) || !conn.closed {
		t.Error("the connection is still open after its queue overflowed")
	}
}

func createTestServer(t *testing.T) *Server {
	cfg := &Config{
		BsPort:   0,
		LogLevel: "none",
	}

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	return srv
}

// twoPeerServer creates a server with two registered clients, each joined to a
// session as peers 1 and 2.
func twoPeerServer(t *testing.T) (*Server, *session.Session, *mockMsgClient, *mockMsgClient) {
	t.Helper()
	srv := createTestServer(t)
	c1 := &mockMsgClient{}
	c2 := &mockMsgClient{}
	srv.AddClient(c1)
	srv.AddClient(c2)

	sess := srv.sessionManager.CreateSession()
	p1 := sess.NewPeer(1, c1)
	p2 := sess.NewPeer(2, c2)
	srv.sessionManager.AddPeer(sess, p1)
	srv.sessionManager.AddPeer(sess, p2)
	return srv, sess, c1, c2
}

type mockNetConn struct {
	closed bool
	data   []byte
	local  net.Addr
	remote net.Addr
}

func (c *mockNetConn) Read(_ []byte) (int, error) {
	if c.closed {
		return 0, net.ErrClosed
	}
	return 0, nil
}

func (c *mockNetConn) Write(b []byte) (int, error) {
	if c.closed {
		return 0, net.ErrClosed
	}
	c.data = append(c.data, b...)
	return len(b), nil
}

func (c *mockNetConn) Close() error {
	c.closed = true
	return nil
}

func (c *mockNetConn) LocalAddr() net.Addr {
	return c.local
}

func (c *mockNetConn) RemoteAddr() net.Addr {
	if c.remote != nil {
		return c.remote
	}
	addr, _ := net.ResolveTCPAddr("tcp", "192.168.1.100:12345")
	return addr
}

func (c *mockNetConn) SetDeadline(_ time.Time) error {
	return nil
}

func (c *mockNetConn) SetReadDeadline(_ time.Time) error {
	return nil
}

func (c *mockNetConn) SetWriteDeadline(_ time.Time) error {
	return nil
}

// assertIdleTimeout verifies that Process closes a client whose lastSeen is
// past the default silent timeout.
func assertIdleTimeout(t *testing.T, process func(int, time.Time), setLastSeen func(time.Time), isClosed func() bool) {
	t.Helper()
	setLastSeen(time.Now().Add(-6 * time.Minute))
	process(100, time.Now())
	if !isClosed() {
		t.Error("client should be closed after idle timeout")
	}
}
