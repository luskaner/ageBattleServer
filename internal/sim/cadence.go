package sim

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// The tick cadence, on three series instead of two.
//
// The open question this file exists to answer is the one the counts raised and
// could not settle: the server emits one 0x0A EndOfGameData per simulation tick,
// so a per-second difference between what we send
// and what the reference server sent is either a different cadence or a
// different PHASE,
// and a total cannot tell them apart. Two captures pinned it down:
//
//	2026-09-28_205830_314309565d6e   26s: original 3 | ours 7
//
// Four more acks than the reference server in the one second where the ack
// stream starts,
// and none anywhere else in a 1m29s match. That reads as a real cadence defect
// and it is not one, and the reason is in the capture itself rather than in an
// argument: the capture recorded BOTH binaries, on the same client and the same
// wall clock, so the run of our own server is sitting in the other log of the very
// same capture. Measured there, second 26 holds 3 acks from the reference server
// and 2
// from our binary, and the following 63 seconds hold 8 each from both, with the
// same 500 acks in total. Our replay sent 7 in that second.
//
// So the difference is produced by the replay, and the reason is the capture's
// clock. The proxy stamps whole seconds (cmd/proxy/packets.go:47) and timeline()
// feeds every message at exactly its stamp, so a message written anywhere inside
// second N is fed at N.000: up to a second early. At one ack per 125 ms that is
// up to eight acks of difference inside the first second of the stream and none
// afterwards, which is exactly the shape the capture measures. It is the phase of
// the tick against the input, not the rate of the tick.
//
// The rule prints that comparison whenever the three series do not agree, so the
// next capture that shows "N | M" in a second says which of the two it was rather
// than leaving the reader to guess. It stays informative: it compares a message
// count, which AGENTS.md rules informative whatever it finds, and it concludes
// nothing about the server on its own.

// RuleTickCadence is the three-series per-second rate of the tick ack.
const RuleTickCadence = "tick-cadence"

// tickAck is the type whose rate is the tick's own: the reference server builds
// one per simulation tick and sends it to every peer, so a
// difference in how many of them there are in a second is a difference in the
// tick and in nothing else.
var tickAck byte = protocol.MsgEndOfGameData

// maxCadenceSeconds bounds the three-way table the same way secondBuckets bounds
// the two-way one, and for the same reason: a run of identical numbers answers
// nothing and buries the seconds that decide.
const maxCadenceSeconds = 12

// cadenceSeries is one side's acks on the shared axis, plus what it is.
type cadenceSeries struct {
	what    string
	elapsed []Elapsed
}

// tickCadence compares the per-second rate of the tick ack on three series: what
// the original sent, what this build's replay sent, and what the capture recorded
// the other side of the very same run sending.
//
// It emits nothing when the three agree second by second, and when there is no
// recorded third series (a capture whose other side could not be read) the rule
// says nothing either rather than guessing with two series.
//
// The finding is a report on the harness as much as on the server, and that is
// the point: "we sent four more acks in the 26th second" is only a fidelity
// finding if our own binary sent four more acks in the 26th second of that same
// run, and the only place that is written down is the other log of the capture.
func tickCadence(rep *Report, want, got *Conn) {
	if len(want.otherDown) == 0 || len(want.Down) == 0 || len(got.Down) == 0 {
		return
	}
	theirs := onAxis(filtered(want.Down, tickAck), rep.origin)
	ours := onAxis(filtered(got.Down, tickAck), rep.started)
	recorded := onAxis(filtered(want.otherDown, tickAck), rep.origin)
	if len(theirs) == 0 || len(ours) == 0 || len(recorded) == 0 {
		return
	}
	table := cadenceSeconds(cadenceSeries{
		what:    "original",
		elapsed: theirs,
	}, cadenceSeries{
		what:    "our replay",
		elapsed: ours,
	}, cadenceSeries{
		what:    "other side of this capture",
		elapsed: recorded,
	})
	if len(table.rows) == 0 {
		return
	}

	var lines []string
	lines = append(lines, fmt.Sprintf(
		"per second, original | our replay | the run's own record of the other binary (the %s, writer=%s, which is the commit "+
			"the capture was made with and not necessarily this build)",
		rep.OtherFile, rep.OtherWriter))
	for _, r := range table.rows {
		lines = append(lines, fmt.Sprintf("  t=%ds  %4d | %4d | %4d%s", r.second, r.theirs, r.ours, r.recorded, r.note))
	}
	if table.dropped > 0 {
		lines = append(lines, fmt.Sprintf("  (%d more seconds differ)", table.dropped))
	}
	lines = append(lines, fmt.Sprintf(
		"in all: the original sent %d, our replay %d, and the run's own record of our binary %d",
		countValid(theirs), countValid(ours), countValid(recorded)))
	lines = append(lines, "sustained per-second rate, over every second after the first one the ack stream occupies: "+renderRates(theirs, ours, recorded))
	lines = append(lines, cadenceConclusion(table, theirs, ours, recorded))
	lines = append(lines, "a count of the tick ack is informative in any case (AGENTS.md): what the server emits per tick "+
		"does not decide what the client can play")

	rep.Informative = append(rep.Informative, Finding{
		Rule:       RuleTickCadence,
		ClientPort: want.ClientPort,
		Detail: fmt.Sprintf("%s: the three sides of this capture do not send the tick ack at the same rate per second, "+
			"and the side that settles it is the third one", protocol.MessageTypeName(tickAck)),
		Lines: lines,
	})
}

