package main

// Cross-package check between the three halves of a capture.
//
// The CSV format is written here, in package main, and read in internal/sim, and
// Go forbids package main from being imported, so nothing in the tree can
// normally prove that the writer and the reader still agree. That is a real
// hazard: change a column and the harness stops reading every capture, SILENTLY,
// and the only symptom is that the first real capture produces an empty report
// and everybody assumes the game is broken.
//
// So the three halves are tied together here instead: the REAL framer decides
// which bytes are a frame, the REAL logger writes the row, and the REAL sim.Load
// reads it back. There are exactly two captures the proxy can make and both go
// through this test:
//
//   - webSocketPort: an HTTP upgrade and then WebSocket frames, one per row.
//   - bsPort: the RLink message itself, [u16 BE length][type][body], one per row,
//     with no 0x82 row anywhere in the file.
//
// The two differ in the msg_type column too: the proxy labels a WebSocket frame
// 0x82 and an RLink row with the message's own type (0x0A, 0x13, 0x09, ...), so a
// reader that took the column for the framing would read NOTHING out of the other
// shape and would return it as a clean capture with nothing to say.
// TestMsgTypeIsNotWhatTheReaderReads is the assertion that the column is
// decoration: the same bytes with the column rewritten are the same capture.
//
// An earlier revision of this file passed over a capture that framed no message
// at all: it fed eight arbitrary bytes to the logger and then checked that the
// reader had found a connection. That is the same silent void one layer down, and
// it is why the reader now refuses a log whose rows decode but frame to nothing
// (internal/sim.loadLog). TestProxyReaderRejectsALogItCannotFrame is the
// falsification, for each of the two shapes.

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/sim"
)

const (
	crosscheckBsPort     = 14001
	crosscheckSecBsPort  = 15001
	crosscheckWSPort     = 14003
	crosscheckSecWSPort  = 15003
	crosscheckClientPort = 51002
	crosscheckHost       = "192.168.0.1"
	crosscheckServerHost = "127.0.0.1"
)

// crosscheckShape is one of the two framings the relay speaks, with the port it
// is spoken on and the framing the reader has to decide for itself.
type crosscheckShape struct {
	name string
	mode frameMode
	port int
	sec  int
	want string
}

// The shapes are the two ports cmd/proxy/tcp_proxy.go:228 switches on.
var crosscheckShapes = []crosscheckShape{
	{name: "bsPort", mode: frameRLink, port: crosscheckBsPort, sec: crosscheckSecBsPort, want: sim.FramingRLink},
	{name: "webSocketPort", mode: frameWebSocket, port: crosscheckWSPort, sec: crosscheckSecWSPort, want: sim.FramingWS},
}

// frameOf builds one frame of this shape as the GAME puts it on the wire, which is
// not the same as what the capture holds: a WebSocket client masks its frames, and
// the framer unmasks them in place (cmd/proxy/framer.go:53), so the row keeps the
// key and the payload behind it is plain. A server frame carries no mask, and an
// RLink frame is the length prefix and the message and nothing else.
//
// The mask matters here and is not decoration: feed this a plain payload and the
// framer masks it a second time, and the reader is right not to unmask what it is
// given (internal/sim.wsPayload), so the capture would hold nonsense and still
// frame cleanly.
func (s crosscheckShape) frameOf(msgType byte, body []byte, client bool) []byte {
	payload := append([]byte{msgType}, body...)
	if s.mode != frameWebSocket {
		n := len(payload)
		return append([]byte{byte(n >> 8), byte(n)}, payload...)
	}
	var key [4]byte = [4]byte{0x12, 0x34, 0x56, 0x78}
	var f []byte
	if client {
		f = append(f, 0x82, byte(len(payload))|0x80)
		f = append(f, key[:]...)
		masked := make([]byte, len(payload))
		for i, b := range payload {
			masked[i] = b ^ key[i%4]
		}
		return append(f, masked...)
	}
	f = append(f, 0x82, byte(len(payload)))
	return append(f, payload...)
}

