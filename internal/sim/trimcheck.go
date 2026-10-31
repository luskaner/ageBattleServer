package sim

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
)

// The proof that a cropped capture is still the capture it was.
//
// A crop is only allowed to remove the tail of a run. Everything else about it has
// to be exactly what the full capture would have said, and "exactly" is a claim
// about a replay that nobody has run, so it cannot be asserted by reading the crop.
// It has to be asserted by replaying BOTH and comparing what came out. That is what
// this file is: two reports in, a list of what the crop lost.
//
// # Why digests and not the messages
//
// The truth side of a 24 minute capture is tens of megabytes and a hundred thousand
// messages, and holding two of those at once to compare them is how a comparison
// becomes an O(n^2) memory problem. Instead every message contributes one 8-byte
// digest, in order, and the digests are what is compared. A crop that removed a
// message from the MIDDLE changes the digests after that point; a crop that removed
// one from the END leaves a prefix. That distinction is the whole proof, and it is
// what a count can never say: "the original sent 100 messages and we sent 100" is
// also what a reordered pair says.
//
// One digest per message, not per stream: a stream-wide rolling hash would answer
// "did anything change", which a prefix crop also answers with "yes", and then the
// proof would be unable to tell a crop from a rewrite.

// ConnMeasure is everything one connection contributed to a verdict, in a form two
// replays can be compared on without either of them being held in memory.
type ConnMeasure struct {
	ClientPort int
	Transport  string
	// Observer is whether the client asked to observe, which is what makes the
	// replay content measures meaningful at all.
	Observer bool
	// TruthUp is the client->server stream as the capture recorded it, and OursUp
	// the one the replay fed. They are the same bytes, and CheckTrim says so
	// rather than assuming it: a crop that altered the input would be comparing a
	// different match.
	TruthUp   StreamMeasure
	TruthDown StreamMeasure
	OursUp    StreamMeasure
	OursDown  StreamMeasure
	// TruthContent and OursContent are what the observer's replay stream carries.
	// Only filled for observers.
	TruthContent ContentMeasure
	OursContent  ContentMeasure
}

// StreamMeasure is one stream: what arrived, in what order, and how much of each
// type.
type StreamMeasure struct {
	// Totals is the per-type message count, which is the number a reader compares
	// by eye and the one the report prints.
	Totals map[byte]int
	// Digests is one entry per message, in order, over [type][body].
	Digests []uint64
}

func (s StreamMeasure) Count() int { return len(s.Digests) }

// Types renders the per-type totals in a stable order, so two measures can be put
// next to each other in a report and read.
func (s StreamMeasure) Types() string {
	if len(s.Totals) == 0 {
		return "none"
	}
	keys := make([]byte, 0, len(s.Totals))
	for k := range s.Totals {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("0x%02X x%d", k, s.Totals[k]))
	}
	return strings.Join(parts, " ")
}

// ContentMeasure is what a replay stream carries, read the way replaycontent.go
// reads it: the ordered sub-messages and the ordered checksum entries, with no
// regard for how many acks they were grouped into.
// ContentMeasure is what a replay stream carries, read the way replaycontent.go
// reads it: the ordered sub-messages and the ordered checksum entries, with no
// regard for how many acks they were grouped into.
//
// WindowStart is the match sub index the observer's first served sub-message sits
// at, and WindowKnown says whether it could be located at all. They are separate
// facts: a window that is not a sub-message of the match is not at index 0.
type ContentMeasure struct {
	Frames       int
	Subs         int
	Entries      int
	SubDigests   []uint64
	EntryDigests []uint64
	Unreadable   int
	WindowStart  int
	WindowKnown  bool
	WindowTotal  int
}