// cadenceConclusion is the sentence, and it is built out of the numbers so it
// cannot say the opposite of what they say. It is two claims, and they need
// different amounts of evidence:
//
//   - THE TICK CADENCE IS THE SAME ON EVERY SIDE. A cadence is a rate, so three
//     sides that sustain the same rate have the same cadence. This needs the
//     sustained rates and nothing else.
//   - WHAT DIFFERS IS THE PHASE. That is a further claim: it needs the difference
//     to be at the start of the stream, because the phase argument is about where
//     the first tick lands inside its second. A difference in a later second is
//     not a cadence difference and the phase does not explain it either, and the
//     sentence says so instead of reaching for the explanation that is to hand.
//
// Where neither holds, nothing is concluded: the numbers do not separate "this
// build changed" from "the replay's input moved", and inventing a cause is worse
// than reporting that there is none.
func cadenceConclusion(t cadenceTable, theirs, ours, recorded []Elapsed) string {
	rt, ct, okt := sustainedRate(theirs)
	ro, co, oko := sustainedRate(ours)
	rr, cr, okr := sustainedRate(recorded)
	// The first second any side's acks occupy. The phase of the stream is decided
	// there and nowhere else.
	firstOccupied := -1
	for _, in := range [][]Elapsed{theirs, ours, recorded} {
		if s := firstAckSecondOf(in); s >= 0 && (firstOccupied < 0 || s < firstOccupied) {
			firstOccupied = s
		}
	}
	differsAtStart := false
	for _, r := range t.rows {
		if r.second == firstOccupied {
			differsAtStart = true
		}
	}
	if !(okt && oko && okr && rt == ro && ro == rr) {
		return cadenceUnseparated(t, ours, recorded)
	}
	out := fmt.Sprintf(
		"the sustained per-second rate is %d on all three sides, over %d, %d and %d of their seconds respectively, so the "+
			"TICK CADENCE IS THE SAME ON EVERY SIDE: a cadence is a rate, and there is none of it to find",
		rt, ct, co, cr)
	if !differsAtStart {
		return out + fmt.Sprintf(". The seconds that differ (%s) are not the second the ack stream starts in (t=%ds), "+
			"so this is not the phase of the start of the stream either, and no cause is claimed for it",
			t.String(), firstOccupied)
	}
	return out + fmt.Sprintf(
		". The second that differs IS the second the ack stream starts in (t=%ds), and what differs there is the PHASE: "+
			"the capture's clock is whole seconds (cmd/proxy/packets.go:47) and the replay feeds every message at exactly "+
			"its stamp, so a message written anywhere inside second N is fed at N.000, up to a second early, and at one "+
			"ack per tick that moves acks inside that one second and none in any second after it. Checked on the capture "+
			"itself by TestTheReplayFeedsEveryMessageAtItsWholeSecondStamp",
		firstOccupied)
}

// cadenceUnseparated is the shape where the numbers do not decide, and it says
// which of the two remaining readings the per-second gaps point at, if either.
func cadenceUnseparated(t cadenceTable, ours, recorded []Elapsed) string {
	worse := 0
	for _, r := range t.rows {
		if abs(r.ours-r.recorded) > abs(r.ours-r.theirs) {
			worse++
		}
	}
	if worse > 0 {
		return fmt.Sprintf(
			"this does not separate a rate difference from a phase one: the sustained rates do not agree across the three "+
				"sides. What it does say is that in %d of the %d second(s) that differ, our replay is FURTHER from the run's "+
				"own record of our binary than from the original, so that part of the difference belongs to the replay and "+
				"not to the server",
			worse, len(t.rows))
	}
	return "this does not separate the two explanations: the sustained rates do not agree across the three sides, and in " +
		"every second that differs our replay is no further from the run's own record of our binary than from the original. " +
		"Nothing is concluded: it may be the phase of the replay's input, or a change in this build since the capture was " +
		"made, and only the two together can say which"
}

// firstAckSecondOf is the whole second the first message of a series falls in.
func firstAckSecondOf(in []Elapsed) int {
	for _, e := range in {
		if e.Valid {
			return int(e.D / time.Second)
		}
	}
	return -1
}