var (
	crosscheckUpgradeRequest = []byte("GET / HTTP/1.1\r\nHost: 127.0.0.1:14003\r\nUpgrade: websocket\r\n" +
		"Connection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n")
	crosscheckUpgradeResponse = []byte("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=\r\n\r\n")
)

// What the game and the server say in this capture. Written out here and not taken
// from the code under test, because a fixture built from the implementation
// cannot fail when the implementation is wrong, which is the one thing a fixture
// is for.
var (
	crosscheckLobbyConnect = protocol.BuildLobbyConnect(0x1234, 0x2a, 0x0123456789abcdef, 0, 0, 5, 0, 0)
	crosscheckPeer         = []byte{0x04, 0, 0, 0, 0xff, 0xff, 0xff, 0xff}
	crosscheckAck          = []byte{0xfe, 0x00, 0x01, 0x01, 0x00, 0x00}
	crosscheckStatus       = []byte{0x00}
)

// crosscheckMsg is one message of the exchange both sides of the capture see.
type crosscheckMsg struct {
	msgType byte
	body    []byte
	client  bool
}

// crosscheckExchange is the whole conversation: the client connects, uploads a
// page and asks to observe, and the original answers. The original answers 0x19
// with status 0 and the store's first ack; ours answers 3 and serves nothing, and
// that difference is the whole StartObserve bug expressed as bytes.
func crosscheckExchange() []crosscheckMsg {
	return []crosscheckMsg{
		{protocol.MsgLobbyConnect, crosscheckLobbyConnect, true},
		{protocol.MsgGameData, (&protocol.GameDataMessage{InnerType: 0xFE, Payload: []byte{0xAA, 0x01}}).Serialize(), true},
		{protocol.MsgLobbyConnect, crosscheckLobbyConnect, false},
		{protocol.MsgPeer, crosscheckPeer, false},
		{protocol.MsgEndOfGameData, crosscheckAck, false},
		{protocol.MsgStartObserveResp, crosscheckStatus, false},
	}
}

// writeCrosscheckCapture writes a capture of shape s into dir with the production
// logger, through the production framer: what reaches the CSV is exactly what a
// capture of this transport holds.
//
// Every frame is fed to the assembler in two halves, because that is what a TCP
// read landing in the middle of a frame looks like, and a reader that only coped
// with one frame per row would not be the reader a real capture needs.
func writeCrosscheckCapture(t *testing.T, dir string, s crosscheckShape) {
	t.Helper()
	main := NewCSVLogger(filepath.Join(dir, "main.log"))
	sec := NewCSVLogger(filepath.Join(dir, "secondary.log"))

	up := newFrameAssembler(s.mode)
	down := newFrameAssembler(s.mode)
	log := func(frame []byte, client bool) {
		if client {
			main.LogMain(crosscheckHost, crosscheckClientPort, crosscheckServerHost, s.port, frame)
			sec.LogSecondary(crosscheckHost, crosscheckClientPort, crosscheckServerHost, s.sec, frame)
			return
		}
		main.LogMain(crosscheckServerHost, s.port, crosscheckHost, crosscheckClientPort, frame)
		sec.LogSecondary(crosscheckServerHost, s.sec, crosscheckHost, crosscheckClientPort, frame)
	}
	feed := func(asm *frameAssembler, chunks ...[]byte) {
		for _, chunk := range chunks {
			for _, frame := range asm.feed(chunk) {
				log(frame, asm == up)
			}
		}
	}

	// The upgrade is HTTP, which cannot start a frame: the framer flushes it raw
	// and the reader has to skip it. Only a WebSocket connection carries one.
	if s.mode == frameWebSocket {
		feed(up, crosscheckUpgradeRequest[:60], crosscheckUpgradeRequest[60:])
		feed(down, crosscheckUpgradeResponse[:40], crosscheckUpgradeResponse[40:])
	}
	for _, m := range crosscheckExchange() {
		frame := s.frameOf(m.msgType, m.body, m.client)
		half := len(frame) / 2
		asm := down
		if m.client {
			asm = up
		}
		feed(asm, frame[:half], frame[half:])
	}
	if err := main.Close(); err != nil {
		t.Fatalf("close main logger: %v", err)
	}
	if err := sec.Close(); err != nil {
		t.Fatalf("close secondary logger: %v", err)
	}
}

