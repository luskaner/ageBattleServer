package main

import (
	"bytes"
	"encoding/base64"
	"os"
	"strings"
	"testing"
)

func readAllFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

func frame(msgType byte, body []byte) []byte {
	out := make([]byte, 0, 3+len(body))
	out = append(out, byte((1+len(body))>>8), byte(1+len(body)), msgType)
	out = append(out, body...)
	return out
}

func TestFrameAssemblerReassemblesSegmented(t *testing.T) {
	asm := newFrameAssembler(frameRLink)
	a := frame(0x09, bytes.Repeat([]byte{0xAB}, 300))
	b := frame(0x15, []byte{0x01})

	var got [][]byte
	// Split a single frame across three reads.
	got = append(got, asm.feed(a[:5])...)
	if len(got) != 0 {
		t.Fatalf("partial first read must not emit a frame, got %d", len(got))
	}
	got = append(got, asm.feed(a[5:])...)
	if len(got) != 1 {
		t.Fatalf("completed frame must emit exactly one record, got %d", len(got))
	}
	if !bytes.Equal(got[0], a) {
		t.Fatal("reassembled frame differs from the original")
	}
	// Multiple frames coalesced in one read: two more records on top of the
	// one already reassembled above.
	got = append(got, asm.feed(append(b, b...))...)
	if len(got) != 3 {
		t.Fatalf("coalesced read must emit two records, got %d total", len(got))
	}
	for _, g := range got[1:] {
		if !bytes.Equal(g, b) {
			t.Fatal("coalesced frame differs from the original")
		}
	}
}

func TestFrameAssemblerSplitAtBoundary(t *testing.T) {
	asm := newFrameAssembler(frameRLink)
	a := frame(0x09, bytes.Repeat([]byte{0x01}, 100))
	b := frame(0x66, []byte{0xAA})
	// Deliver the first frame plus the first byte of the second.
	got := asm.feed(append(a, b[:1]...))
	if len(got) != 1 || !bytes.Equal(got[0], a) {
		t.Fatalf("expected exactly the first complete frame, got %d", len(got))
	}
	// The leftover byte must combine with the next read.
	got = append(got, asm.feed(b[1:])...)
	if len(got) != 2 {
		t.Fatalf("expected one frame after the second read, got %d", len(got))
	}
	if !bytes.Equal(got[1], b) {
		t.Fatalf("leftover-combined frame mismatch: got %x want %x", got[1], b)
	}
}

func TestFrameAssemblerHandshakePassthrough(t *testing.T) {
	asm := newFrameAssembler(frameWebSocket)
	handshake := []byte("GET / HTTP/1.1\r\nUpgrade: websocket\r\n\r\n")
	got := asm.feed(handshake)
	if len(got) != 1 || !bytes.Equal(got[0], handshake) {
		t.Fatalf("handshake text should pass through raw, got %d records", len(got))
	}
}

func TestRlinkFrameLen(t *testing.T) {
	b := frame(0x09, bytes.Repeat([]byte{0x00}, 10))
	if n := rlinkFrameLen(b); n != len(b) {
		t.Fatalf("rlinkFrameLen = %d, want %d", n, len(b))
	}
	if n := rlinkFrameLen(b[:2]); n != 0 {
		t.Fatalf("partial header should report 0, got %d", n)
	}
	// A declared length of zero is impossible: framing error.
	if n := rlinkFrameLen([]byte{0x00, 0x00, 0xAA}); n != -1 {
		t.Fatalf("zero declared length should report -1, got %d", n)
	}
}

// A zero RLink length used to stall reassembly forever (reported as
// "incomplete"); it must flush the buffered bytes raw so the capture
// resynchronizes on the next real frame.
func TestFrameAssemblerRLinkZeroLengthResyncs(t *testing.T) {
	asm := newFrameAssembler(frameRLink)
	good := frame(0x15, []byte{0x01})
	got := asm.feed([]byte{0x00, 0x00, 0xFF})
	got = append(got, asm.feed(good)...)
	if len(got) != 2 {
		t.Fatalf("expected raw flush + recovered frame, got %d records", len(got))
	}
	if !bytes.Equal(got[0], []byte{0x00, 0x00, 0xFF}) {
		t.Fatalf("raw flush mismatch: %x", got[0])
	}
	if !bytes.Equal(got[1], good) {
		t.Fatal("frame after resync differs from the original")
	}
}

// maskedWSFrame builds a client-style masked WebSocket binary frame.
func maskedWSFrame(payload []byte, mask [4]byte) []byte {
	out := make([]byte, 0, 2+4+len(payload))
	out = append(out, 0x82, byte(0x80|len(payload)))
	out = append(out, mask[:]...)
	for i, b := range payload {
		out = append(out, b^mask[i%4])
	}
	return out
}

