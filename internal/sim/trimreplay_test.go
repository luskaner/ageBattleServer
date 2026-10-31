package sim_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/sim"
)

// The end-to-end half of the crop proof: two real replays, the same rules, and the
// verdict compared.
//
// trim_test.go drives the crop and the comparison directly, which is where the
// falsability lives: a crop that emptied a stream, moved the observer's window or
// swallowed a must-match finding is caught there without a single socket. What that
// cannot show is that a crop of a REAL capture is one, because nothing about a
// hand-built report is the harness. So this file does the whole thing twice: it
// replays a capture, crops it, replays the crop, and asks sim.CheckTrim whether the
// two runs are the same run with a suffix missing.
//
// The fixture is the _26 shape (the synthetic one in testdata) spread over three
// seconds instead of one, so a crop has somewhere to cut and still leaves the
// observer a replay stream to be served: the crop point has to be after the
// handshake, or the check would be proving that an emptied stream is caught rather
// than that a shortened one is safe.
//
// What this test costs is two real-time replays, about six seconds in total, which
// is the price of a check that a crop cannot be quietly wrong. It is not cheap
// enough to run against the 24 minute capture, which is what cmd/trimsim -verify is
// for: same comparison, same rules, run by the tool that writes the crop and
// recorded in the crop's own metadata.

// trimAcks are the five replay pages this fixture uploads, and the 0x0A acks the
// session built for them. The shape is fixtureAcks': [0xfe][sub count] + subs +
// [pending count][pending slots] + [u16 LE len] block, with the sub being
// [slot][u16 LE len][payload].
//
// They are written out rather than taken from the code under test, for the reason
// every fixture in this package is written out: a fixture built from the
// implementation cannot fail when the implementation is wrong.
var trimPages = [][]byte{
	{0xAA, 0x01},
	{0xBB, 0x02, 0x03},
	{0xCC, 0x04},
	{0xDD, 0x05},
	{0xEE, 0x06},
	{0xFF, 0x07},
}

func trimAckFor(page int) []byte {
	p := trimPages[page]
	out := []byte{0xFE, 0x01, 0x01, byte(len(p)), byte(len(p) >> 8)}
	out = append(out, p...)
	return append(out, 0x00, 0x00, 0x00)
}

func trimUploadFor(page int) []byte {
	return (&protocol.GameDataMessage{InnerType: 0xFE, Payload: trimPages[page]}).Serialize()
}