// TestProxyWriterAndSimReaderAgreeOnBothTransports is the tie itself, for each of
// the two captures the proxy can make. The assertions are about MESSAGES and not
// about the shape of the result: a reader that found the connections and framed
// nothing out of them would pass a weaker version of this, which is exactly what
// an earlier revision of this file did.
func TestProxyWriterAndSimReaderAgreeOnBothTransports(t *testing.T) {
	for _, s := range crosscheckShapes {
		t.Run(s.name, func(t *testing.T) {
			dir := t.TempDir()
			writeProxyMetadata(t, dir)
			writeCrosscheckCapture(t, dir, s)

			cap, err := sim.Load(dir)
			if err != nil {
				t.Fatalf("sim.Load could not read a capture written by the proxy's own logger: %v", err)
			}
			// The truth comes from metadata.json, not from the file name.
			if cap.Reference.File != "main.log" || cap.Reference.Writer != sim.WriterOriginal {
				t.Fatalf("the reference should be main.log from the original, it is %q from %q", cap.Reference.File, cap.Reference.Writer)
			}
			if len(cap.Conns) != 1 {
				t.Fatalf("one client connection and no more, there are %d", len(cap.Conns))
			}
			conn := cap.Conns[0]
			if conn.ClientPort != crosscheckClientPort {
				t.Errorf("the address was resolved wrong: client %d, expected %d", conn.ClientPort, crosscheckClientPort)
			}
			if conn.Transport != s.want {
				t.Errorf("the transport was decided as %q and the case is %s (notes: %v)", conn.Transport, s.name, conn.Notes)
			}
			if conn.TransportNote != "" {
				t.Errorf("a single-transport connection should leave no doubt about it: %s", conn.TransportNote)
			}

			wantDown := []byte{protocol.MsgLobbyConnect, protocol.MsgPeer, protocol.MsgEndOfGameData, protocol.MsgStartObserveResp}
			if got := typesOf(conn.Down); !equalTypes(got, wantDown) {
				t.Errorf("what the original answered is not what was written.\n  read:     %s\n  expected: %s", renderTypes(got), renderTypes(wantDown))
			}
			wantUp := []byte{protocol.MsgLobbyConnect, protocol.MsgGameData}
			if got := typesOf(conn.Up); !equalTypes(got, wantUp) {
				t.Errorf("what the client sent has not been read whole.\n  read:     %s\n  expected: %s", renderTypes(got), renderTypes(wantUp))
			}
			if conn.Down[3].Body[0] != 0 {
				t.Errorf("the original's 0x19 carries status 0, %v was read", conn.Down[3])
			}
			if !bytesEqual(conn.Up[0].Body, crosscheckLobbyConnect) {
				t.Errorf("the client's LobbyConnect body did not come out whole: %s", hexOf(conn.Up[0].Body))
			}
			// The upgrade is HTTP and not a battle message: it has to be skipped
			// and not read as one, and not counted as an unfinished frame either.
			if s.mode == frameWebSocket {
				skipped := false
				for _, n := range conn.Notes {
					if strings.Contains(n, "WebSocket handshake") {
						skipped = true
					}
					if strings.Contains(n, "trailing bytes") {
						t.Errorf("the handshake has been counted as bytes of an unfinished frame: %s", n)
					}
				}
				if !skipped {
					t.Errorf("a connection with an upgrade must say how much HTTP it skipped, notes: %v", conn.Notes)
				}
			}
		})
	}
}