// sustainedRate is the per-second rate a series holds once its first second is
// out of the way, and how many seconds hold it. The first second is excluded
// because it is the partial one: the ack stream starts when the first game data
// arrives, and how much of that second is left is not a rate.
//
// It reports ok=false when no single second value is strictly the most common,
// which is what happens on a stream with few seconds or a genuinely varying one.
// A rate that cannot be established is not a rate of 1 or of 0: guessing it would
// put a number in the middle of the sentence the conclusion is built out of.
func sustainedRate(in []Elapsed) (rate, seconds int, ok bool) {
	counts := map[int]int{}
	first := -1
	bySecond := map[int]int{}
	for _, e := range in {
		if !e.Valid {
			continue
		}
		s := int(e.D / time.Second)
		bySecond[s]++
		if first < 0 || s < first {
			first = s
		}
	}
	for s, n := range bySecond {
		if s == first {
			continue
		}
		counts[n]++
	}
	best, bestN := 0, 0
	for n, c := range counts {
		if c > bestN || (c == bestN && n > best) {
			best, bestN = n, c
		}
	}
	if bestN < 2 || bestN*2 < totalSeconds(bySecond, first) {
		return 0, bestN, false
	}
	return best, bestN, true
}

func totalSeconds(bySecond map[int]int, first int) int {
	n := 0
	for s := range bySecond {
		if s != first {
			n++
		}
	}
	return n
}

// renderRates names the sustained rate of each series, or says that it could not
// be established, which is a different fact from a rate of zero.
func renderRates(series ...[]Elapsed) string {
	what := []string{"original", "our replay", "the run's own record"}
	parts := make([]string, 0, len(series))
	for i, in := range series {
		rate, seconds, ok := sustainedRate(in)
		name := what[i]
		if i >= len(what) {
			name = fmt.Sprintf("series %d", i)
		}
		if !ok {
			parts = append(parts, name+" not established")
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %d/s over %d second(s)", name, rate, seconds))
	}
	return strings.Join(parts, ", ")
}

// cadenceRow is one second where the three series disagree.
type cadenceRow struct {
	second   int
	theirs   int
	ours     int
	recorded int
	note     string
}

type cadenceTable struct {
	rows    []cadenceRow
	dropped int
}

// cadenceSeconds buckets three series by whole second and keeps the seconds where
// they are not all equal. Whole seconds are the resolution both recorded sides can
// be read at, so this is the finest comparison of where a rate difference sits
// that does not invent precision the capture does not have.
func cadenceSeconds(series ...cadenceSeries) cadenceTable {
	buckets := make([]map[int]int, len(series))
	seen := map[int]bool{}
	for i, s := range series {
		buckets[i] = map[int]int{}
		for _, e := range s.elapsed {
			if !e.Valid {
				continue
			}
			sec := int(e.D / time.Second)
			buckets[i][sec]++
			seen[sec] = true
		}
	}
	var secs []int
	for sec := range seen {
		same := true
		for i := 1; i < len(buckets); i++ {
			if buckets[i][sec] != buckets[0][sec] {
				same = false
				break
			}
		}
		if !same {
			secs = append(secs, sec)
		}
	}
	sort.Ints(secs)
	t := cadenceTable{}
	shown := secs
	if len(shown) > maxCadenceSeconds {
		t.dropped = len(shown) - maxCadenceSeconds
		half := maxCadenceSeconds / 2
		shown = append(append([]int{}, shown[:half]...), shown[len(secs)-half:]...)
	}
	for _, sec := range shown {
		r := cadenceRow{second: sec, theirs: buckets[0][sec], ours: buckets[1][sec], recorded: buckets[2][sec]}
		// Which side is the odd one out, said where it can be said: a second where
		// the run's own record matches the original and our replay does not is the
		// shape the phase argument needs, and naming it is what saves a reader from
		// reading the row as a cadence difference.
		switch {
		case r.recorded == r.theirs && r.ours != r.theirs:
			r.note = "  <- our replay alone"
		case r.recorded != r.theirs && r.ours == r.theirs:
			r.note = "  <- the recorded run alone"
		}
		t.rows = append(t.rows, r)
	}
	return t
}

func countValid(in []Elapsed) int {
	n := 0
	for _, e := range in {
		if e.Valid {
			n++
		}
	}
	return n
}

// cadenceNote is the one-line form used by the tests, so a test can assert on the
// table without reading a rendered report.
func (t cadenceTable) String() string {
	parts := make([]string, 0, len(t.rows))
	for _, r := range t.rows {
		parts = append(parts, fmt.Sprintf("%ds: %d|%d|%d", r.second, r.theirs, r.ours, r.recorded))
	}
	if t.dropped > 0 {
		parts = append([]string{fmt.Sprintf("(%d more)", t.dropped)}, parts...)
	}
	return strings.Join(parts, " ")
}
