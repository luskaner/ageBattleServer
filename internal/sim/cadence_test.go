package sim

import (
	"strings"
	"testing"
	"time"
)

// The four acks in the twenty-sixth second.
//
// This file is the measurement that closed the question, and the test that keeps
// it closed. Both halves of it are cheap on purpose: it reads the capture's logs
// and does NOT replay it, so it runs in the fast path and costs about a
// hundredth of what TestSim_c2026_09_28_205830 costs.
//
// The finding it pins, and how it was reached:
//
//	2026-09-28_205830_314309565d6e, WebSocket, ORIGINAL at the front, so this
//	capture recorded the original AND our own binary on the same client and the
//	same wall clock. The harness replayed it and reported, per second:
//
//	    26s: original 3 | ours 7
//
//	Four more EndOfGameData in the one second where the ack stream starts, and
//	none in the other 62 seconds of the match. In the RECORDED run of our own
//	build, the same second holds 2, not 7: the two recorded sides hold 8 each in
//	every one of the seconds 27..88 and both send 500 acks in all. So the surplus
//	is not in the server.
//
//	The mechanism is the capture's clock and it is checkable without any argument:
//	the proxy stamps whole seconds (cmd/proxy/packets.go:47) and timeline() feeds
//	every message at exactly its stamp, so a message written anywhere inside
//	second N is fed at N.000. At one ack per 125 ms tick that is up to eight acks
//	of difference inside the first second of the stream and none afterwards, which
//	is precisely the shape. The phase of the tick against the input moved; the
//	rate did not.

// captureWithBothBinaries is the only capture among the five where the question
// can be asked at all in this form, and the only one the original sat in front of,
// so it is the reference for this file.
const bothBinariesCapture = "testdata/2026-09-28_205830_314309565d6e"

// TestTheCapturedRunDidNotTickFasterThanTheOriginal is the finding itself, read
// out of the capture instead of out of a replay.
//
// The assertions are the load-bearing ones and each of them would fail if the
// server's tick had changed:
//
//   - the run's own record of our binary sends the SAME TOTAL as the original
//     (500 each). A surplus would show here before it showed anywhere else.
//   - in the 26th second, the one where the replay reported "3 | 7", the recorded
//     run of our binary sent 2. This is the seven that is not.
//   - from the 27th to the 88th second both sides send exactly 8. That is the
//     cadence: one ack per 125 ms tick, on both, with no second of drift over a
//     minute.
func TestTheCapturedRunDidNotTickFasterThanTheOriginal(t *testing.T) {
	cap, err := Load(bothBinariesCapture)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cap.Other.Writer != WriterOurs {
		t.Fatalf("this capture has to hold the run of our own binary to answer the question, and the other side says writer=%q", cap.Other.Writer)
	}
	conn := acksConn(t, cap)
	if conn == nil {
		t.Fatal("the player connection was not read")
	}

	theirs := acksPerSecond(cap, conn, conn.Down)
	ours := acksPerSecond(cap, conn, conn.otherDown)
	if len(theirs) == 0 || len(ours) == 0 {
		t.Fatalf("no acks on either side: truth %d seconds, ours %d", len(theirs), len(ours))
	}
	first := firstAckSecond(conn.Down, cap.origin())
	if first != 26 {
		t.Fatalf("the first ack of the match is in second %d of this capture, and the whole question is about that second", first)
	}
	t.Logf("in the %ds: original %d, our own binary as the capture recorded it %d, the replay 7 "+
		"(that number is TestSim_c2026_09_28_205830, and it is the only one the replay has)", first, theirs[first], ours[first])

	if got, want := ours[first], 2; got != want {
		t.Errorf("in second %d our own binary sent %d acks as this capture recorded it, want %d: "+
			"if it had sent 7 the surplus would be the server's and not the replay's", first, got, want)
	}
	if got, want := theirs[first], 3; got != want {
		t.Errorf("the original sent %d acks in second %d, want %d", got, first, want)
	}

	// The cadence: 8 a second, on both sides, for a minute.
	for sec := first + 1; sec <= 88; sec++ {
		if theirs[sec] != 8 || ours[sec] != 8 {
			t.Fatalf("second %d: original %d acks, our recorded run %d, want 8 and 8. "+
				"A second that is not 8 on both sides is a rate difference, which is the thing this file says there is not",
				sec, theirs[sec], ours[sec])
		}
	}
	if got, want := sumSeconds(theirs), 500; got != want {
		t.Errorf("the original sent %d acks in all, want %d", got, want)
	}
	if got, want := sumSeconds(ours), 500; got != want {
		t.Errorf("our own binary sent %d acks in the same run, want %d: a surplus in the recorded run would be "+
			"the server's and would be visible here before it was visible in a replay", got, want)
	}
}

