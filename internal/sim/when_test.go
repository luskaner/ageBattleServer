package sim

import (
	"strings"
	"testing"
	"time"

	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// This file tests the part of the report that answers a question, and it tests it
// from inside the package because the answer is built from the two streams' own
// timestamps: the whole point is to move a timestamp and watch the conclusion
// change. A test that only checks the report has the right shape would pass with
// every verdict inverted, and the verdict is the deliverable.

// origin is the capture's first message, which is what the truth's elapsed times
// are counted from. started is when the replay began feeding the input, a few
// milliseconds later, exactly as the two clocks differ in a real run.
var (
	truthZero = time.Date(2026, 9, 28, 21, 5, 8, 0, time.UTC)
	ourZero   = truthZero.Add(137 * time.Millisecond)
)

// at builds a message on one side's axis.
func at(typ byte, body string, d time.Duration, zero time.Time) *Msg {
	return &Msg{Type: typ, Body: []byte(body), Framing: FramingWS, At: zero.Add(d)}
}

func seconds(n int) []Elapsed {
	out := make([]Elapsed, 0, n)
	for i := range n {
		out = append(out, Elapsed{D: time.Duration(i+1) * time.Second, Valid: true})
	}
	return out
}

// TestJudgeTailSettlesOnly is the first of the two answers: the extra messages
// all arrive after the last input was fed, which is the harness still reading.
func TestJudgeTailSettlesOnly(t *testing.T) {
	w := Window{Start: 10 * time.Second, End: 10400 * time.Millisecond, Speed: 1}
	ours := seconds(4) // 1s..4s
	ours = append(ours, Elapsed{D: 10 * time.Second, Valid: true})
	theirs := seconds(4)

	v := judgeTail(ours, theirs, w)
	if !v.SettleOnly {
		t.Fatalf("one extra message, all of it at or after the window, must be settle only: %+v", v)
	}
	if !strings.Contains(v.Reason, "MAY BE AN ARTIFACT") {
		t.Errorf("the reason must say the difference may be an artifact of the harness, got: %s", v.Reason)
	}
	if strings.Contains(v.Reason, "INTERLEAVED") {
		t.Errorf("the two conclusions are mutually exclusive, got both: %s", v.Reason)
	}
	if v.Extra != 1 || v.ExtrasBeforeWindow != 0 {
		t.Errorf("extra=%d before=%d, want 1 and 0", v.Extra, v.ExtrasBeforeWindow)
	}
}

// TestJudgeTailInterleaved is the other answer, and the one the numbers have to
// be able to produce on their own: the extra arrived while the match was still
// being fed, so no settle window can explain it.
func TestJudgeTailInterleaved(t *testing.T) {
	w := Window{Start: 10 * time.Second, End: 10400 * time.Millisecond, Speed: 1}
	ours := seconds(5) // 1s..5s, the fifth one past the original's last at 4s
	theirs := seconds(4)

	v := judgeTail(ours, theirs, w)
	if v.SettleOnly {
		t.Fatalf("an extra message emitted before the window opened is not a settle artifact: %+v", v)
	}
	if !strings.Contains(v.Reason, "THE EXTRAS ARE INTERLEAVED THROUGH THE MATCH") {
		t.Errorf("the reason must say the extras are interleaved, got: %s", v.Reason)
	}
	if v.ExtrasBeforeWindow != 1 {
		t.Errorf("ExtrasBeforeWindow = %d, want 1", v.ExtrasBeforeWindow)
	}
}

// TestJudgeTailVerdictFlipsWithOneTimestamp is the falsifiability of the whole
// exercise: the same two streams, the same counts, the same window, and one
// message's time moved by 100ms across the window's edge. The conclusion has to
// change, and it has to change in the direction the move implies. A verdict that
// cannot flip is a verdict that is not being computed.
func TestJudgeTailVerdictFlipsWithOneTimestamp(t *testing.T) {
	w := Window{Start: 10 * time.Second, End: 10400 * time.Millisecond, Speed: 1}
	theirs := seconds(4)
	// The original sent 4 and we sent 5. The fifth of ours is the extra, and the
	// only thing that decides the question is when it was produced.
	build := func(fifth time.Duration) []Elapsed {
		return []Elapsed{
			{D: time.Second, Valid: true},
			{D: 2 * time.Second, Valid: true},
			{D: 3 * time.Second, Valid: true},
			{D: 4 * time.Second, Valid: true},
			{D: fifth, Valid: true},
		}
	}
	atEdge := judgeTail(build(10*time.Second), theirs, w)            // exactly on the edge: inside
	beforeEdge := judgeTail(build(9900*time.Millisecond), theirs, w) // 100ms before: outside

	if !atEdge.SettleOnly {
		t.Errorf("a message stamped exactly at the window's opening is at or after it, so it is a settle artifact: %+v", atEdge)
	}
	if beforeEdge.SettleOnly {
		t.Errorf("a message stamped 100ms before the window opened was emitted during the match, so it is interleaved: %+v", beforeEdge)
	}
	if atEdge.Reason == beforeEdge.Reason {
		t.Errorf("the two verdicts must not read the same:\n  %s\n  %s", atEdge.Reason, beforeEdge.Reason)
	}
	// The counts cannot have changed, only the conclusion about them.
	if atEdge.Extra != beforeEdge.Extra || atEdge.Ours != beforeEdge.Ours {
		t.Errorf("moving a timestamp must not change any count: %+v vs %+v", atEdge, beforeEdge)
	}
}

// TestJudgeTailArithmeticCloses: the split into "before the window" and "at or
// after it" has to account for exactly the number of extra messages, on both
// sides, or the conclusion rests on numbers that do not add up.
func TestJudgeTailArithmeticCloses(t *testing.T) {
	cases := []struct {
		name   string
		ours   int
		theirs int
		window time.Duration
		inEnd  int
	}{
		{"extras interleaved", 12, 9, 10 * time.Second, 0},
		{"extras in the window", 12, 5, 10 * time.Second, 7},
		{"we are short", 4, 9, 10 * time.Second, 0},
		{"one side empty", 0, 3, 10 * time.Second, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := Window{Start: c.window, End: c.window + 400*time.Millisecond, Speed: 1}
			var ours, theirs []Elapsed
			for i := range c.ours {
				d := time.Duration(i+1) * time.Second
				if i >= c.ours-c.inEnd {
					d = c.window + time.Duration(i)*time.Millisecond
				}
				ours = append(ours, Elapsed{D: d, Valid: true})
			}
			for i := range c.theirs {
				theirs = append(theirs, Elapsed{D: time.Duration(i+1) * time.Second, Valid: true})
			}
			v := judgeTail(ours, theirs, w)
			if got := v.InWindow + v.BeforeWindow; got != v.Ours {
				t.Errorf("our side: %d inside + %d before = %d, want the %d we sent", v.InWindow, v.BeforeWindow, got, v.Ours)
			}
			if got := v.TheirsBefore + v.TheirsAfter; got != v.Theirs {
				t.Errorf("the original's side: %d before + %d after = %d, want the %d it sent", v.TheirsBefore, v.TheirsAfter, got, v.Theirs)
			}
			after := v.InWindow - v.TheirsAfter
			if v.ExtrasBeforeWindow+after != v.Extra {
				t.Errorf("the split does not account for the extras: %d before + %d at or after = %d, want %d",
					v.ExtrasBeforeWindow, after, v.ExtrasBeforeWindow+after, v.Extra)
			}
			if v.Reason == "" {
				t.Error("every verdict has to carry a sentence: the report prints it verbatim")
			}
		})
	}
}

