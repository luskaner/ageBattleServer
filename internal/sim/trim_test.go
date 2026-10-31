package sim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// The crop itself, and the proof that a crop is allowed to be believed.
//
// internal/sim/trim.go cuts a capture's wall clock; internal/sim/trimcheck.go is
// what decides whether the cut lost anything. The two are tested apart from any
// replay here, because the shapes they have to get right (which rows survive, which
// messages are a prefix, what an emptied stream looks like) are exactly the ones a
// replay would hide: a replay that agrees for the wrong reason looks identical to
// one that agrees for the right one. The end-to-end half, with two real replays, is
// trimreplay_test.go.

// stampAt is the fixture's clock: whole seconds, because the proxy stamps whole
// seconds (cmd/proxy/packets.go:47) and a fixture with finer rows would be testing
// a precision the capture format does not have.
var stampAt = []struct {
	at   int // seconds from the first row
	down []byte
	typ  byte
}{
	{at: 0, typ: 0x0A, down: ack([]byte{0xAA})},
	{at: 1, typ: 0x0A, down: ack([]byte{0xBB})},
	{at: 2, typ: 0x0A, down: ack([]byte{0xCC})},
	{at: 3, typ: 0x0A, down: ack([]byte{0xCC, 0xDD})},
	{at: 4, typ: 0x0A, down: ack([]byte{0xEE})},
}

// timedConn is a connection whose truth stream has one 0x0A per second for five
// seconds, and one observer served the first three of them.
func timedConn() (truth, ours *Conn) {
	truth = &Conn{ClientPort: 5001, Transport: FramingWS}
	truth.Up = []*Msg{{Type: 0x15, Body: []byte{0x01}}, {Type: 0x1C, Body: make([]byte, 24)}}
	for _, s := range stampAt {
		truth.Down = append(truth.Down, ackMsg(s.down))
	}
	ours = &Conn{ClientPort: 5001, Transport: FramingWS}
	ours.Up = append([]*Msg{}, truth.Up...)
	// Ours groups the first two acks together and serves the rest one by one, which
	// is the difference that replay-page-order cannot see and replay-content can.
	ours.Down = []*Msg{
		ackMsg(ack([]byte{0xAA}, []byte{0xBB})),
		ackMsg(ack([]byte{0xCC})),
		ackMsg(ack([]byte{0xCC, 0xDD})),
		ackMsg(ack([]byte{0xEE})),
	}
	return truth, ours
}

func measurePair(t *testing.T, truth, ours *Conn) ConnMeasure {
	t.Helper()
	content := &matchContent{
		theirs: readAcks(replayStream(truth.Down)).subs,
		ours:   readAcks(replayStream(ours.Down)).subs,
	}
	return measureConn(truth, ours, content)
}

func reportOf(measures ...ConnMeasure) *Report {
	return &Report{Measures: measures}
}

func problemsOf(c TrimCheck, rule string) []string {
	var out []string
	for _, p := range c.Problems {
		if p.Rule == rule {
			out = append(out, p.Detail)
		}
	}
	return out
}

// TestCheckTrimPassesOnACropOfTheTail is the positive case, and it is the whole
// claim: keep the first three seconds of a five second run and every property the
// verdict rests on is still there, with a suffix missing.
func TestCheckTrimPassesOnACropOfTheTail(t *testing.T) {
	truth, ours := timedConn()
	full := reportOf(measurePair(t, truth, ours))

	// The crop: the truth and our own stream both cut after the third 0x0A.
	truthCut, oursCut := timedConn()
	truthCut.Down = truthCut.Down[:3]
	oursCut.Down = oursCut.Down[:2]
	trimmed := reportOf(measurePair(t, truthCut, oursCut))

	c := CheckTrim(full, trimmed)
	if !c.OK() {
		t.Fatalf("a crop of the tail lost something the verdict rests on: %v", c.Problems)
	}
	if len(c.Lines) == 0 {
		t.Error("the check must record what it established, or a reader cannot tell an empty comparison from a passing one")
	}
}

