package sim

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Everything the report says about WHEN a divergence happened, in one file.
//
// The question this exists to answer is whether the messages we send and the
// original did not are an artifact of the harness reading on after the last
// captured input, or a difference in what the server emits during the match. A
// total cannot tell those apart: "we sent eleven messages more" reads the same
// either way. Only the times can, because the settle window has an end and the
// stream does not.
//
// One axis for both sides. Every elapsed time is counted from Capture.origin,
// which is the first message the client sent and is also the instant replay()
// schedules that same message from, so a message of ours and a message of the
// original at the same elapsed time were produced at the same moment of the same
// match.
//
// The axis is not equally precise on the two sides, and the report says so
// instead of pretending otherwise: the proxy stamps whole seconds
// (cmd/proxy/packets.go:47), so the original's moments are exact to the second
// while ours come from the replay's own clock. That asymmetry does not blur the
// question asked here, because the window's start is itself a difference of two
// whole-second stamps and therefore lands on a second boundary: a message
// stamped with that very second was sent at or after the boundary whatever
// instant inside the second it really was, and one stamped with any earlier
// second was sent before it.

// Elapsed is a moment on the replay's axis. Valid is false when the side has no
// recorded time at all, which is a different fact from "at zero": it means the
// message exists and nothing says when.
type Elapsed struct {
	D     time.Duration
	Valid bool
}

// elapsedAt is t counted from zero, and the zero cases are refused rather than
// folded into 0s: a difference between "0s" and "no time" would be invisible.
func elapsedAt(t, zero time.Time) Elapsed {
	if t.IsZero() || zero.IsZero() {
		return Elapsed{}
	}
	d := t.Sub(zero)
	if d < 0 {
		d = 0
	}
	return Elapsed{D: d, Valid: true}
}

// elapsedOf is a message's own moment counted from zero.
func elapsedOf(m *Msg, zero time.Time) Elapsed {
	if m == nil {
		return Elapsed{}
	}
	return elapsedAt(m.At, zero)
}

// String renders the moment as a duration, rounded to 10ms so the millisecond
// noise of the replay's own clock is not read as a fact about the protocol.
func (e Elapsed) String() string {
	if !e.Valid {
		return "no time recorded"
	}
	return e.D.Round(10 * time.Millisecond).String()
}

// Window is the time the replay kept reading after the last captured input, so
// the session ticks had room to deliver what they owe. A message of ours that
// arrives inside it was produced with no input left to cause it, which is the
// whole difference between "the harness read on past the last message" and "the
// server emits more than the original".
type Window struct {
	// Start is when the last captured input was fed, which is the span of the
	// capture's client->server stream: its last message, on this axis.
	Start time.Duration
	// End is when the connection was closed: Start plus the settle.
	End time.Duration
	// Speed is the factor the input was fed at. Anything but 1 means the two
	// series are not on the same axis, because the truth's moments are real time
	// and the compressed replay's are not, and the report has to say so rather
	// than print two series that cannot be laid on top of each other.
	Speed float64
}

func (w Window) String() string {
	if w.Speed != 0 && w.Speed != 1 {
		return fmt.Sprintf("[t=%s, t=%s, the input was fed at x%.2f]", w.Start, w.End, w.Speed)
	}
	return fmt.Sprintf("[t=%s, t=%s]", w.Start, w.End)
}

// after reports whether a moment is at or after the window opened.
func (w Window) after(e Elapsed) bool { return e.Valid && e.D >= w.Start }

// Where an extra message is, and why not by lining the two streams up.
//
// The obvious way to find an extra message is to align the two streams and look
// at what is left over, resynchronising wherever they stop agreeing. On this
// protocol that is not usable, and the reason is worth recording: the server
// emits one 0x0A per simulation tick and hundreds of them in a row are the same
// six bytes, so a walk down two such streams resynchronises at nearly every step
// and declares a message extra at each one. Measured on the first real capture,
// that is what it did: 382 "placed" extras for 7 real ones. A resynchronisation
// there is a guess, and a guess with a timestamp on it is worse than no guess,
// because it reads as a measurement.
//
// So nothing is located by aligning. Payloads are counted instead (payloadDiff),
// which is the one statement about an extra message that does not depend on the
// order the two streams happened to be in: a payload we sent one more time than
// the original ever did is a payload the original did not send that many times.
// When even that is ambiguous, because the surplus is N copies of a payload that
// repeats all match long, the report says so and counts per second, which is
// evidence and not an attribution.

// surplus is one payload we sent more often than the original did, and when we
// sent it. The counts are facts about the two byte streams; the times are ours,
// and they are not presented as the times of the surplus copies, because which
// M-N of our M copies are the extra ones is not something the bytes say.
type surplus struct {
	body   []byte
	ours   int // how many we sent in all
	theirs int
	times  []Elapsed
}