// TestMsgTypeIsNotWhatTheReaderReads: msg_type is the proxy's guess at what a
// frame holds, so it reads 0x82 on a WebSocket capture and the message's own type
// on a bsPort one. The reader takes neither, and the same bytes with the column
// rewritten are the same capture. A reader that filtered on 0x82 would drop every
// row of every bsPort capture, and a capture whose rows are all dropped compares
// nothing and looks clean.
func TestMsgTypeIsNotWhatTheReaderReads(t *testing.T) {
	for _, s := range crosscheckShapes {
		t.Run(s.name, func(t *testing.T) {
			dir := t.TempDir()
			writeProxyMetadata(t, dir)
			writeCrosscheckCapture(t, dir, s)
			before, err := sim.Load(dir)
			if err != nil {
				t.Fatalf("sim.Load: %v", err)
			}
			// What this capture says each row is, so the two shapes are shown to
			// differ here and not only in theory.
			t.Logf("msg_type the proxy wrote: %s", strings.Join(msgTypeColumn(t, filepath.Join(dir, "main.log")), " "))

			rewriteColumn(t, filepath.Join(dir, "main.log"), 5, "0x00")
			rewriteColumn(t, filepath.Join(dir, "secondary.log"), 5, "0xFF")
			after, err := sim.Load(dir)
			if err != nil {
				t.Fatalf("with the msg_type column rewritten the reader has to read the same: %v", err)
			}
			if len(after.Conns) != len(before.Conns) {
				t.Fatalf("changing msg_type has changed how many connections there are: %d before, %d after", len(before.Conns), len(after.Conns))
			}
			for i := range before.Conns {
				b, a := before.Conns[i], after.Conns[i]
				if b.Transport != a.Transport {
					t.Errorf("changing msg_type has changed the transport: %q and %q", b.Transport, a.Transport)
				}
				if !equalTypes(typesOf(b.Up), typesOf(a.Up)) || !equalTypes(typesOf(b.Down), typesOf(a.Down)) {
					t.Errorf("changing msg_type has changed the messages read: %s/%s against %s/%s",
						renderTypes(typesOf(b.Up)), renderTypes(typesOf(b.Down)),
						renderTypes(typesOf(a.Up)), renderTypes(typesOf(a.Down)))
				}
				for j := range b.Down {
					if !bytesEqual(b.Down[j].Body, a.Down[j].Body) {
						t.Errorf("the body of downstream message %d changed when msg_type was rewritten", j)
					}
				}
			}
		})
	}
}

// TestProxyReaderRejectsALogItCannotFrame is the falsification for both shapes: a
// capture whose rows are all there, all decode, and frame no message at all must
// be refused rather than returned as an empty capture. Returning it is the one
// outcome this harness must never produce, because a test over it passes and its
// report says nothing.
func TestProxyReaderRejectsALogItCannotFrame(t *testing.T) {
	for _, s := range crosscheckShapes {
		t.Run(s.name, func(t *testing.T) {
			dir := t.TempDir()
			writeProxyMetadata(t, dir)
			writeCrosscheckCapture(t, dir, s)
			// Three bytes that start no frame of either framing: a zero length is
			// impossible on this protocol, and the high bit a WebSocket head needs
			// is not set either.
			unframable := base64.StdEncoding.EncodeToString([]byte{0x00, 0x00, 0x00})
			for _, name := range []string{"main.log", "secondary.log"} {
				rewriteColumn(t, filepath.Join(dir, name), 9, unframable)
			}
			cap, err := sim.Load(dir)
			if err == nil {
				t.Fatalf("the reader swallowed %d connection(s) of a log that does not frame and returned a clean capture", len(cap.Conns))
			}
			if !strings.Contains(err.Error(), "not one message could be framed") {
				t.Fatalf("the error does not explain that the framing does not add up: %v", err)
			}
			t.Logf("detected as it must be: %v", err)
		})
	}
}