// TestCheckTrimOnAnUncroppedCapture is the control for the check itself: a capture
// compared against itself has to pass. A checker that reported problems here would
// be measuring its own bookkeeping, and one that did not would prove nothing about
// a real crop.
func TestCheckTrimOnAnUncroppedCapture(t *testing.T) {
	truth, ours := timedConn()
	rep := reportOf(measurePair(t, truth, ours))
	if c := CheckTrim(rep, rep); !c.OK() {
		t.Fatalf("a capture compared against itself must pass, got %v", c.Problems)
	}
}

// TestCheckTrimCatchesAMessageTakenFromTheMiddle is the property a count cannot
// have and a digest can: a crop that removed one message from the middle leaves the
// same totals as one that removed it from the end, and it is a different run.
func TestCheckTrimCatchesAMessageTakenFromTheMiddle(t *testing.T) {
	truth, ours := timedConn()
	full := reportOf(measurePair(t, truth, ours))

	cut, oursCut := timedConn()
	cut.Down = append(cut.Down[:1], cut.Down[2:]...) // drop the second second
	trimmed := reportOf(measurePair(t, cut, oursCut))

	c := CheckTrim(full, trimmed)
	if len(problemsOf(c, TrimRuleTruthOutput)) == 0 {
		t.Fatalf("a message taken out of the middle of the truth's stream must be caught, got %v", c.Problems)
	}
}

// TestCheckTrimCatchesAnAlteredInput: a crop that changed what the game said is
// replaying a different match, and it can still agree with the original on the
// observer by luck.
func TestCheckTrimCatchesAnAlteredInput(t *testing.T) {
	truth, ours := timedConn()
	full := reportOf(measurePair(t, truth, ours))

	cut, oursCut := timedConn()
	cut.Down = cut.Down[:3]
	cut.Up = append([]*Msg{}, cut.Up...)
	cut.Up[1] = &Msg{Type: 0x1C, Body: make([]byte, 25)} // one byte different
	trimmed := reportOf(measurePair(t, cut, oursCut))

	c := CheckTrim(full, trimmed)
	if len(problemsOf(c, TrimRuleTruthInput)) == 0 {
		t.Fatalf("an altered client->server message must be caught, got %v", c.Problems)
	}
}

// TestCheckTrimCatchesAnEmptyStream is the trap of SESSION-JOURNAL.md 24, and the
// one this whole arrangement exists to make impossible. A crop deep enough to leave
// the observer with no replay frames makes replay-page-order and replay-content
// return without claiming anything, and the report then shows a clean pass over two
// empty things. The check has to say so instead of passing.
func TestCheckTrimCatchesAnEmptyStream(t *testing.T) {
	for name, empty := range map[string]func(*Conn){
		"ours":   func(c *Conn) { c.Down = nil },
		"theirs": func(c *Conn) { c.Down = nil },
		"both":   func(c *Conn) { c.Down = nil },
		// The shape section 24 is about: our side is still there and still says
		// something, it just has no replay in it any more, so both stream totals
		// are non-zero and only the CONTENT check can see it.
		"ours keeps the handshake but no replay": func(c *Conn) {
			c.Down = []*Msg{{Type: protocol.MsgLobbyConnect, Body: []byte{0x01}}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			truth, ours := timedConn()
			full := reportOf(measurePair(t, truth, ours))
			cut, oursCut := timedConn()
			cut.Down = cut.Down[:3]
			oursCut.Down = oursCut.Down[:2]
			empty(oursCut)
			c := CheckTrim(full, reportOf(measurePair(t, cut, oursCut)))
			if c.OK() {
				t.Fatalf("a crop that emptied the stream must fail loudly, and it passed with %v", c.Lines)
			}
			if len(problemsOf(c, TrimRuleCoverage))+len(problemsOf(c, TrimRuleContent)) == 0 {
				t.Fatalf("the emptiness must be named as coverage or content, got %v", c.Problems)
			}
		})
	}
}