// maxSurplusListed bounds the payload table. A type whose every message differs
// can have hundreds of surplus payloads and the reader wants the ones that
// matter, which is why the window split and the per-second counts come first.
const maxSurplusListed = 6

// payloadDiff compares the two streams of one type payload by payload.
//
// over is the payloads we sent more often than the original and under is how many
// messages we sent fewer of. The two are what decide whether the payload counts
// localise the difference: over-theirs, with under zero, means the surplus is
// exactly a set of payloads the original did not send, and anything else means
// the two sides' payloads are not the same bytes at all, which is a different
// finding and has to be reported as one.
func payloadDiff(ours, theirs []*Msg, zero time.Time) (over []surplus, under int) {
	theirsCount := map[string]int{}
	for _, m := range theirs {
		theirsCount[string(m.Body)]++
	}
	oursCount := map[string]int{}
	times := map[string][]Elapsed{}
	for _, m := range ours {
		k := string(m.Body)
		oursCount[k]++
		if e := elapsedOf(m, zero); e.Valid {
			times[k] = append(times[k], e)
		}
	}
	for k, n := range oursCount {
		if d := n - theirsCount[k]; d > 0 {
			over = append(over, surplus{body: []byte(k), ours: n, theirs: theirsCount[k], times: times[k]})
		}
	}
	// The other half has to be counted too, from the payloads we never sent at
	// all: a type whose payloads differ between two runs (the real tick acks do)
	// has a surplus on both sides, and counting only ours would make the
	// difference look like a big one-sided extra.
	for k, n := range theirsCount {
		if _, ok := oursCount[k]; !ok {
			under += n
		}
	}
	sort.Slice(over, func(i, j int) bool { return over[i].ours-over[i].theirs > over[j].ours-over[j].theirs })
	return over, under
}

// total is how many surplus messages the whole list accounts for.
func total(list []surplus) int {
	n := 0
	for _, s := range list {
		n += s.ours - s.theirs
	}
	return n
}

// String renders one surplus payload: what it is, how many of it each side sent
// and when. The times are ours and are NOT presented as the times of the surplus
// copies: when we sent a payload M times and the original N < M times, which
// M-N of ours are the extra ones is not something the bytes say. The counts are
// facts, and the line gives only facts.
func (s surplus) String() string {
	return fmt.Sprintf("len=%d: we sent it %d time(s) at %s, the original %d",
		len(s.body), s.ours, renderTimes(s.times, 4), s.theirs)
}

// maxSecondBuckets bounds the per-second table. A run of ninety identical
// numbers answers nothing and buries the two that decide the question, so the
// line keeps the first half and the last half and says how many it dropped.
const maxSecondBuckets = 12

// secondBuckets counts both sides per whole second of elapsed time and renders
// only the seconds where the two disagree, as "26s: 5|5, 89s: 1|4" with the
// original first. Whole seconds are the resolution both sides can be read at,
// so this is the finest comparison of where a difference sits that does not
// invent precision the truth side does not have.
func secondBuckets(ours, theirs []Elapsed, max int) string {
	if max <= 0 {
		max = maxSecondBuckets
	}
	mine, theirsBy := map[int]int{}, map[int]int{}
	first := map[int]bool{}
	for _, e := range ours {
		if !e.Valid {
			continue
		}
		s := int(e.D / time.Second)
		mine[s]++
		first[s] = true
	}
	for _, e := range theirs {
		if !e.Valid {
			continue
		}
		s := int(e.D / time.Second)
		theirsBy[s]++
		first[s] = true
	}
	var secs []int
	for s := range first {
		if mine[s] != theirsBy[s] {
			secs = append(secs, s)
		}
	}
	if len(secs) == 0 {
		return ""
	}
	sort.Ints(secs)
	shown := secs
	dropped := 0
	if len(shown) > max {
		dropped = len(shown) - max
		half := max / 2
		shown = append(append([]int{}, shown[:half]...), shown[len(secs)-half:]...)
	}
	parts := make([]string, 0, len(shown)+1)
	for _, s := range shown {
		parts = append(parts, fmt.Sprintf("%ds: %d|%d", s, theirsBy[s], mine[s]))
	}
	if dropped > 0 {
		parts = append([]string{fmt.Sprintf("(%d more seconds differ)", dropped)}, parts...)
	}
	return strings.Join(parts, " ")
}