// TestProxyReaderRejectsACorruptLog is the same falsification one layer down: the
// payload column emptied, which is what a renamed or reordered data_b64 looks
// like from the reader's side.
func TestProxyReaderRejectsACorruptLog(t *testing.T) {
	dir := t.TempDir()
	writeProxyMetadata(t, dir)

	const header = "timestamp,src_host,src_port,dest_host,dest_port,msg_type,type_name,body_len,body_hex,data_b64"
	// The ports still parse, so the row counts as a row; only the payload is
	// missing.
	row := "2026-09-28T04:00:00+02:00,192.168.0.1,51002,127.0.0.1,14001,0x82,0x82,3,,"
	for _, name := range []string{"main.log", "secondary.log"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(header+"\n"+row+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	cap, err := sim.Load(dir)
	if err == nil {
		// The one outcome that must not happen: a clean, empty capture. That would
		// pass every test and produce a blank report, and the first sign of a
		// format change would be a real capture coming out empty.
		t.Fatalf("the reader swallowed a CSV with no payload and returned a clean capture: %+v", cap.Conns)
	}
	if !strings.Contains(err.Error(), "not one of them could be read") {
		t.Fatalf("the error does not explain that the format does not add up: %v", err)
	}
	t.Logf("detected as it must be: %v", err)
}

func writeProxyMetadata(t *testing.T, dir string) {
	t.Helper()
	meta := map[string]any{
		"format_version": 1,
		"started_at":     "2026-09-28 04:00:00",
		"commit":         "0000000000000000000000000000000000000000",
		"dirty":          "false",
		"primary": map[string]any{
			"writer":   "original",
			"exe_path": filepath.Join(dir, "BattleServerOrig.exe"),
			"sha256":   strings.Repeat("a", 64),
		},
		"secondary": map[string]any{
			"writer":   "ours",
			"exe_path": filepath.Join(dir, "BattleServerReimpl.exe"),
			"sha256":   strings.Repeat("b", 64),
		},
		"primary_args": []string{
			"-relaybroadcastPort", "0",
			"-bsPort", itoa(crosscheckBsPort),
			"-outofbandPort", "14002",
			"-websocketPort", itoa(crosscheckWSPort),
		},
		"secondary_args": []string{
			"-relaybroadcastPort", "0",
			"-bsPort", itoa(crosscheckSecBsPort),
			"-outofbandPort", "15002",
			"-websocketPort", itoa(crosscheckSecWSPort),
		},
	}
	body, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
}

// rewriteColumn replaces one column of a capture log, row by row, and writes the
// file back. Used to mangle a capture the way a renamed column or an unreadable
// payload would look from the reader's side.
func rewriteColumn(t *testing.T, path string, col int, value string) {
	t.Helper()
	rows, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	recs, err := csv.NewReader(strings.NewReader(string(rows))).ReadAll()
	if err != nil {
		t.Fatalf("%s: %v", filepath.Base(path), err)
	}
	for i := 1; i < len(recs); i++ {
		if col >= len(recs[i]) {
			t.Fatalf("%s: row %d has %d columns and column %d does not exist", filepath.Base(path), i, len(recs[i]), col)
		}
		recs[i][col] = value
	}
	out := &strings.Builder{}
	w := csv.NewWriter(out)
	if err := w.WriteAll(recs); err != nil {
		t.Fatal(err)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(out.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

// msgTypeColumn is what the proxy labelled each row of a log, so a test can show
// what a capture says about itself.
func msgTypeColumn(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	head, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	idx := -1
	for i, h := range head {
		if strings.TrimSpace(h) == "msg_type" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("%s has no msg_type column", filepath.Base(path))
	}
	var out []string
	for {
		rec, err := r.Read()
		if err != nil {
			break
		}
		out = append(out, rec[idx])
	}
	return out
}

func typesOf(msgs []*sim.Msg) []byte {
	out := make([]byte, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, m.Type)
	}
	return out
}

func equalTypes(a, b []byte) bool { return string(a) == string(b) }

func renderTypes(t []byte) string {
	parts := make([]string, 0, len(t))
	for _, x := range t {
		parts = append(parts, fmt0x(x))
	}
	return strings.Join(parts, " ")
}

func fmt0x(b byte) string { return "0x" + strings.ToUpper(strconv.FormatUint(uint64(b), 16)) }

func hexOf(b []byte) string {
	parts := make([]string, 0, len(b))
	for _, x := range b {
		parts = append(parts, strconv.FormatUint(uint64(x), 16))
	}
	return strings.Join(parts, " ")
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func itoa(n int) string { return strconv.Itoa(n) }