// buildTrimLog writes one side's log over three seconds:
//
//	t=0  the player connects and uploads three pages
//	t=1  an observer connects, is admitted, and is served the retained window
//	t=2  two more pages, acked to the player and to the observer as they are built
//	t=3  one more page, which is what a crop to two seconds cuts
//
// The observer is served the session's own 0x0A stream in the order the store holds
// it, which is what replayPageOrder and replayContent read, and the last page is
// there so that a crop really does shorten the stream instead of only moving the
// end of the run.
func buildTrimLog(t *testing.T, path string, s sideFixture) {
	t.Helper()
	ack := func(page int, at time.Duration, client int) row {
		return row{client: client, server: portOf(s, client), at: at,
			data: rlinkFrame(protocol.MsgEndOfGameData, trimAckFor(page))}
	}
	var rows []row

	// t=0: the player.
	rows = append(rows, row{client: fixturePlayerPort, toServer: true, server: s.bsPort,
		data: rlinkFrame(protocol.MsgLobbyConnect, fixtureLobbyConnect())})
	for page := range 3 {
		rows = append(rows, row{client: fixturePlayerPort, toServer: true, server: s.bsPort,
			data: rlinkFrame(protocol.MsgGameData, trimUploadFor(page))})
	}
	rows = append(rows, row{client: fixturePlayerPort, server: s.bsPort,
		data: rlinkFrame(protocol.MsgLobbyConnect, fixtureLobbyConnect())})
	for page := range 3 {
		rows = append(rows, ack(page, 0, fixturePlayerPort))
	}

	// t=1: the observer, and the window it is put at.
	rows = append(rows,
		row{client: fixtureObserverPort, toServer: true, server: s.wsPort, at: time.Second, data: fixtureUpgradeRequest, splitInto: 2},
		row{client: fixtureObserverPort, server: s.wsPort, at: time.Second, data: fixtureUpgradeResponse, splitInto: 3},
		row{client: fixtureObserverPort, toServer: true, server: s.wsPort, at: time.Second,
			data: wsLoggedFrame(append([]byte{0x1C}, fixtureObserveRequest()...), true)},
		row{client: fixtureObserverPort, toServer: true, server: s.wsPort, at: time.Second,
			data: wsLoggedFrame([]byte{protocol.MsgActivateObserver}, true)},
		row{client: fixtureObserverPort, server: s.wsPort, at: time.Second,
			data: wsLoggedFrame([]byte{protocol.MsgStartObserveResp, s.status}, false)},
		row{client: fixtureObserverPort, server: s.wsPort, at: time.Second,
			data: wsLoggedFrame(append([]byte{protocol.MsgActivateObserverResp}, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88), false)},
	)
	// The store is a number of blocks and not a number of frames (SESSION-JOURNAL.md
	// 24, "the store is a number of BLOCKS"), so all three acks are still in it
	// when the observer arrives, and the original serves the whole retained window
	// in one burst. That is what puts the observer at the start of the match, which
	// is the _26 shape and the whole point of the fixture: a truth that started at
	// page 1 would be the _30 symptom written into the fixture by hand.
	for page := range 3 {
		rows = append(rows, row{client: fixtureObserverPort, server: s.wsPort, at: time.Second,
			data: wsLoggedFrame(append([]byte{protocol.MsgEndOfGameData}, trimAckFor(page)...), false)})
	}

	// t=2: two more pages, and the observer is served them as the store builds them.
	for _, page := range []int{3, 4} {
		rows = append(rows, row{client: fixturePlayerPort, toServer: true, server: s.bsPort, at: 2 * time.Second,
			data: rlinkFrame(protocol.MsgGameData, trimUploadFor(page))})
	}
	for _, page := range []int{3, 4} {
		rows = append(rows, ack(page, 2*time.Second, fixturePlayerPort))
		r := ack(page, 2*time.Second, fixtureObserverPort)
		r.data = wsLoggedFrame(append([]byte{protocol.MsgEndOfGameData}, trimAckFor(page)...), false)
		rows = append(rows, r)
	}

	// t=3: the page a crop to two seconds takes away.
	rows = append(rows,
		row{client: fixturePlayerPort, toServer: true, server: s.bsPort, at: 3 * time.Second,
			data: rlinkFrame(protocol.MsgGameData, trimUploadFor(5))},
		ack(5, 3*time.Second, fixturePlayerPort),
	)

	writeLog(t, path, rows)
}

func portOf(s sideFixture, client int) int {
	if client == fixtureObserverPort {
		return s.wsPort
	}
	return s.bsPort
}

// writeTimedCapture writes the three second capture the crop is cut from.
func writeTimedCapture(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "larga")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	primary := primaryFixture()
	primary.frames = nil // the frames are placed by hand, second by second
	secondary := secondaryFixture()
	secondary.frames = nil
	buildTrimLog(t, filepath.Join(dir, "main.log"), primary)
	buildTrimLog(t, filepath.Join(dir, "secondary.log"), secondary)
	writeMeta(t, dir, primary, secondary)
	return dir
}

