package loadgen

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

// observer is an observer client: TCP connect + manual WebSocket upgrade, then
// masked WS frames outbound and unmasked WS frames inbound, on the same
// connection as any other client on the port. The server does not switch an
// observer to a raw protocol stream.
type observer struct {
	conn  net.Conn
	rd    *bufio.Reader
	stats *Stats
	done  chan struct{}
	once  sync.Once
}

var observerMask = [4]byte{0x12, 0x34, 0x56, 0x78}

func dialObserver(host string, port int, stats *Stats) (*observer, error) {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}
	key := "dGhlIHNhbXBsZSBub25jZQ=="
	req := fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n", addr, key)
	if _, err := conn.Write([]byte(req)); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("upgrade write: %w", err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil || string(buf[:n]) == "" {
		_ = conn.Close()
		return nil, fmt.Errorf("upgrade read: %w", err)
	}
	// find the status line
	status := string(buf[:min(n, len(buf))])
	if len(status) < 12 || status[9:12] != "101" {
		_ = conn.Close()
		return nil, fmt.Errorf("upgrade failed: %q", status[:min(len(status), 60)])
	}
	_ = conn.SetReadDeadline(time.Time{})
	return &observer{
		conn:  conn,
		rd:    bufio.NewReader(conn),
		stats: stats,
		done:  make(chan struct{}),
	}, nil
}

// writeMsg sends one masked binary WS frame with the given payload.
func (o *observer) writeMsg(payload []byte) error {
	frame := make([]byte, 0, 10+len(payload))
	frame = append(frame, 0x82) // FIN | opcode 2 (binary)
	n := len(payload)
	switch {
	case n < 126:
		frame = append(frame, byte(n)|0x80)
	case n <= 0xFFFF:
		frame = append(frame, 126|0x80, byte(n>>8), byte(n))
	default:
		frame = append(frame, 127|0x80)
		var ext [8]byte
		binary.BigEndian.PutUint64(ext[:], uint64(n))
		frame = append(frame, ext[:]...)
	}
	frame = append(frame, observerMask[:]...)
	masked := make([]byte, n)
	for i := range masked {
		masked[i] = payload[i] ^ observerMask[i%4]
	}
	frame = append(frame, masked...)
	_, err := o.conn.Write(frame)
	return err
}

// readFrame reads one unmasked server-to-client WebSocket frame and splits the
// payload into [type][body].
//
// The server frames observer traffic like any other client on the port: every
// outbound send goes through the connection's own socket manager, and the
// per-peer observer send asks that manager for a frame in the connection's
// transport. This client used to read a raw [u16 BE length][type][body] stream
// instead, on the assumption that the observer path bypassed the WebSocket
// layer. It does not.
func (o *observer) readFrame(timeout time.Duration) (byte, []byte, bool) {
	_ = o.conn.SetReadDeadline(time.Now().Add(timeout))

	var hdr [2]byte
	if _, err := io.ReadFull(o.rd, hdr[:]); err != nil {
		return 0, nil, false
	}
	o.stats.recordRecv(2)
	opcode := hdr[0] & 0x0F
	masked := hdr[1]&0x80 != 0
	length := uint64(hdr[1] & 0x7F)

	switch length {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(o.rd, ext[:]); err != nil {
			return 0, nil, false
		}
		o.stats.recordRecv(2)
		length = uint64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(o.rd, ext[:]); err != nil {
			return 0, nil, false
		}
		o.stats.recordRecv(8)
		length = binary.BigEndian.Uint64(ext[:])
	}

	// A server must not mask, but tolerate it rather than misparse the stream.
	var mask [4]byte
	if masked {
		if _, err := io.ReadFull(o.rd, mask[:]); err != nil {
			return 0, nil, false
		}
		o.stats.recordRecv(4)
	}

	switch opcode {
	case 0x8: // close
		return 0, nil, false
	case 0x9: // ping -> pong, then keep reading
		payload := make([]byte, length)
		if _, err := io.ReadFull(o.rd, payload); err != nil {
			return 0, nil, false
		}
		o.stats.recordRecv(len(payload))
		if err := o.writePong(payload); err != nil {
			return 0, nil, false
		}
		return o.readFrame(timeout)
	case 0xA: // pong
		if length > 0 {
			if _, err := io.CopyN(io.Discard, o.rd, int64(length)); err != nil {
				return 0, nil, false
			}
			o.stats.recordRecv(int(length))
		}
		return o.readFrame(timeout)
	}

	if length < 1 {
		return 0, nil, false
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(o.rd, payload); err != nil {
		return 0, nil, false
	}
	o.stats.recordRecv(len(payload))
	if masked {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}
	return payload[0], payload[1:], true
}

// writePong answers a server ping so the connection is not reaped mid-match.
func (o *observer) writePong(payload []byte) error {
	frame := make([]byte, 0, 10+len(payload))
	frame = append(frame, 0x8A) // FIN | opcode 10 (pong)
	n := len(payload)
	switch {
	case n < 126:
		frame = append(frame, byte(n)|0x80)
	default:
		frame = append(frame, 126|0x80, byte(n>>8), byte(n))
	}
	frame = append(frame, observerMask[:]...)
	masked := make([]byte, n)
	for i := range masked {
		masked[i] = payload[i] ^ observerMask[i%4]
	}
	frame = append(frame, masked...)
	_ = o.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := o.conn.Write(frame)
	_ = o.conn.SetWriteDeadline(time.Time{})
	return err
}

// startObserve subscribes the observer to the session and waits for the
// status-3 StartObserveResponse (raw protocol after the first message).
// The message type 0x1C selects the Vx layout (titleID in the high bits of
// the first u64). 0x18 is the legacy bare-sessionID form; it is not used here
// because it is not known whether the server compares the titleID it forces
// internally (0xFFFF) during the lookup. See the open item in
// SESSION-JOURNAL.md.
func (o *observer) startObserve(sessionID uint64) error {
	if err := o.writeMsg(wireFrame(0x1C, startObserveBody(sessionID, 0))); err != nil {
		return fmt.Errorf("start observe: %w", err)
	}
	typ, body, ok := o.readFrame(5 * time.Second)
	if !ok || typ != 0x19 || len(body) != 1 || body[0] != 3 {
		return fmt.Errorf("StartObserveResp status != 3 (type=0x%02x body=%x)", typ, body)
	}
	return nil
}

// activate sends ActivateObserver and waits for the response.
func (o *observer) activate() error {
	if err := o.writeMsg(wireFrame(0x1a, nil)); err != nil {
		return fmt.Errorf("activate: %w", err)
	}
	typ, body, ok := o.readFrame(5 * time.Second)
	if !ok || typ != 0x1e || len(body) < 9 || body[0] != 0 {
		return fmt.Errorf("ActivateObserverResp not ok (type=0x%02x body=%x)", typ, body)
	}
	return nil
}

// run drains the observer page stream until done or the conn dies.
func (o *observer) run() {
	for {
		select {
		case <-o.done:
			return
		default:
		}
		typ, _, ok := o.readFrame(250 * time.Millisecond)
		if !ok {
			select {
			case <-o.done:
				return
			default:
			}
			continue
		}
		// 0x09 = replay pages, 0x22 = end-of-game marker; both count as the
		// expected observer stream.
		if typ != 0x09 && typ != 0x22 {
			o.stats.recordError()
		}
	}
}

func (o *observer) close() {
	o.once.Do(func() {
		close(o.done)
		_ = o.conn.Close()
	})
}