// TestCheckTrimCatchesAMovedWindow is the _30 symptom in a different dress: an
// observer put past the start of the match has no start-of-match data, and no suffix
// of a stream can move where it starts, so a crop that moved it is not a crop.
func TestCheckTrimCatchesAMovedWindow(t *testing.T) {
	truth, ours := timedConn()
	full := reportOf(measurePair(t, truth, ours))

	cut, oursCut := timedConn()
	cut.Down = cut.Down[:3]
	// The window now starts on the second sub of the match instead of the first.
	cut.Down[0] = ackMsg(ack([]byte{0xBB}))
	oursCut.Down = oursCut.Down[:1]
	oursCut.Down[0] = ackMsg(ack([]byte{0xBB}))
	trimmed := reportOf(measurePair(t, cut, oursCut))

	c := CheckTrim(full, trimmed)
	if len(problemsOf(c, TrimRuleContent)) == 0 {
		t.Fatalf("a window that starts elsewhere must be caught, got %v", c.Problems)
	}
}

// TestCheckTrimCatchesADroppedSubAndADroppedEntry: the two sequences
// replay-content compares, on the truth side. A crop cannot shorten one from the
// front, so anything but a suffix is a different run.
func TestCheckTrimCatchesADroppedSubAndADroppedEntry(t *testing.T) {
	for name, mutate := range map[string]func(*Conn){
		"a sub swapped for another": func(c *Conn) {
			c.Down[1] = ackMsg(ack([]byte{0x11, 0x22}))
		},
		"a checksum entry swapped for another": func(c *Conn) {
			c.Down[2] = ackMsg(ackWithEntry([]byte{0xCC}, 0x99999999, 3))
		},
	} {
		t.Run(name, func(t *testing.T) {
			truth, ours := timedConn()
			// The truth carries one entry, so the entry sequence is not empty and
			// the rule has something to compare.
			truth.Down[2] = ackMsg(ackWithEntry([]byte{0xCC}, 0x11111111, 1))
			ours.Down[1] = ackMsg(ackWithEntry([]byte{0xCC}, 0x11111111, 1))
			full := reportOf(measurePair(t, truth, ours))

			cut, oursCut := timedConn()
			cut.Down = append([]*Msg{}, cut.Down[:3]...)
			oursCut.Down = append([]*Msg{}, oursCut.Down[:2]...)
			mutate(cut)
			c := CheckTrim(full, reportOf(measurePair(t, cut, oursCut)))
			if len(problemsOf(c, TrimRuleContent)) == 0 {
				t.Fatalf("a changed sub-message or checksum entry must be caught, got %v", c.Problems)
			}
		})
	}
}

// TestCheckTrimCatchesAChangedVerdict, both ways. A crop that hides a must-match
// finding is the worst outcome there is, because the suite goes green over a
// server that is still wrong; a crop that invents one is the suite going red over
// nothing.
func TestCheckTrimCatchesAChangedVerdict(t *testing.T) {
	truth, ours := timedConn()
	full := reportOf(measurePair(t, truth, ours))
	full.MustMatch = []Finding{{Rule: RuleStartObserveStatus, ClientPort: 5001}}

	cut, oursCut := timedConn()
	cut.Down, oursCut.Down = cut.Down[:3], oursCut.Down[:2]
	trimmed := reportOf(measurePair(t, cut, oursCut))
	if c := CheckTrim(full, trimmed); len(problemsOf(c, TrimRuleVerdict)) == 0 {
		t.Fatalf("a crop that swallowed a must-match finding must be rejected, got %v", c.Problems)
	}

	trimmed.MustMatch = []Finding{{Rule: RuleActivateObserverResponse, ClientPort: 5001}}
	if c := CheckTrim(full, trimmed); len(problemsOf(c, TrimRuleVerdict)) == 0 {
		t.Fatalf("a crop that invented a must-match finding must be rejected, got %v", c.Problems)
	}
}