// TestTheReplayFeedsEveryMessageAtItsWholeSecondStamp is the mechanism, stated as
// a property of the code and checked against the capture rather than argued.
//
// timeline() exists because the capture's stamps are whole seconds
// (cmd/proxy/packets.go:47): it has no choice but to schedule each message at its
// stamp, and that is up to a second earlier than the write it was recorded from.
// If this test ever fails, the phase argument in cadence.go has lost its premise
// and the conclusion has to be re-derived rather than quoted.
func TestTheReplayFeedsEveryMessageAtItsWholeSecondStamp(t *testing.T) {
	cap, err := Load(bothBinariesCapture)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	sched, span, speed := timeline(cap.origin(), cap.Conns, Options{ReplaySpeed: 1})
	if speed != 1 {
		t.Fatalf("a replay with no wall-clock ceiling runs at x%v, and this test is about real time", speed)
	}
	if span <= 0 {
		t.Fatalf("the capture has no input span (%s), so there is nothing to schedule", span)
	}
	n := 0
	for port, list := range sched {
		for _, s := range list {
			n++
			if s.at%time.Second != 0 {
				t.Fatalf("client port %d: a message is scheduled at t=%s, which is not a whole second: "+
					"the input would be fed at a moment the capture cannot name, and the phase argument in cadence.go "+
					"would have no premise", port, s.at)
			}
		}
	}
	if n == 0 {
		t.Fatal("no input was scheduled at all")
	}
	t.Logf("%d captured client messages, every one scheduled at a whole second of a %s span", n, span)
}