// TestJudgeTailUntimedIsNotZero: a message with no recorded time is a different
// fact from one at t=0, and it must not be counted on either side of the split.
func TestJudgeTailUntimedIsNotZero(t *testing.T) {
	w := Window{Start: 10 * time.Second, End: 10400 * time.Millisecond, Speed: 1}
	ours := []Elapsed{{D: time.Second, Valid: true}, {}}
	theirs := []Elapsed{{D: time.Second, Valid: true}, {}}

	v := judgeTail(ours, theirs, w)
	if v.InWindow != 0 {
		t.Errorf("a message with no time cannot be inside the window, got %d", v.InWindow)
	}
	if v.TheirsBefore != 1 {
		t.Errorf("only the timed message counts as before the window, got %d", v.TheirsBefore)
	}
	if e := (Elapsed{}).String(); e != "no time recorded" {
		t.Errorf("an absent time must read as absent and not as t=0s, got %q", e)
	}
	if e := (Elapsed{D: 26130 * time.Millisecond, Valid: true}).String(); e != "26.13s" {
		t.Errorf("Elapsed.String = %q, want 26.13s", e)
	}
}

// TestJudgeTailCompressedRunSaysSo: a run that fed the input faster than real
// time put the two series on different clocks, and the report has to refuse to
// compare them rather than print a number nobody can read.
func TestJudgeTailCompressedRunSaysSo(t *testing.T) {
	w := Window{Start: 10 * time.Second, End: 10400 * time.Millisecond, Speed: 2.97}
	v := judgeTail(seconds(6), seconds(4), w)
	if !v.AxisWarn {
		t.Fatal("a compressed run must set the axis warning")
	}
	if !strings.Contains(v.Reason, "x2.97") {
		t.Errorf("the reason must name the speed it was run at, got: %s", v.Reason)
	}
}