// measureConn takes one connection apart into its measures. The digests are over
// the exact bytes the rules compare, so a message that the byte comparison would
// mask (the 0x1E start time) is digested masked: a crop must not be able to pass by
// moving a field the comparison already ignores.
func measureConn(want, got *Conn, content *matchContent) ConnMeasure {
	m := ConnMeasure{
		ClientPort: want.ClientPort,
		Transport:  want.Transport,
		Observer:   requestedObservation(want),
		TruthUp:    measureStream(want.Up),
		TruthDown:  measureStream(want.Down),
		OursUp:     measureStream(got.Up),
		OursDown:   measureStream(got.Down),
	}
	if m.Observer {
		theirs := readAcks(replayStream(want.Down))
		ours := readAcks(replayStream(got.Down))
		m.TruthContent = contentMeasure(theirs, content.theirs)
		m.OursContent = contentMeasure(ours, content.ours)
	}
	return m
}

func measureStream(msgs []*Msg) StreamMeasure {
	s := StreamMeasure{Totals: map[byte]int{}}
	if len(msgs) == 0 {
		return s
	}
	s.Digests = make([]uint64, 0, len(msgs))
	for _, msg := range msgs {
		s.Totals[msg.Type]++
		s.Digests = append(s.Digests, digestOf(normalise(msg)))
	}
	return s
}

func contentMeasure(s ackStream, match []ackSub) ContentMeasure {
	c := ContentMeasure{Frames: s.frames, Subs: len(s.subs), Entries: len(s.entries), Unreadable: s.unreadable, WindowStart: -1}
	c.SubDigests = make([]uint64, 0, len(s.subs))
	for _, sub := range s.subs {
		c.SubDigests = append(c.SubDigests, digestOf(append([]byte{sub.slot}, sub.payload...)))
	}
	c.EntryDigests = make([]uint64, 0, len(s.entries))
	for _, e := range s.entries {
		c.EntryDigests = append(c.EntryDigests, digestOf(e))
	}
	if len(s.subs) > 0 {
		if i := subIndex(match, s.subs[0]); i >= 0 {
			c.WindowStart, c.WindowKnown, c.WindowTotal = i, true, len(match)
		}
	}
	return c
}

// digestOf is one 8-byte FNV-1a over b. It only has to be stable and different for
// different messages, and a hash collision on a 64-bit digest would have to invent
// a message that is byte-identical to another one in the same stream.
func digestOf(b []byte) uint64 {
	h := fnv.New64a()
	_, _ = h.Write(b)
	return h.Sum64()
}

// prefixProblem returns what a short stream has that the long one does not, at the
// first position where they part. A crop that kept a prefix has nothing to report.
// removedTypes is what the crop took off one stream, per type. It is a DELTA and
// not a second list of totals: a record that prints the untrimmed capture's counts
// under the word "dropped" reports a connection as having lost every message it
// sent, which for an observer that was already quiet is the opposite of the truth.
func removedTypes(short, long StreamMeasure) string {
	if short.Count() == long.Count() {
		return "nothing: the crop ended after this connection had said its last word"
	}
	keys := make([]byte, 0, len(long.Totals))
	for k := range long.Totals {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		if n := long.Totals[k] - short.Totals[k]; n > 0 {
			parts = append(parts, fmt.Sprintf("0x%02X x%d", k, n))
		}
	}
	if len(parts) == 0 {
		return "nothing"
	}
	return strings.Join(parts, " ")
}

func prefixProblem(shortName string, short, long StreamMeasure) string {
	n := min(len(short.Digests), len(long.Digests))
	for i := range n {
		if short.Digests[i] != long.Digests[i] {
			return fmt.Sprintf("%s message #%d is not the same message the untrimmed capture recorded there (%016x against %016x): the crop changed the middle of the stream, not its tail",
				shortName, i, short.Digests[i], long.Digests[i])
		}
	}
	return ""
}

func prefixListProblem(shortName string, short, long []uint64) string {
	n := min(len(short), len(long))
	for i := range n {
		if short[i] != long[i] {
			return fmt.Sprintf("%s #%d differs from the one the untrimmed capture recorded in that position: the crop changed the middle of the sequence, not its tail", shortName, i)
		}
	}
	return ""
}