// TestCheckTrimCatchesADroppedConnection: a crop can remove a whole connection, and
// a report over the connections that are left looks perfectly clean.
func TestCheckTrimCatchesADroppedConnection(t *testing.T) {
	truth, ours := timedConn()
	full := reportOf(measurePair(t, truth, ours))

	// A second connection that the crop has removed entirely.
	gone := &Conn{ClientPort: 5002, Transport: FramingWS}
	gone.Up = []*Msg{{Type: 0x1C, Body: make([]byte, 24)}}
	gone.Down = []*Msg{ackMsg(ack([]byte{0xAA}))}
	full.Measures = append(full.Measures, measurePair(t, gone, gone))

	cut, oursCut := timedConn()
	cut.Down, oursCut.Down = cut.Down[:3], oursCut.Down[:2]
	if c := CheckTrim(full, reportOf(measurePair(t, cut, oursCut))); len(problemsOf(c, TrimRuleCoverage)) == 0 {
		t.Fatalf("a connection the crop removed must be caught, got %v", c.Problems)
	}
}

// TestCheckTrimCatchesAnAddedConnection is the mirror: traffic the run never
// recorded cannot appear in a crop of it.
func TestCheckTrimCatchesAnAddedConnection(t *testing.T) {
	truth, ours := timedConn()
	cut, oursCut := timedConn()
	cut.Down, oursCut.Down = cut.Down[:3], oursCut.Down[:2]
	trimmed := reportOf(measurePair(t, cut, oursCut))
	// A connection the untrimmed capture never had, which a crop cannot invent.
	extra := &Conn{ClientPort: 5003, Transport: FramingWS}
	extra.Down = []*Msg{ackMsg(ack([]byte{0xAA}))}
	trimmed.Measures = append(trimmed.Measures, measurePair(t, extra, extra))

	if c := CheckTrim(reportOf(measurePair(t, truth, ours)), trimmed); c.OK() {
		t.Fatal("a connection the untrimmed capture never had must be caught")
	}
}

// TestCheckTrimCatchesMoreTrafficThanTheRunHad: a crop cannot make the server talk
// more. If the two runs disagree in that direction the two reports are not two runs
// of the same thing, and comparing them means nothing.
func TestCheckTrimCatchesMoreTrafficThanTheRunHad(t *testing.T) {
	truth, ours := timedConn()
	cut, oursCut := timedConn()
	cut.Down, oursCut.Down = cut.Down[:3], oursCut.Down[:2]
	if c := CheckTrim(reportOf(measurePair(t, cut, oursCut)), reportOf(measurePair(t, truth, ours))); len(problemsOf(c, TrimRuleTruthOutput)) == 0 {
		t.Fatalf("a run with more traffic than the capture must be caught, got %v", c.Problems)
	}
}

// TestCheckTrimRecordsTheVerdictItFound: the evidence travels with the data, so a
// reader of a trimmed capture can see what was established and not have to take the
// crop on trust.
func TestCheckTrimRecordsTheVerdictItFound(t *testing.T) {
	truth, ours := timedConn()
	c := CheckTrim(reportOf(measurePair(t, truth, ours)), reportOf(measurePair(t, truth, ours)))
	joined := strings.Join(c.Lines, "\n")
	for _, want := range []string{"verdict:", "the client's own messages", "the original's answers", "the observer's replay:"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the evidence must mention %q, got:\n%s", want, joined)
		}
	}
}

// # The crop on disk