// TestSecondBucketsListsOnlyTheSecondsThatDiffer: a table of ninety identical
// numbers answers nothing and buries the two that decide the question.
func TestSecondBucketsListsOnlyTheSecondsThatDiffer(t *testing.T) {
	theirs := make([]Elapsed, 0, 8)
	ours := make([]Elapsed, 0, 8)
	for s := range 8 {
		theirs = append(theirs, Elapsed{D: time.Duration(s) * time.Second, Valid: true})
		ours = append(ours, Elapsed{D: time.Duration(s) * time.Second, Valid: true})
	}
	if got := secondBuckets(ours, theirs, maxSecondBuckets); got != "" {
		t.Fatalf("two identical streams must produce no table, got %q", got)
	}
	ours = append(ours, Elapsed{D: 26 * time.Second, Valid: true})
	ours = append(ours, Elapsed{D: 89 * time.Second, Valid: true})
	got := secondBuckets(ours, theirs, maxSecondBuckets)
	if got != "26s: 0|1 89s: 0|1" {
		t.Errorf("the table must hold only the seconds that differ, got %q", got)
	}
}

// TestSecondBucketsDropsTheMiddleAndSaysHowMany: with more differing seconds than
// fit, the ones that get dropped are counted out loud, because a truncated table
// that does not say so is a table that lies about its own coverage.
func TestSecondBucketsDropsTheMiddleAndSaysHowMany(t *testing.T) {
	var ours, theirs []Elapsed
	for s := range 20 {
		ours = append(ours, Elapsed{D: time.Duration(s) * time.Second, Valid: true})
	}
	got := secondBuckets(ours, theirs, 6)
	if !strings.HasPrefix(got, "(14 more seconds differ)") {
		t.Errorf("the dropped seconds must be reported, got %q", got)
	}
	if !strings.HasSuffix(got, "19s: 0|1") {
		t.Errorf("the last differing second must survive, got %q", got)
	}
}

// TestSecondBucketsIgnoresUntimed: a message with no time has no second.
func TestSecondBucketsIgnoresUntimed(t *testing.T) {
	ours := []Elapsed{{}, {D: 3 * time.Second, Valid: true}}
	theirs := []Elapsed{{D: 3 * time.Second, Valid: true}}
	if got := secondBuckets(ours, theirs, maxSecondBuckets); got != "" {
		t.Errorf("an untimed message cannot make a second differ, got %q", got)
	}
}

// TestPayloadDiffClean: when the payloads line up, the surplus IS a set of
// payloads the original did not send, and the report may say so with their times.
func TestPayloadDiffClean(t *testing.T) {
	ours := []*Msg{
		at(protocol.MsgEndOfGameData, "ack", time.Second, ourZero),
		at(protocol.MsgEndOfGameData, "ack", 2*time.Second, ourZero),
		at(protocol.MsgLobbyRelay, "relay", 3*time.Second, ourZero),
	}
	theirs := []*Msg{
		at(protocol.MsgEndOfGameData, "ack", time.Second, truthZero),
		at(protocol.MsgLobbyRelay, "relay", 3*time.Second, truthZero),
	}
	over, under := payloadDiff(ours, theirs, ourZero)
	if under != 0 {
		t.Fatalf("nothing was sent fewer times than the original did, got %d", under)
	}
	if total(over) != 1 {
		t.Fatalf("the surplus must be the one extra ack, got %d", total(over))
	}
	if len(over) != 1 || over[0].ours != 2 || over[0].theirs != 1 {
		t.Fatalf("unexpected surplus: %+v", over)
	}
	if !strings.Contains(over[0].String(), "t=1s, t=2s") {
		t.Errorf("the surplus must carry when we sent the payload, got %q", over[0].String())
	}
}

// TestPayloadDiffMixed: the real tick acks carry bytes that differ between two
// runs, and then the payload counts do not localise anything. The analysis has to
// notice that instead of quoting a surplus of 101 for an 8 message difference.
func TestPayloadDiffMixed(t *testing.T) {
	ours := []*Msg{
		at(protocol.MsgEndOfGameData, "ack-run-a", time.Second, ourZero),
		at(protocol.MsgEndOfGameData, "ack-run-b", 2*time.Second, ourZero),
	}
	theirs := []*Msg{
		at(protocol.MsgEndOfGameData, "ack-run-c", time.Second, truthZero),
		at(protocol.MsgEndOfGameData, "ack-run-d", 2*time.Second, truthZero),
	}
	over, under := payloadDiff(ours, theirs, ourZero)
	if total(over) != 2 || under != 2 {
		t.Fatalf("two payloads we sent that it did not and two it sent that we did not: over=%d under=%d", total(over), under)
	}
	if total(over) == len(ours)-len(theirs)+2*under {
		t.Error("the mixed case must not be mistaken for a clean surplus")
	}
}