// Client WebSocket frames are always masked; the capture must log the real
// message bytes (payload unmasked) while keeping the frame structure intact.
func TestFrameAssemblerUnmasksClientFrames(t *testing.T) {
	asm := newFrameAssembler(frameWebSocket)
	mask := [4]byte{0x12, 0x34, 0x56, 0x78}
	payload := []byte{0x15, 0x6D, 0xDB, 0x4F, 0x00, 0x01}
	wire := maskedWSFrame(payload, mask)

	got := asm.feed(wire)
	if len(got) != 1 || len(got[0]) != len(wire) {
		t.Fatalf("expected one complete frame of %d bytes, got %d records", len(wire), len(got))
	}
	f := got[0]
	if f[1]&0x80 == 0 {
		t.Fatal("mask bit must be preserved in the logged frame")
	}
	unmasked := f[6:]
	if !bytes.Equal(unmasked, payload) {
		t.Fatalf("payload not unmasked: got %x want %x", unmasked, payload)
	}
}

// An absurd WebSocket length is a framing error and must flush raw instead of
// buffering forever.
func TestFrameAssemblerWSOversizedLengthFlushesRaw(t *testing.T) {
	asm := newFrameAssembler(frameWebSocket)
	raw := make([]byte, 12)
	raw[0] = 0x82
	raw[1] = 0x7F // 64-bit length follows
	for i := 2; i < 10; i++ {
		raw[i] = 0xFF
	}
	got := asm.feed(raw)
	if len(got) != 1 || !bytes.Equal(got[0], raw) {
		t.Fatalf("oversized length must flush raw, got %d records", len(got))
	}
}

func TestDecodePacketCSVRow(t *testing.T) {
	data := frame(0x11, []byte{0xAA})
	p := decodePacket(data, "127.0.0.1", "13001", "127.0.0.1", "14001")
	if p.TypeName != "Ping" {
		t.Fatalf("TypeName = %q, want Ping", p.TypeName)
	}
	row := p.row()
	if row[9] != base64.StdEncoding.EncodeToString(data) {
		t.Fatal("data_b64 column must contain the raw frame base64")
	}
	if strings.Join(csvHeader, ",") != "timestamp,src_host,src_port,dest_host,dest_port,msg_type,type_name,body_len,body_hex,data_b64" {
		t.Fatalf("unexpected CSV header: %v", csvHeader)
	}
}

func TestCSVLoggerWritesProgressively(t *testing.T) {
	l := NewCSVLogger(t.TempDir() + "/capture.csv")
	l.LogMain("a", 1, "b", 2, frame(0x11, nil))
	l.LogMain("a", 1, "b", 2, frame(0x09, []byte{0xCC}))
	if err := l.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	raw, err := readAllFile(l.logName)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected header + 2 rows, got %d lines", len(lines))
	}
	if lines[0] != strings.Join(csvHeader, ",") {
		t.Fatalf("unexpected header line: %q", lines[0])
	}
}

// TestSecondaryLoggerSeparateFile ensures LogSecondary writes to its own CSV
// (SECONDARY_LOG_PATH) and not into the main log: regression for the proxy
// writing both flows into main_log.csv and never creating secondary_log.csv.
func TestSecondaryLoggerSeparateFile(t *testing.T) {
	dir := t.TempDir()
	mainLogger := NewCSVLogger(dir + "/main_log.csv")
	secLogger := NewCSVLogger(dir + "/secondary_log.csv")

	// Primary flow: client -> Target 1 response routed back to the client.
	mainLogger.LogMain("127.0.0.1", 14001, "192.168.1.5", 50789, frame(0x11, []byte{0xAA}))
	// Secondary flow: Target 2 response that gets drained.
	secLogger.LogSecondary("127.0.0.1", 15001, "192.168.1.5", 50789, frame(0x09, []byte{0xBB}))

	if err := mainLogger.Close(); err != nil {
		t.Fatalf("main Close error: %v", err)
	}
	if err := secLogger.Close(); err != nil {
		t.Fatalf("secondary Close error: %v", err)
	}

	mainRaw, err := readAllFile(mainLogger.logName)
	if err != nil {
		t.Fatalf("read main csv: %v", err)
	}
	secRaw, err := readAllFile(secLogger.logName)
	if err != nil {
		t.Fatalf("read secondary csv: %v", err)
	}

	mainLines := strings.Split(strings.TrimSpace(mainRaw), "\n")
	secLines := strings.Split(strings.TrimSpace(secRaw), "\n")
	if len(mainLines) != 2 {
		t.Fatalf("main log: expected header + 1 row, got %d lines", len(mainLines))
	}
	if len(secLines) != 2 {
		t.Fatalf("secondary log: expected header + 1 row, got %d lines", len(secLines))
	}
	if strings.Contains(mainRaw, "0x09") {
		t.Fatal("main log must not contain the secondary (Target 2) frame")
	}
	if !strings.Contains(secRaw, "0x09") {
		t.Fatal("secondary log must contain the Target 2 frame")
	}
}
