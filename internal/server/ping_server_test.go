package server

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestWsAcceptKey(t *testing.T) {
	// RFC 6455 sample.
	got := wsAcceptKey("dGhlIHNhbXBsZSBub25jZQ==")
	want := "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
	if got != want {
		t.Errorf("wsAcceptKey = %q, want %q", got, want)
	}
}

func TestWsFrameLengths(t *testing.T) {
	short := wsFrame([]byte{0x01, 0x02, 0x03})
	if len(short) != 2+3 || short[0] != 0x82 || short[1] != 3 {
		t.Errorf("short frame = %x", short)
	}

	mid := wsFrame(make([]byte, 126))
	if len(mid) != 2+2+126 || mid[1] != 126 || mid[2] != 0 || mid[3] != 126 {
		t.Errorf("126-byte frame = %x", mid[:6])
	}

	large := wsFrame(make([]byte, 70000))
	if len(large) != 2+8+70000 || large[1] != 127 {
		t.Fatalf("large frame header = %x", large[:10])
	}
	if large[2] != 0 || large[9] != 0x70 {
		t.Errorf("large frame length bytes = %x", large[2:10])
	}
}

// serveWSClient dials a serveWS peer over net.Pipe, performs the WS handshake,
// and returns the client conn, its buffered reader, and the serveWS-done signal.
func serveWSClient(t *testing.T) (net.Conn, *bufio.Reader, chan struct{}) {
	t.Helper()
	a, b := net.Pipe()
	t.Cleanup(func() { _ = a.Close() })
	t.Cleanup(func() { _ = b.Close() })
	done := make(chan struct{})
	go func() {
		serveWS(b, bufio.NewReader(b))
		close(done)
	}()
	_, _ = a.Write([]byte("GET / HTTP/1.1\r\nHost: x\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n\r\n"))
	br := bufio.NewReader(a)
	_ = a.SetReadDeadline(time.Now().Add(2 * time.Second))
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			t.Fatalf("read response: %v", err)
		}
		if line == "\r\n" {
			break
		}
	}
	return a, br, done
}

func TestServeWS(t *testing.T) {
	a, b := net.Pipe()
	defer func() { _ = a.Close() }()
	defer func() { _ = b.Close() }()

	done := make(chan struct{})
	go func() {
		serveWS(b, bufio.NewReader(b))
		close(done)
	}()

	req := "GET / HTTP/1.1\r\nHost: localhost\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n\r\n"
	if _, err := a.Write([]byte(req)); err != nil {
		t.Fatalf("write request: %v", err)
	}

	_ = a.SetReadDeadline(time.Now().Add(2 * time.Second))
	br := bufio.NewReader(a)
	resp := make([]byte, 0)
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			t.Fatalf("read response: %v", err)
		}
		resp = append(resp, line...)
		if line == "\r\n" {
			break
		}
	}
	if !strings.HasPrefix(string(resp), "HTTP/1.1 101") {
		t.Fatalf("expected 101 response, got %q", resp)
	}
	if !strings.Contains(string(resp), "Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=") {
		t.Fatalf("missing accept key in response: %q", resp)
	}

	// Masked 0x14-typed payload is echoed re-framed.
	ping := []byte{0x81, 0x81, 0x01, 0x02, 0x03, 0x04, 0x15}
	if _, err := a.Write(ping); err != nil {
		t.Fatalf("write ping frame: %v", err)
	}
	echo := make([]byte, 3)
	if _, err := br.Read(echo); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	want := wsFrame([]byte{0x14})
	if !bytes.Equal(echo, want) {
		t.Fatalf("echo = %x, want %x", echo, want)
	}

	// Close frame ends the loop.
	if _, err := a.Write([]byte{0x88, 0x00}); err != nil {
		t.Fatalf("write close: %v", err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("serveWS did not return after close frame")
	}
}

func TestServeWSNoEchoForOtherTypes(t *testing.T) {
	a, _, done := serveWSClient(t)

	// Non-0x14 payload is not echoed; the loop keeps reading.
	masked := []byte{0x82, 0x83, 0x01, 0x02, 0x03, 0x04, 0x09, 0x09, 0x09}
	_, _ = a.Write(masked)

	closeFrame := []byte{0x88, 0x80, 0x01, 0x02, 0x03, 0x04}
	_, _ = a.Write(closeFrame)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("serveWS did not return after close")
	}
}