// cropFixture writes a capture with rows spread over six seconds, so a crop has a
// window to cut.
func cropFixture(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "capture")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{mainLogName, secondaryLogName} {
		body := "timestamp,src_host,src_port,dest_host,dest_port,msg_type,type_name,body_len,body_hex,data_b64\n"
		for s := range 6 {
			body += "2026-09-30T01:05:3" + string(rune('0'+s)) + "Z,127.0.0.1,15001,127.0.0.1,51001,0x0A,EndOfGameData,3,0a0000,AAAAAAAAAAA=\n"
			body += "2026-09-30T01:05:3" + string(rune('0'+s)) + "Z,127.0.0.1,51001,127.0.0.1,15001,0x09,GameData,2,0900,AQID\n"
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	meta := `{"format_version":1,"started_at":"2026-09-30 01:05:30","commit":"abc1234","dirty":"false",
	  "primary":{"writer":"ours","exe_path":"ours.exe","sha256":"11"},
	  "secondary":{"writer":"original","exe_path":"orig.exe","sha256":"22"},
	  "primary_args":["-bsPort","14001"],"secondary_args":["-bsPort","15001"]}`
	if err := os.WriteFile(filepath.Join(dir, metadataName), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "secondary.meta.json"), []byte(`{"file":"secondary.log","writer":"original"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "COMMIT.txt"), []byte("abc1234\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestTrimKeepsTheWindowAndSaysSo: the rows that survive are the rows inside the
// window, the rows after it are gone, the timestamps are the ones the run recorded
// (so a moment in a trimmed report means the moment it meant before), and the
// record says what was cut.
func TestTrimKeepsTheWindowAndSaysSo(t *testing.T) {
	src := cropFixture(t)
	dst := filepath.Join(t.TempDir(), "crop")
	rec, err := Trim(src, dst, TrimOptions{Keep: 3 * time.Second, Source: "capture", Why: "the test"})
	if err != nil {
		t.Fatalf("Trim: %v", err)
	}
	if rec.KeptSpan != "3s" {
		t.Errorf("the record must say how much was kept, got %q", rec.KeptSpan)
	}
	if rec.Dropped != "" && !strings.Contains(rec.Dropped, "of 24 rows") {
		t.Errorf("the record must say how much was dropped, got %q", rec.Dropped)
	}

	for _, name := range []string{mainLogName, secondaryLogName} {
		rows, err := readFixtureRows(filepath.Join(dst, name))
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != 6 {
			t.Fatalf("%s: a three second window of a six second capture keeps three seconds of rows, got %d", name, len(rows))
		}
		for _, ts := range rows {
			if !strings.HasPrefix(ts, "2026-09-30T01:05:3") {
				t.Errorf("%s: the timestamps must not be rewritten, got %q", name, ts)
			}
			if ts >= "2026-09-30T01:05:33Z" {
				t.Errorf("%s: kept a row outside the window: %q", name, ts)
			}
		}
	}
	// The sidecars and the run's own record travel with the crop, or a cropped
	// capture could not be traced back to the run that produced it.
	for _, name := range []string{"secondary.meta.json", "COMMIT.txt"} {
		if _, err := os.Stat(filepath.Join(dst, name)); err != nil {
			t.Errorf("%s must be copied into the crop: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dst, "main.meta.json")); !os.IsNotExist(err) {
		t.Error("a sidecar the source did not have must not be invented")
	}
}

// TestTrimRefusesTheThreeThingsThatWouldDestroyARecordOfARun.
func TestTrimRefusesTheThreeThingsThatWouldDestroyARecordOfARun(t *testing.T) {
	src := cropFixture(t)
	t.Run("overwrites a capture", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "crop")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "main.log"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Trim(src, dir, TrimOptions{Keep: time.Second, Why: "t"}); err == nil {
			t.Fatal("a crop over a non-empty directory must be refused")
		}
	})
	t.Run("crops a crop", func(t *testing.T) {
		first := filepath.Join(t.TempDir(), "one")
		if _, err := Trim(src, first, TrimOptions{Keep: 4 * time.Second, Why: "t"}); err != nil {
			t.Fatal(err)
		}
		second := filepath.Join(t.TempDir(), "two")
		_, err := Trim(first, second, TrimOptions{Keep: 2 * time.Second, Why: "t"})
		if err == nil || !strings.Contains(err.Error(), "already trimmed") {
			t.Fatalf("cropping a crop must be refused, got %v", err)
		}
	})
	t.Run("keeps nothing", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "crop")
		if _, err := Trim(src, dir, TrimOptions{Keep: 0, Why: "t"}); err == nil {
			t.Fatal("a crop of nothing must be refused")
		}
		if _, err := Trim(src, dir, TrimOptions{Keep: -time.Second, Why: "t"}); err == nil {
			t.Fatal("a negative window must be refused")
		}
	})
}

// TestLoadRefusesATrimmedCaptureWithNoRecord: format_version 2 says the capture is a
// crop, and a crop that does not say what it cut is a capture nobody can audit.
func TestLoadRefusesATrimmedCaptureWithNoRecord(t *testing.T) {
	src := cropFixture(t)
	meta := `{"format_version":2,"started_at":"2026-09-30 01:05:30","commit":"abc1234","dirty":"false",
	  "primary":{"writer":"ours","exe_path":"ours.exe","sha256":"11"},
	  "secondary":{"writer":"original","exe_path":"orig.exe","sha256":"22"},
	  "primary_args":["-bsPort","14001"],"secondary_args":["-bsPort","15001"]}`
	if err := os.WriteFile(filepath.Join(src, metadataName), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(src); err == nil || !strings.Contains(err.Error(), "trim record") {
		t.Fatalf("format_version 2 without a trim record must be refused, got %v", err)
	}
}

// TestReportSaysWhenTheCaptureIsATrimOfALongerRun: the reader of a failing test has
// to know that the capture under it is three minutes of a twenty-five minute run.
// A report that hid that would let a crop be believed because the test was green.
func TestReportSaysWhenTheCaptureIsATrimOfALongerRun(t *testing.T) {
	rep := &Report{
		Capture: "capture",
		Trim: &TrimRecord{
			Tool: "cmd/trimsim", Source: "larga", KeptFrom: "2026-09-30T01:05:31Z", KeptTo: "2026-09-30T01:09:31Z",
			KeptSpan: "4m0s", Dropped: "10m0s of 14m0s", Why: "the observer is activated at 4m2s and nothing else happens after it",
			Verified: []string{"verdict: no must-match divergence, and the same findings on the same connections"},
		},
	}
	out := rep.Render()
	for _, want := range []string{"TRIMMED", "4m0s", "10m0s of 14m0s", "verified:", "real time"} {
		if !strings.Contains(out, want) {
			t.Errorf("the report must say %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "NOT VERIFIED") {
		t.Error("a crop with evidence must not be rendered as unverified")
	}

	rep.Trim.Verified = nil
	if out := rep.Render(); !strings.Contains(out, "NOT VERIFIED") {
		t.Error("a crop with no evidence must read as unchecked, because 'no evidence' and 'checked and fine' are different facts")
	}
}

// TestDigestSeparatesMessages: the whole prefix argument rests on two different
// messages having two different digests.
func TestDigestSeparatesMessages(t *testing.T) {
	a := digestOf(ack([]byte{0xAA}))
	b := digestOf(ack([]byte{0xBB}))
	if a == b {
		t.Fatal("two different sub-messages must not digest the same")
	}
	if a != digestOf(ack([]byte{0xAA})) {
		t.Fatal("the digest must be stable for the same bytes")
	}
	if got := len(measureStream([]*Msg{ackMsg(ack([]byte{0xAA})), ackMsg(ack([]byte{0xBB}))}).Digests); got != 2 {
		t.Fatalf("one digest per message, got %d", got)
	}
}

func readFixtureRows(path string) ([]string, error) {
	rows, err := newRowReader(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for {
		rec, err := rows.Next()
		if err != nil {
			break
		}
		out = append(out, rec.fields[0])
	}
	return out, nil
}