// TestCadenceConclusionOnlyClaimsWhatTheNumbersShow drives the sentence the rule
// prints, with both shapes and with the one that has no claim.
//
// The strong shape needs three series at the same sustained rate; the middle shape
// needs a replay that is further from the run's own record of our binary than from
// the original; and the third series of cases below must conclude nothing. A
// conclusion that fired on any of the wrong numbers would turn the report's most
// reassuring line into the most dangerous one.
func TestCadenceConclusionOnlyClaimsWhatTheNumbersShow(t *testing.T) {
	steady := func(firstSecond, perSecond, seconds int) []Elapsed {
		var out []Elapsed
		for s := firstSecond; s < firstSecond+seconds; s++ {
			for i := range perSecond {
				out = append(out, Elapsed{D: time.Duration(s)*time.Second + time.Duration(i)*125*time.Millisecond, Valid: true})
			}
		}
		return out
	}
	// The shape of 2026-09-28_205830 second 26 and the 62 that follow it: a short
	// first second on each side, then 8 a second everywhere.
	partial := func(perSecond int) []Elapsed {
		var out []Elapsed
		for i := range perSecond {
			out = append(out, Elapsed{D: 26*time.Second + time.Duration(i)*125*time.Millisecond, Valid: true})
		}
		return append(out, steady(27, 8, 62)...)
	}

	t.Run("same rate on all three sides means the cadence is the same, and the phase is named only at the start", func(t *testing.T) {
		tbl := cadenceSeconds(cadenceSeries{elapsed: partial(3)}, cadenceSeries{elapsed: partial(7)}, cadenceSeries{elapsed: partial(2)})
		got := cadenceConclusion(tbl, partial(3), partial(7), partial(2))
		if !strings.Contains(got, "TICK CADENCE IS THE SAME ON EVERY SIDE") {
			t.Fatalf("three sides at the same sustained rate have the same cadence by definition, so the conclusion "+
				"has to say so. Got: %s", got)
		}
		if !strings.Contains(got, "The second that differs IS the second the ack stream starts in") {
			t.Errorf("the difference IS in the second the ack stream starts in, so the phase sentence is the one that "+
				"applies and it is missing: %s", got)
		}
		if !strings.Contains(got, "what differs there is the PHASE") {
			t.Errorf("the phase sentence has to say what moved: %s", got)
		}
	})

	t.Run("a difference in a later second is not the phase either", func(t *testing.T) {
		theirs := steady(26, 8, 62)
		ours := append([]Elapsed{}, theirs...)
		ours = append(ours, Elapsed{D: 80 * time.Second, Valid: true}, Elapsed{D: 80*time.Second + 125*time.Millisecond, Valid: true})
		recorded := steady(26, 8, 62)
		tbl := cadenceSeconds(cadenceSeries{elapsed: theirs}, cadenceSeries{elapsed: ours}, cadenceSeries{elapsed: recorded})
		got := cadenceConclusion(tbl, theirs, ours, recorded)
		if !strings.Contains(got, "TICK CADENCE IS THE SAME ON EVERY SIDE") {
			t.Errorf("the rates agree, so the cadence claim holds: %s", got)
		}
		if strings.Contains(got, "is the second the ack stream starts in") {
			t.Errorf("the second that differs is t=80s and the stream starts at t=26s, so the phase sentence must not "+
				"be used: %s", got)
		}
		if !strings.Contains(got, "no cause is claimed") {
			t.Errorf("a difference that is neither a rate nor a start-of-stream phase difference has to be reported "+
				"without a cause: %s", got)
		}
	})

	t.Run("a different rate is not a phase difference", func(t *testing.T) {
		theirs, ours, recorded := steady(26, 8, 62), steady(26, 4, 62), steady(26, 8, 62)
		tbl := cadenceSeconds(cadenceSeries{elapsed: theirs}, cadenceSeries{elapsed: ours}, cadenceSeries{elapsed: recorded})
		got := cadenceConclusion(tbl, theirs, ours, recorded)
		if strings.Contains(got, "TICK CADENCE IS THE SAME") {
			t.Fatalf("our replay sustains half the rate, so the cadence did move and the conclusion must not deny it. Got: %s", got)
		}
		if !strings.Contains(got, "Nothing is concluded") {
			t.Errorf("with the rates apart and the per-second gaps not pointing anywhere, the only honest sentence is "+
				"that nothing is concluded. Got: %s", got)
		}
	})

	t.Run("a replay further from its own recorded run than from the original", func(t *testing.T) {
		// The truth and our replay hold 8 a second and the run's own record of our
		// binary holds 4: in every one of those seconds the replay is 4 away from
		// the record and 0 away from the original, which is the one reading the
		// per-second gaps can support on their own.
		truth := append(steady(26, 8, 40), steady(66, 8, 20)...)
		ours := append([]Elapsed{}, truth...)
		recorded := append(steady(26, 4, 40), steady(66, 8, 20)...)
		tbl := cadenceSeconds(cadenceSeries{elapsed: truth}, cadenceSeries{elapsed: ours}, cadenceSeries{elapsed: recorded})
		got := cadenceConclusion(tbl, truth, ours, recorded)
		if strings.Contains(got, "TICK CADENCE IS THE SAME") {
			t.Fatalf("the recorded run sustains half the rate, so this must not claim a shared cadence. Got: %s", got)
		}
		if !strings.Contains(got, "belongs to the replay and not to the server") {
			t.Errorf("the per-second gaps point at the replay and the sentence has to say that much: %s", got)
		}
	})

	t.Run("no third series means no claim", func(t *testing.T) {
		rep := &Report{}
		want := &Conn{ClientPort: 1, Down: []*Msg{{Type: tickAck, At: time.Now()}}}
		got := &Conn{ClientPort: 1, Down: []*Msg{{Type: tickAck, At: time.Now()}}}
		tickCadence(rep, want, got)
		if len(rep.Informative) != 0 {
			t.Fatalf("with no recorded third series the rule has nothing to compare and must say nothing, got %d finding(s)",
				len(rep.Informative))
		}
	})
}

// steady is perSecond acks a second for seconds seconds in a row, which is what a
// tick at 125 ms looks like from outside.
func steady(firstSecond int, perSecond, seconds int) []Elapsed {
	var out []Elapsed
	for s := 0; s < seconds; s++ {
		for i := range perSecond {
			out = append(out, Elapsed{
				D:     time.Duration(firstSecond+s)*time.Second + time.Duration(i)*125*time.Millisecond,
				Valid: true,
			})
		}
	}
	return out
}