// TestCompareConnReportsTheTimes is the end of the chain in one test: two
// streams, put on the two axes, through the real rules, and the report has to
// come out saying when each divergence happened and where the extras are.
func TestCompareConnReportsTheTimes(t *testing.T) {
	w := Window{Start: 10 * time.Second, End: 10400 * time.Millisecond, Speed: 1}
	rep := &Report{started: ourZero, origin: truthZero, window: w}

	// The original: four acks and two relays, all inside the match.
	theirs := &Conn{ClientPort: 5000, Transport: FramingWS}
	for s := 1; s <= 4; s++ {
		theirs.Down = append(theirs.Down, at(protocol.MsgEndOfGameData, "ack", time.Duration(s)*time.Second, truthZero))
	}
	theirs.Down = append(theirs.Down,
		at(protocol.MsgLobbyRelay, "relay", 5*time.Second, truthZero),
		at(protocol.MsgLobbyRelay, "relay", 6*time.Second, truthZero))

	// Ours: one extra ack during the match, one extra relay inside the window.
	got := &Conn{ClientPort: 5000, Transport: FramingWS}
	for s := 1; s <= 5; s++ {
		got.Down = append(got.Down, at(protocol.MsgEndOfGameData, "ack", time.Duration(s)*time.Second, ourZero))
	}
	got.Down = append(got.Down,
		at(protocol.MsgLobbyRelay, "relay", 5*time.Second, ourZero),
		at(protocol.MsgLobbyRelay, "relay", 6*time.Second, ourZero),
		at(protocol.MsgLobbyRelay, "relay", 10200*time.Millisecond, ourZero))

	compareConn(rep, theirs, got, &matchIndex{}, &matchContent{})
	out := rep.Render()

	for _, want := range []string{
		"the original sent 6 messages and we sent 8",
		"when: ours t=10.2s | original t=6s",
		"EndOfGameData: the original sent 4, we sent 5",
		"LobbyRelay: the original sent 2, we sent 3",
		"window [t=10s, t=10.4s]",
		"1 of the 2 message(s) we sent and it did not were emitted BEFORE the settle window opened",
		"THE EXTRAS ARE INTERLEAVED THROUGH THE MATCH",
		"the split is exact: 1 before the window plus 1 at or after it is the 2 extra",
		"by payload: we sent these 1 payload(s) more often than the original, which accounts for all 1 extra",
		"ours from t=1s to t=5s | original from t=1s to t=4s",
		"ours from t=5s to t=10.2s | original from t=5s to t=6s",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the report must contain %q\n---\n%s", want, out)
		}
	}
	// Both differing seconds, one per type, and no other: the acks differ in the
	// fifth second and the relays in the tenth.
	for _, want := range []string{"only the seconds that differ: 5s: 0|1", "only the seconds that differ: 10s: 0|1"} {
		if !strings.Contains(out, want) {
			t.Errorf("the report must contain %q\n---\n%s", want, out)
		}
	}
	if !rep.OK() {
		t.Error("none of this is a must-match divergence, so the report must still be OK")
	}
}

// TestMessageTailIsNotAMustMatch is a property, not a finding: whether the extras
// are a settle artifact depends on the harness's own settle, so the rule can
// never be promoted to fail a test. A promotion would be a decision about the two
// explanations, and this is not it.
func TestMessageTailIsNotAMustMatch(t *testing.T) {
	rep := &Report{started: ourZero, origin: truthZero, window: Window{Start: 10 * time.Second}}
	theirs := &Conn{ClientPort: 7, Transport: FramingWS, Down: []*Msg{
		at(protocol.MsgEndOfGameData, "ack", time.Second, truthZero),
	}}
	got := &Conn{ClientPort: 7, Transport: FramingWS, Down: []*Msg{
		at(protocol.MsgEndOfGameData, "ack", time.Second, ourZero),
		at(protocol.MsgEndOfGameData, "ack", 11*time.Second, ourZero),
	}}
	compareConn(rep, theirs, got, &matchIndex{}, &matchContent{})
	for _, f := range rep.MustMatch {
		if f.Rule == RuleMessageTail {
			t.Fatal("the message-tail rule is informative and must never fail a test")
		}
	}
	found := false
	for _, f := range rep.Informative {
		if f.Rule == RuleMessageTail {
			found = true
		}
	}
	if !found {
		t.Fatal("one extra message must be reported by the message-tail rule")
	}
}