func TestServeWSFrameSizes(t *testing.T) {
	a, br, done := serveWSClient(t)

	readEcho := func() []byte {
		var hdr [2]byte
		if _, err := io.ReadFull(br, hdr[:]); err != nil {
			t.Fatalf("read echo header: %v", err)
		}
		n := int64(hdr[1] & 0x7F)
		if n == 126 {
			var ext [2]byte
			if _, err := io.ReadFull(br, ext[:]); err != nil {
				t.Fatalf("read ext len: %v", err)
			}
			n = int64(ext[0])<<8 | int64(ext[1])
		} else if n == 127 {
			var ext [8]byte
			if _, err := io.ReadFull(br, ext[:]); err != nil {
				t.Fatalf("read ext8 len: %v", err)
			}
			for i := range 8 {
				n = n<<8 | int64(ext[i])
			}
		}
		out := make([]byte, n)
		if _, err := io.ReadFull(br, out); err != nil {
			t.Fatalf("read echo body: %v", err)
		}
		return out
	}

	// 126-length frame, unmasked, 0x14-prefixed payload.
	mid := make([]byte, 200)
	mid[0] = 0x14
	_, _ = a.Write(append([]byte{0x82, 0x7E, 0x00, 200}, mid...))
	echo := readEcho()
	if len(echo) != 200 || echo[0] != 0x14 {
		t.Fatalf("126-length echo mismatch: %d bytes, first=%x", len(echo), echo[0])
	}

	// 127-length frame.
	big := make([]byte, 70000)
	big[0] = 0x14
	frame := []byte{0x82, 0x7F, 0, 0, 0, 0, 0, 1, 0x11, 0x70}
	_, _ = a.Write(append(frame, big...))
	echo = readEcho()
	if len(echo) != 70000 || echo[0] != 0x14 {
		t.Fatalf("127-length echo mismatch: %d bytes", len(echo))
	}

	// Control ping -> pong with the same payload.
	_, _ = a.Write([]byte{0x89, 0x82, 0x01, 0x02, 0x03, 0x04, 0xAA ^ 0x01, 0xBB ^ 0x02})
	pong := make([]byte, 2+2)
	if _, err := io.ReadFull(br, pong); err != nil {
		t.Fatalf("read pong: %v", err)
	}
	if pong[0] != 0x8A || pong[1] != 2 || pong[2] != 0xAA || pong[3] != 0xBB {
		t.Fatalf("pong mismatch: %x", pong)
	}

	// Oversized frame (>1MB) terminates the loop.
	_, _ = a.Write([]byte{0x82, 0x7F, 0, 0, 0, 0, 0x10, 0x00, 0x01, 0x00})
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("serveWS did not return after oversized frame")
	}
}

func TestPingServerGetPortBeforeStart(t *testing.T) {
	s := NewPingServer(12345, "", "")
	if got := s.GetPort(); got != 0 {
		t.Errorf("GetPort before Start = %d, want 0", got)
	}
}

func TestPingServerGetPortAfterStart(t *testing.T) {
	s := NewPingServer(0, "", "")
	ctx := t.Context()
	if err := s.Start(ctx); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer s.Stop()

	if got := s.GetPort(); got == 0 {
		t.Error("GetPort after Start = 0, want valid port")
	}
}

func TestPingServerStopIdempotent(t *testing.T) {
	s := NewPingServer(0, "", "")
	s.Stop()
	s.Stop()
}

// BenchmarkGenerateDefaultCert measures the one-time boot-time generation of
// the default self-signed certificate (used when no -sslCert/-sslKey is given).
// Before P11 it was RSA-2048; after it is ECDSA P-256, whose key generation is
// far faster and whose certificate is far smaller on the wire for every client
// TLS handshake.
func BenchmarkGenerateDefaultCert(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := generateDefaultCert(); err != nil {
			b.Fatal(err)
		}
	}
}