// TestCroppedCaptureStillAnswersTheSame is the check this whole arrangement is for,
// run against a real replay. The crop is refused if anything about the two verdicts
// differs, and the report of the crop is logged so a reader can see the same numbers
// the check compared.
func TestCroppedCaptureStillAnswersTheSame(t *testing.T) {
	src := writeTimedCapture(t)
	full, err := sim.Check(src, sim.Options{})
	if err != nil {
		t.Fatalf("replaying the capture: %v", err)
	}
	if !full.OK() {
		t.Fatalf("the untrimmed fixture must agree with the original it records, got %v\n%s",
			full.MustMatch, full.FirstDivergence())
	}
	if !hasObserverFrames(full) {
		t.Fatalf("the fixture has to serve the observer a replay, or the crop would be proved on nothing:\n%s", full.Render())
	}

	crop := filepath.Join(t.TempDir(), "recortada")
	rec, err := sim.Trim(src, crop, sim.TrimOptions{
		Keep:   2 * time.Second,
		Source: "larga",
		Why:    "the observer is admitted at t=1s and served the retained window at once; t=3s is the last page of a fixture whose only purpose is to have a tail to cut",
	})
	if err != nil {
		t.Fatalf("crop: %v", err)
	}
	if rec.KeptSpan != "2s" {
		t.Fatalf("the crop must say how much it kept, got %q", rec.KeptSpan)
	}

	trimmed, err := sim.Check(crop, sim.Options{})
	if err != nil {
		t.Fatalf("replaying the crop: %v", err)
	}
	t.Log("\n" + trimmed.Render())

	check := sim.CheckTrim(full, trimmed)
	for _, l := range check.Lines {
		t.Logf("verified: %s", l)
	}
	if !check.OK() {
		for _, p := range check.Problems {
			t.Errorf("[%s] conn %d: %s", p.Rule, p.Conn, p.Detail)
		}
		t.Fatalf("a crop of the tail changed what the harness concludes, and it is not a crop of the same run")
	}

	// The crop is only worth anything if the comparison it still makes is a real
	// one: a shorter stream is a shorter comparison, an empty one is a pass that
	// means nothing (SESSION-JOURNAL.md 24).
	if !hasObserverFrames(trimmed) {
		t.Fatalf("after the crop the observer is served nothing, so the harness would be comparing two empty streams:\n%s", trimmed.Render())
	}
	if got := replayFrameCounts(trimmed, fixtureObserverPort); got == 0 {
		t.Fatal("the crop left the observer with no replay frames at all")
	}
}

// TestCropPastTheHandshakeIsRejected is the other end of the same claim. A crop that
// takes the observer with it leaves a report over the connections that are left,
// which looks clean and compares nothing; the check has to name it.
func TestCropPastTheHandshakeIsRejected(t *testing.T) {
	src := writeTimedCapture(t)
	full, err := sim.Check(src, sim.Options{})
	if err != nil {
		t.Fatalf("replaying the capture: %v", err)
	}
	// Half a second: the player is there and the observer is not.
	crop := filepath.Join(t.TempDir(), "recortada")
	if _, err := sim.Trim(src, crop, sim.TrimOptions{Keep: 500 * time.Millisecond, Source: "larga", Why: "test"}); err != nil {
		t.Fatalf("crop: %v", err)
	}
	trimmed, err := sim.Check(crop, sim.Options{})
	if err != nil {
		t.Fatalf("replaying the crop: %v", err)
	}
	c := sim.CheckTrim(full, trimmed)
	if c.OK() {
		t.Fatalf("a crop that took the observer away must be rejected, and it passed with:\n%s", strings.Join(c.Lines, "\n"))
	}
	found := false
	for _, p := range c.Problems {
		if strings.Contains(p.Detail, "is not in the trimmed one") || strings.Contains(p.Detail, "no game data") {
			found = true
		}
	}
	if !found {
		t.Fatalf("the problem must name the connection or the stream that went away, got %v", c.Problems)
	}
}

func hasObserverFrames(rep *sim.Report) bool { return replayFrameCounts(rep, fixtureObserverPort) > 0 }

// replayFrameCounts is how many 0x0A the replay was served on one connection,
// counted from the measure the trim check compares. The rules report it in prose
// inside a finding, and a test that had to parse prose to assert a number would be
// asserting the prose.
func replayFrameCounts(rep *sim.Report, clientPort int) int {
	for _, m := range rep.Measures {
		if m.ClientPort == clientPort {
			return m.TruthContent.Frames + m.OursContent.Frames
		}
	}
	return 0
}