// TestCadenceNamesTheRecordedRunAndNotOnlyTheOriginal is the rule itself: a
// capture whose recorded sides disagree has to produce a finding that quotes all
// three series and concludes from the third.
//
// It is driven with three series built by hand so that the conclusion has to
// follow the numbers: swap the recorded series for one equal to the replay and the
// same finding has to stop claiming a phase difference.
func TestCadenceNamesTheRecordedRunAndNotOnlyTheOriginal(t *testing.T) {
	base := time.Now().Truncate(time.Hour)
	axis := func(start time.Duration, perSecond, seconds int) []*Msg {
		var out []*Msg
		for s := 0; s < seconds; s++ {
			for i := range perSecond {
				out = append(out, &Msg{
					Type: tickAck,
					At:   base.Add(start + time.Duration(s)*time.Second + time.Duration(i)*125*time.Millisecond),
				})
			}
		}
		return out
	}
	rep := &Report{origin: base, started: base, OtherFile: "secondary.log", OtherWriter: WriterOurs}
	want := &Conn{ClientPort: 7, Down: axis(26*time.Second, 8, 62)}
	want.otherDown = axis(26*time.Second, 8, 62)
	got := &Conn{ClientPort: 7, Down: axis(26*time.Second, 8, 62)}

	tickCadence(rep, want, got)
	if len(rep.Informative) != 0 {
		t.Fatalf("three identical series have nothing to report, got %d finding(s)", len(rep.Informative))
	}

	// Now make the replay's stream start early in the second, which is what the
	// whole-second stamp does, and the rule has to come out with a finding.
	firstSecond := 26 * time.Second
	partial := func(n int) []*Msg {
		var out []*Msg
		for i := range n {
			out = append(out, &Msg{Type: tickAck, At: base.Add(firstSecond + time.Duration(i)*125*time.Millisecond)})
		}
		return out
	}
	got.Down = append(partial(7), axis(firstSecond, 8, 62)...)
	rep.Informative = nil
	tickCadence(rep, want, got)
	if len(rep.Informative) != 1 {
		t.Fatalf("a stream that disagrees in one second has to produce exactly one finding, got %d", len(rep.Informative))
	}
	f := rep.Informative[0]
	if f.Rule != RuleTickCadence {
		t.Fatalf("rule = %q, want %q", f.Rule, RuleTickCadence)
	}
	joined := strings.Join(f.Lines, "\n")
	for _, want := range []string{
		"per second, original | our replay |",
		"secondary.log, writer=ours",
		"in all: the original sent 496, our replay 503",
		"TICK CADENCE IS THE SAME ON EVERY SIDE",
		"The second that differs IS the second the ack stream starts in",
		"informative in any case",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("the finding must contain %q\n---\n%s", want, joined)
		}
	}

	// The falsification of the whole rule: with the recorded run moved to a
	// different rate from both, the cadence claim has to go with it, or the rule
	// would print the most reassuring sentence in the report off two series.
	want.otherDown = axis(26*time.Second, 4, 62)
	rep.Informative = nil
	tickCadence(rep, want, got)
	if len(rep.Informative) != 1 {
		t.Fatalf("still one finding expected, got %d", len(rep.Informative))
	}
	moved := strings.Join(rep.Informative[0].Lines, "\n")
	if strings.Contains(moved, "TICK CADENCE IS THE SAME") {
		t.Errorf("the recorded run now sustains a different rate, so the shared-cadence claim has to disappear:\n%s", moved)
	}
	if !strings.Contains(moved, "informative in any case") {
		t.Error("the finding lost its own caveat")
	}
}

// acksConn is the connection that carries the tick acks: the one with the most of
// them, which on a capture with an observer is the player's.
func acksConn(t *testing.T, cap *Capture) *Conn {
	t.Helper()
	var best *Conn
	for _, c := range cap.Conns {
		n := len(filtered(c.Down, tickAck))
		if best == nil || n > len(filtered(best.Down, tickAck)) {
			best = c
		}
	}
	if best == nil {
		t.Fatal("the capture has no connection")
	}
	return best
}

// acksPerSecond buckets one side's tick acks by whole second on the capture's own
// clock, which is the only clock both recorded sides can be read at.
func acksPerSecond(cap *Capture, c *Conn, msgs []*Msg) map[int]int {
	out := map[int]int{}
	origin := cap.origin()
	for _, m := range filtered(msgs, tickAck) {
		e := elapsedOf(m, origin)
		if !e.Valid {
			continue
		}
		out[int(e.D/time.Second)]++
	}
	return out
}

func sumSeconds(m map[int]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}

func firstAckSecond(msgs []*Msg, origin time.Time) int {
	for _, m := range filtered(msgs, tickAck) {
		if e := elapsedOf(m, origin); e.Valid {
			return int(e.D / time.Second)
		}
	}
	return -1
}