// The rules CheckTrim decides by name, so a failure says which property of the
// crop broke instead of only that something did.
const (
	// TrimRuleVerdict is the whole point: the crop may not change the answer. The
	// same must-match findings, on the same connections, or the crop is rejected.
	TrimRuleVerdict = "trim-verdict"
	// TrimRuleTruthInput is the client's own bytes. A crop that altered what the
	// game said would be comparing a different match, and the prefix property is
	// what says it did not.
	TrimRuleTruthInput = "trim-truth-input"
	// TrimRuleTruthOutput is the original's answers. A crop may only take a suffix
	// off them.
	TrimRuleTruthOutput = "trim-truth-output"
	// TrimRuleContent is what the observer's replay carries: the sub-message
	// sequence, the checksum-entry sequence and the point of the match the window
	// starts at. All three are must-match in their own right, and a crop that kept
	// only a prefix cannot move any of them.
	TrimRuleContent = "trim-replay-content"
	// TrimRuleCoverage is the one that catches the failure mode this whole
	// arrangement exists to prevent: a crop so deep that the harness compares two
	// empty streams and says the capture passed. A connection that was compared
	// before has to be compared after, on both sides, and an observer that was
	// served frames has to be served frames again.
	TrimRuleCoverage = "trim-coverage"
)

// TrimProblem is one property the crop failed to preserve.
type TrimProblem struct {
	Conn   int
	Rule   string
	Detail string
}

// TrimCheck is the comparison of a cropped replay against the untrimmed one: what
// the crop cost, and whether any of it was something the verdict depends on.
type TrimCheck struct {
	// Kept and Dropped are the window, for the report.
	Kept    string
	Dropped string
	// Lines is the evidence, one per property, with the numbers it was decided on.
	// It goes into the capture's metadata (TrimRecord.Verified) and into the tool's
	// output, so the claim "this crop is safe" travels with the data it is about.
	Lines    []string
	Problems []TrimProblem
}

func (c TrimCheck) OK() bool { return len(c.Problems) == 0 }

func (c *TrimCheck) fail(conn int, rule, format string, args ...any) {
	c.Problems = append(c.Problems, TrimProblem{Conn: conn, Rule: rule, Detail: fmt.Sprintf(format, args...)})
}

func (c *TrimCheck) keep(conn int, format string, args ...any) {
	c.Lines = append(c.Lines, fmt.Sprintf("conn %d: %s", conn, fmt.Sprintf(format, args...)))
}

