package sim_test

// What every committed capture has to be, whatever transport it was made on.
//
// A capture is a test, and a test over a capture the reader did not read is a test
// that passes over nothing. That is not a hypothetical: the reader used to decide
// a connection's framing from the msg_type column, which reads 0x82 on a
// WebSocket capture and the message's own type on a bsPort one, so a bsPort
// capture came back as a clean capture with no messages in it, and the generated
// test for it went green without comparing anything. The refusal in
// internal/sim.loadLog is the reader's half of the answer; this is the harness'
// half, and it covers the captures themselves rather than a fixture.
//
// It reads and asserts, and never replays: a replay costs what the recorded run
// cost, and what is being checked here is what the READER made of the bytes.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/sim"
)

// TestEveryCommittedCaptureIsActuallyRead loads every capture under testdata and
// requires that each one came back with something to compare.
func TestEveryCommittedCaptureIsActuallyRead(t *testing.T) {
	captures := committedCaptures(t)
	if len(captures) == 0 {
		t.Fatal("there is no capture at all under testdata: this test would not be checking anything")
	}
	for _, name := range captures {
		t.Run(name, func(t *testing.T) {
			cap, err := sim.Load(filepath.Join("testdata", name))
			if err != nil {
				t.Fatalf("a committed capture the reader cannot read: %v", err)
			}
			if len(cap.Conns) == 0 {
				t.Fatalf("the capture produced no connection: nothing would be compared")
			}
			truthMessages := 0
			for _, conn := range cap.Conns {
				if conn.Transport != sim.FramingWS && conn.Transport != sim.FramingRLink {
					t.Errorf("connection %d: transport %q, which is neither of the two the proxy writes", conn.ClientPort, conn.Transport)
				}
				if conn.TransportNote != "" {
					t.Errorf("connection %d: the transport could not be decided from the bytes: %s", conn.ClientPort, conn.TransportNote)
				}
				if len(conn.Up)+len(conn.Down) == 0 {
					t.Errorf("connection %d: there are rows and not one message: the reader framed nothing (%s)",
						conn.ClientPort, strings.Join(conn.Notes, "; "))
					continue
				}
				if len(conn.Down) == 0 {
					t.Errorf("connection %d: the client sent %d messages and the server answered none, and a capture with no truth compares nothing",
						conn.ClientPort, len(conn.Up))
				}
				truthMessages += len(conn.Down)
			}
			if truthMessages == 0 {
				t.Errorf("the original said nothing in the whole capture")
			}
			t.Logf("truth=%s conns=%d original messages=%d", cap.Reference.File, len(cap.Conns), truthMessages)
		})
	}
}

// TestBothTransportsAreCoveredByACommittedCapture: the two framings the proxy
// writes are not two theories, they are two captures in testdata, and a reader
// that only understood one of them would be caught by the test above. Stated on
// its own so that losing one of the two cannot pass unnoticed: the bsPort capture
// is 2026-09-30_201122_c4e7c284ff1c, the first of this transport ever recorded.
func TestBothTransportsAreCoveredByACommittedCapture(t *testing.T) {
	seen := map[string]string{}
	for _, name := range committedCaptures(t) {
		cap, err := sim.Load(filepath.Join("testdata", name))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, conn := range cap.Conns {
			if conn.Transport == "" {
				continue
			}
			if _, ok := seen[conn.Transport]; !ok {
				seen[conn.Transport] = name
			}
		}
	}
	for _, want := range []string{sim.FramingWS, sim.FramingRLink} {
		if seen[want] == "" {
			t.Errorf("no committed capture goes over %s, and the reader has to read both", want)
		} else {
			t.Logf("%s: %s", want, seen[want])
		}
	}
}

// TestTheBsPortCaptureCarriesTheInnerProtocolTypes: on the bsPort the row IS the
// RLink message, so the types the reader has to recover are the protocol's own
// (0x15 LobbyConnect, 0x09 GameData, 0x0A acks) and not a frame head.
//
// The totals are the ones the capture's own msg_type column carries, read off the
// committed truth log (secondary.log, 150 rows): 0x09 x2, 0x0A x8, 0x0F x3, 0x13
// x113, 0x14 x20, 0x15 x4. They are a cross-check and not a restatement of the
// reader: the column is written by the proxy from the raw bytes of each row
// (cmd/proxy/packets.go:33), so the reader agreeing with it row by row means the
// two halves recovered the same messages out of the same file.
func TestTheBsPortCaptureCarriesTheInnerProtocolTypes(t *testing.T) {
	const name = "testdata/2026-09-30_201122_c4e7c284ff1c"
	if _, err := os.Stat(filepath.Join(name, "metadata.json")); err != nil {
		t.Skipf("the bsPort capture is not in this tree: %v", err)
	}
	cap, err := sim.Load(name)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cap.Conns) != 2 {
		t.Fatalf("this capture has two client connections, it has %d", len(cap.Conns))
	}
	total := map[byte]int{}
	connects := 0
	for _, conn := range cap.Conns {
		if conn.Transport != sim.FramingRLink {
			t.Errorf("connection %d: transport %q, and the game went over the bsPort", conn.ClientPort, conn.Transport)
		}
		for _, m := range conn.Up {
			total[m.Type]++
			if m.Type == protocol.MsgLobbyConnect {
				connects++
			}
		}
		for _, m := range conn.Down {
			total[m.Type]++
		}
	}
	if connects != 2 {
		t.Errorf("each connection sends its LobbyConnect, there are %d", connects)
	}
	want := map[byte]int{
		protocol.MsgGameData:      2,
		protocol.MsgEndOfGameData: 8,
		protocol.MsgPeer:          3,
		protocol.MsgLobbyRelay:    113,
		protocol.MsgLobbyPing:     20,
		protocol.MsgLobbyConnect:  4,
	}
	for typ, n := range want {
		if total[typ] != n {
			t.Errorf("0x%02X %s: the reader got %d and the truth's log has %d",
				typ, protocol.MessageTypeName(typ), total[typ], n)
		}
	}
	if len(total) != len(want) {
		t.Errorf("the reader got types the truth does not have: %s", renderTypes(total))
	}
	t.Logf("types read: %s", renderTypes(total))
}

// committedCaptures are the capture directories sitting in testdata, which is what
// the generated tests replay.
func committedCaptures(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join("testdata", e.Name(), "metadata.json")); err != nil {
			continue
		}
		out = append(out, e.Name())
	}
	sort.Strings(out)
	return out
}

func renderTypes(total map[byte]int) string {
	keys := make([]byte, 0, len(total))
	for k := range total {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, "0x"+strings.ToUpper(hexByte(k))+" x"+itoa(total[k]))
	}
	return strings.Join(parts, " ")
}

func hexByte(b byte) string {
	const digits = "0123456789abcdef"
	return string([]byte{digits[b>>4], digits[b&0x0F]})
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