// TailVerdict is the conclusion about one connection: where the messages we
// sent and the original did not are. The counts are exact on both sides, which
// is what makes the conclusion one rather than a range, and the split is checked
// to add up in TestJudgeTailArithmeticCloses.
//
// BeforeWindow and TheirsBefore count messages strictly before the settle
// window opened. The original's is exact rather than a bound: its stamps are
// whole seconds and the window opens on a second boundary, so a message stamped
// with that second is at or after the window whatever instant inside the second
// it really was, and one stamped earlier is before it. A message with no
// recorded time is counted on neither side of the split.
type TailVerdict struct {
	// Extra is ours minus theirs, the number the report has to explain.
	Extra int
	// Ours is what we sent in all, split by the window into BeforeWindow and
	// InWindow.
	Ours         int
	BeforeWindow int
	InWindow     int
	// Theirs is what the original sent in all, split the same way.
	Theirs       int
	TheirsBefore int
	TheirsAfter  int
	// ExtrasBeforeWindow is how many of the extras had to be emitted before the
	// settle window opened, which is what separates the two explanations.
	ExtrasBeforeWindow int
	// SettleOnly is true when every extra message falls at or after the window
	// opened, which is what makes the difference possibly an artifact of the
	// harness reading on past the last captured input.
	SettleOnly bool
	// AxisWarn is set when the input was fed faster than real time, in which case
	// our moments and the original's are not on the same axis and the split
	// above measures the harness rather than the server.
	AxisWarn bool
	// Reason is the sentence the report prints. It is built from the counts, so
	// it cannot say the opposite of what the numbers say.
	Reason string
}

// onAxis puts a stream of messages on the axis: every moment counted from its
// own side's zero, which for the original is the capture's origin and for ours
// the instant the replay started. The two are the same instant by construction,
// which is what lets the two series be laid on top of each other.
func onAxis(msgs []*Msg, zero time.Time) []Elapsed {
	out := make([]Elapsed, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, elapsedOf(m, zero))
	}
	return out
}

// judgeTail splits both streams around the settle window and says where the
// difference is. It is a pure function of two lists of moments and the window,
// so the conclusion can be driven both ways by a test: extras in the window, and
// extras before it. A moment with no time is counted on neither side of the
// split, and the report says so rather than assuming it to be zero.
func judgeTail(ours, theirs []Elapsed, w Window) TailVerdict {
	v := TailVerdict{
		Ours:     len(ours),
		Theirs:   len(theirs),
		Extra:    len(ours) - len(theirs),
		AxisWarn: w.Speed != 0 && w.Speed != 1,
	}
	for _, e := range ours {
		if e.Valid && w.after(e) {
			v.InWindow++
		}
	}
	for _, e := range theirs {
		if e.Valid && !w.after(e) {
			v.TheirsBefore++
		}
	}
	v.BeforeWindow = v.Ours - v.InWindow
	v.TheirsAfter = v.Theirs - v.TheirsBefore
	v.ExtrasBeforeWindow = v.BeforeWindow - v.TheirsBefore
	// Every extra is at or after the window exactly when none of them is before
	// it. A negative number of extras before the window means the original sent
	// more before the window than we did, which is not "settle only" either and
	// must not be read as one.
	v.SettleOnly = v.Extra > 0 && v.ExtrasBeforeWindow <= 0
	v.Reason = v.explain(w)
	return v
}

// explain is the sentence. It reads the numbers rather than the numbers being
// read into it: the two conclusions are mutually exclusive and the test drives
// both.
func (v TailVerdict) explain(w Window) string {
	var b strings.Builder
	switch {
	case v.Extra > 0 && v.SettleOnly:
		fmt.Fprintf(&b, "all %d message(s) we sent and it did not fall at or after the moment the settle window opened at t=%s, "+
			"and none before it: the capture stops when the client disconnects and the harness keeps reading for the settle "+
			"afterwards, so this MAY BE AN ARTIFACT of the harness reading on rather than of the tick cadence", v.Extra, w.Start)
	case v.Extra > 0:
		fmt.Fprintf(&b, "%d of the %d message(s) we sent and it did not were emitted BEFORE the settle window opened at t=%s "+
			"(our stream holds %d message(s) before that moment against the original's %d, which is exact because the "+
			"original's stamps are whole seconds and the window opens on a second boundary), and %d of ours at or after it "+
			"against %d of its: THE EXTRAS ARE INTERLEAVED THROUGH THE MATCH, not an artifact of the settle",
			v.ExtrasBeforeWindow, v.Extra, w.Start, v.BeforeWindow, v.TheirsBefore, v.InWindow, v.TheirsAfter)
	case v.Extra < 0:
		fmt.Fprintf(&b, "we sent %d message(s) and the original %d, so we are %d short: our stream ends earlier than "+
			"the original's, which the settle window cannot explain (it makes us read longer, not shorter)",
			v.Ours, v.Theirs, -v.Extra)
	default:
		b.WriteString("both streams carry the same number of messages")
	}
	if v.AxisWarn {
		fmt.Fprintf(&b, ". CAREFUL: this run fed the input at x%.2f, so our moments and the original's are not on the same axis and this split measures the harness", w.Speed)
	}
	return b.String()
}