// CheckTrim decides whether a cropped capture still says the same thing as the
// capture it was cut from, and says so with numbers instead of adjectives.
//
// Both reports must come from Check on the two directories: the comparison is about
// what two REPLAYS produced, and a crop that only rewrote the data on disk would
// pass against a report nobody replayed.
//
// What is compared, and why each one:
//
//   - the verdict. Same must-match findings on the same connections. This is what
//     the crop exists to preserve, and it is checked first because everything else
//     is evidence for it.
//   - the truth's client->server stream, as a prefix. A crop that changed what the
//     game said would be replaying a different match and could still happen to
//     agree on the observer.
//   - the truth's server->client stream, as a prefix. This is the crop itself: the
//     recorded answer, minus a suffix, and nothing else.
//   - the observer's sub-message sequence, checksum-entry sequence and window
//     start, as prefixes of the same. They are the must-match measurements of
//     replaycontent, and no amount of cropping moves a sequence's head.
//   - coverage. Every connection compared before is compared after, and every
//     observer that was served frames is served frames again, on BOTH sides. A
//     crop that emptied a stream would make replay-page-order and replay-content
//     return without claiming anything, and the report would show a clean pass
//     over two empty things (SESSION-JOURNAL.md 24). That is the one outcome this
//     function exists to make impossible.
func CheckTrim(full, trimmed *Report) TrimCheck {
	c := TrimCheck{}
	if trimmed.Trim != nil {
		c.Kept, c.Dropped = trimmed.Trim.KeptSpan, trimmed.Trim.Dropped
	} else {
		c.Kept = "the whole capture (nothing was cut)"
	}

	if full.OK() != trimmed.OK() {
		c.fail(0, TrimRuleVerdict, "the two runs do not agree: the untrimmed capture %s, the trimmed one %s",
			verdictOf(full), verdictOf(trimmed))
	}
	if d := diffMustMatch(full, trimmed); d != "" {
		c.fail(0, TrimRuleVerdict, "%s", d)
	}
	c.keep(0, "verdict: %s, and the same must-match findings on the same connections (%d)", verdictOf(trimmed), len(trimmed.MustMatch))

	fullByPort := measuresByPort(full)
	trimmedByPort := measuresByPort(trimmed)
	for _, fm := range full.Measures {
		tm, ok := trimmedByPort[fm.ClientPort]
		if !ok {
			c.fail(fm.ClientPort, TrimRuleCoverage,
				"the connection the untrimmed capture compared is not in the trimmed one: the crop cut the whole of it, so nothing about it is being compared any more")
			continue
		}
		if fm.OursDown.Count() > 0 && tm.OursDown.Count() == 0 {
			c.fail(fm.ClientPort, TrimRuleCoverage,
				"we answered %d message(s) on this connection in the untrimmed capture and none in the trimmed one: the harness would be comparing an empty stream of ours against the original's, which is a pass that means nothing",
				fm.OursDown.Count())
		}
		if fm.TruthDown.Count() > 0 && tm.TruthDown.Count() == 0 {
			c.fail(fm.ClientPort, TrimRuleCoverage,
				"the original answered %d message(s) on this connection in the untrimmed capture and none in the trimmed one: there is no truth left here to compare against",
				fm.TruthDown.Count())
		}
		if d := prefixProblem("client->server", tm.TruthUp, fm.TruthUp); d != "" {
			c.fail(fm.ClientPort, TrimRuleTruthInput, "%s", d)
		} else {
			c.keep(fm.ClientPort, "the client's own messages: %d, and they are the first %d of the %d the untrimmed capture recorded. The crop removed %s",
				tm.TruthUp.Count(), tm.TruthUp.Count(), fm.TruthUp.Count(), removedTypes(tm.TruthUp, fm.TruthUp))
		}
		if d := prefixProblem("server->client", tm.TruthDown, fm.TruthDown); d != "" {
			c.fail(fm.ClientPort, TrimRuleTruthOutput, "%s", d)
		} else {
			c.keep(fm.ClientPort, "the original's answers: %d, and they are the first %d of the %d the untrimmed capture recorded. The crop removed %s from this connection",
				tm.TruthDown.Count(), tm.TruthDown.Count(), fm.TruthDown.Count(), removedTypes(tm.TruthDown, fm.TruthDown))
		}
		if tm.OursDown.Count() > fm.OursDown.Count() {
			c.fail(fm.ClientPort, TrimRuleTruthOutput,
				"the trimmed run has us sending %d message(s) on this connection where the untrimmed one had %d: a crop cannot make the server talk more, so the two runs are not the same experiment",
				tm.OursDown.Count(), fm.OursDown.Count())
		}
		if !fm.Observer {
			continue
		}
		if d := prefixListProblem("game-data sub-message", tm.TruthContent.SubDigests, fm.TruthContent.SubDigests); d != "" {
			c.fail(fm.ClientPort, TrimRuleContent, "on the truth side, %s", d)
		}
		if d := prefixListProblem("checksum entry", tm.TruthContent.EntryDigests, fm.TruthContent.EntryDigests); d != "" {
			c.fail(fm.ClientPort, TrimRuleContent, "on the truth side, %s", d)
		}
		if fm.TruthContent.WindowKnown != tm.TruthContent.WindowKnown || fm.TruthContent.WindowStart != tm.TruthContent.WindowStart {
			c.fail(fm.ClientPort, TrimRuleContent,
				"the observer's window starts somewhere else after the crop: %s against %s. The starting point of the replay is must-match, and no crop may move it",
				windowAt(tm.TruthContent), windowAt(fm.TruthContent))
		}
		if tm.TruthContent.Subs == 0 || tm.OursContent.Subs == 0 {
			c.fail(fm.ClientPort, TrimRuleContent,
				"after the crop the observer's replay carries no game data on %s, so replay-content has nothing to compare and would report a pass over two empty streams",
				sides(tm.TruthContent.Subs, tm.OursContent.Subs))
		}
		if fm.TruthContent.Subs > 0 && tm.TruthContent.Subs > fm.TruthContent.Subs {
			c.fail(fm.ClientPort, TrimRuleContent,
				"the crop left the observer with MORE sub-messages than the untrimmed capture (%d against %d), which no suffix of the same stream can do",
				tm.TruthContent.Subs, fm.TruthContent.Subs)
		}
		c.keep(fm.ClientPort, "the observer's replay: %s",
			contentSummary(tm.TruthContent, tm.OursContent, fm.TruthContent, fm.OursContent))
	}

	for _, tm := range trimmedByPort {
		if _, ok := fullByPort[tm.ClientPort]; !ok {
			c.fail(tm.ClientPort, TrimRuleCoverage,
				"the trimmed capture compares a connection the untrimmed one did not: a crop cannot add traffic to a run it did not record")
		}
	}
	return c
}

func measuresByPort(rep *Report) map[int]ConnMeasure {
	out := make(map[int]ConnMeasure, len(rep.Measures))
	for _, m := range rep.Measures {
		out[m.ClientPort] = m
	}
	return out
}

func verdictOf(rep *Report) string {
	if rep.OK() {
		return "matched the original on everything must-match"
	}
	return fmt.Sprintf("diverged from the original in %d must-match finding(s)", len(rep.MustMatch))
}

// diffMustMatch names the first must-match finding that is in one report and not in
// the other. A crop that hides a divergence is the failure this whole check exists
// to catch, and it is checked in both directions: losing a finding is as bad as
// gaining one, because a finding is a fact about the server that the harness is
// supposed to keep reporting.
func diffMustMatch(full, trimmed *Report) string {
	a, b := findingKeys(full), findingKeys(trimmed)
	for _, k := range b {
		if !contains(a, k) {
			return fmt.Sprintf("the trimmed capture has a must-match finding the untrimmed one does not: %s", k)
		}
	}
	for _, k := range a {
		if !contains(b, k) {
			return fmt.Sprintf("the trimmed capture LOST a must-match finding the untrimmed one had: %s. A crop may shorten a stream, but it may not stop a divergence from being reported", k)
		}
	}
	return ""
}

func findingKeys(rep *Report) []string {
	out := make([]string, 0, len(rep.MustMatch))
	for _, f := range rep.MustMatch {
		out = append(out, fmt.Sprintf("[%s] on client port %d", f.Rule, f.ClientPort))
	}
	sort.Strings(out)
	return out
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func windowAt(c ContentMeasure) string {
	if !c.WindowKnown {
		return "a point that is not a sub-message of the match"
	}
	return fmt.Sprintf("sub %d of %d", c.WindowStart, c.WindowTotal)
}

func sides(truth, ours int) string {
	return fmt.Sprintf("the original's (%d subs) and ours (%d subs)", truth, ours)
}

func contentSummary(tm, om, fullTruth, fullOurs ContentMeasure) string {
	return fmt.Sprintf(
		"the original is served %d sub-message(s) and %d checksum entr(ies) in %d ack(s), window at %s; we are served %d and %d in %d. "+
			"The untrimmed run served the original %d and %d in %d, and us %d and %d in %d. Every sequence is a prefix of its untrimmed one",
		tm.Subs, tm.Entries, tm.Frames, windowAt(tm),
		om.Subs, om.Entries, om.Frames,
		fullTruth.Subs, fullTruth.Entries, fullTruth.Frames,
		fullOurs.Subs, fullOurs.Entries, fullOurs.Frames)
}
