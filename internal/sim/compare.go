package sim

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// The verdict levels. A must-match divergence is what the client depends on and
// fails the test. Everything else is informative: recorded and printed, never
// fatal. The reason for the split is that the original has quirks of its own
// (CAPTURES.md row 8: 2 of 26 LOBBYPING answered in _26) and a test that fails
// on those ends up switched off entirely, which costs more than the divergence
// it reported.
const (
	// RuleStartObserveStatus is the byte of the 0x19 StartObserveResponse. It is
	// the accept answer on the wire, and it is the one value that decided the
	// observer investigation: _26 answered 0 and served 7343 pages, _30 answered
	// 3, the client never sent ActivateObserver and reported that the start of
	// the match was gone (SESSION-JOURNAL.md, "Correction of the 22").
	RuleStartObserveStatus = "start-observe-status"
	// RuleActivateObserverResponse is whether a 0x1E ever arrives, and with
	// which status. The client's next step is blocked without it, so its absence
	// is the failure _30 died of.
	RuleActivateObserverResponse = "activate-observer-response"
	// RuleReplayPageOrder is the replay stream the observer is served: the order
	// of its frames, and the point of the match the first one is. The stream is
	// the session's own 0x0A ack frames, because that is what the frame store
	// holds (it is fed the ack the tick just built) and what
	// observers are served. An earlier revision read 0x09 here, which
	// made the rule vacuous: no observer is sent a 0x09 at all, so
	// "the reference served 0 and we served 489" compared nothing and the second
	// observe attempt of 2026-09-29_012528_03cb67c7cf19 came out with 0
	// must-match divergences while it was serving 70 messages against 803.
	RuleReplayPageOrder = "replay-page-order"
	// RuleReplayContent is the same stream read by what it carries rather than
	// by which frames of the match it is made of, so it still says something when
	// the two sides group the acks differently. See replaycontent.go.
	RuleReplayContent = "replay-content"
	// RuleUpgradeResponse is the WebSocket 101 answer, which is bytes the original
	// puts on the wire before any battle message exists. It was the one part of
	// the exchange nothing compared, and it is where a header the original sends
	// and we do not was living. See upgrade.go.
	RuleUpgradeResponse = "upgrade-response"

	// The informative rules follow.
	RuleMessageSequence = "message-sequence"
	RuleBodyMismatch    = "body-mismatch"
	RuleMessageCount    = "message-count"
	// RuleMessageTail is where the messages we sent and the original did not are:
	// produced during the match, or produced during the settle window the harness
	// keeps reading for after the last captured input. A total cannot tell those
	// apart, which is why it is a rule of its own and not a remark.
	RuleMessageTail    = "message-tail"
	RuleFraming        = "framing"
	RuleInputAgreement = "input-agreement"
	RuleCoverage       = "not-compared"
	RuleDroppedFlag    = "dropped-flag"
	RuleDeviation      = "config-deviation"
	RuleConfigDiff     = "config-diff"
)

// MaskedField is a byte range of a message body that is not ours to reproduce.
// The list is printed in every report, so what the comparison ignored is
// visible instead of being buried in the code.
type MaskedField struct {
	Type byte
	From int
	To   int
	Why  string
}

var maskedFields = []MaskedField{
	{
		Type: protocol.MsgActivateObserverResp,
		From: 1, To: 9,
		Why: "the estimated game start time: Unix seconds plus an offset derived from the page count " +
			"(internal/session/session.go, Session.ActivateObserver). It is the moment this process started, " +
			"which no replay can reproduce, and the client uses it as a start time, not as a check",
	},
}

// maskFill stands in for a masked byte. It has to be a value a real body is
// unlikely to carry, so that a difference in a masked range can never pass for a
// real one.
const maskFill = 0xA5

// Finding is one divergence, with both sides rendered for the report. A test
// that fails without the two frames side by side sends the reader back to the
// capture, which is the thing this harness exists to avoid.
type Finding struct {
	Rule       string
	MustMatch  bool
	ClientPort int
	Index      int
	Detail     string
	Ours       string
	Theirs     string
	// AtOurs and AtTheirs are when the two sides said it, counted from the start
	// of the replay. They are what separates "the extras are all in the settle
	// window" from "the extras are spread through the match", which no index and
	// no type can say. Valid is false when that side has no time at all, and
	// that is a different fact from t=0.
	AtOurs   Elapsed
	AtTheirs Elapsed
	// Lines carry what the rule has to add about the finding: the moments, the
	// per-second counts, the conclusion. They are kept out of Detail so the
	// headline stays one line and a reader can scan the list and stop at the
	// finding that answers the question.
	Lines []string
}

// ConnReport is what happened to one connection, so a report can show that a
// connection was compared at all. A connection nobody looked at must never look
// like one that passed.
type ConnReport struct {
	ClientPort int
	Transport  string
	Compared   bool
	Sent       int
	Ours       int
	Theirs     int
	// OursLast and TheirsLast are when each side stopped sending. Two streams of
	// different lengths are otherwise two numbers, and the question of whether
	// the difference is at the end or through the middle starts with these two.
	OursLast   Elapsed
	TheirsLast Elapsed
	Notes      []string
}

// Report is the whole verdict.
type Report struct {
	Capture         string
	Commit          string
	Dirty           string
	Reference       string
	ReferenceWriter string
	OtherSide       string
	// OtherFile and OtherWriter name the other log of the capture on its own,
	// because the tick-cadence rule quotes them as a source: it is the run of the
	// other binary on the same client and the same clock, and a reader has to be
	// able to see which log that is and who wrote it.
	OtherFile   string
	OtherWriter string
	CommandLine []string
	TruthCommand []string
	ConfigDiff      []string
	DroppedFlags    []string
	Running         []Deviation
	Masked          []MaskedField
	Notes           []string

	Conns       []ConnReport
	MustMatch   []Finding
	Informative []Finding
	// Measures is every connection reduced to ordered digests and per-type totals,
	// so two replays of the same capture can be compared without holding either
	// one's messages. It is what CheckTrim reads, and nothing else: the rules
	// above work on the messages, and a rule that read digests would not be able
	// to show its two frames.
	Measures []ConnMeasure
	// Trim is this capture's crop record, when it is a cropped one. The report
	// prints it in full: a reader of a failing test has to know that the capture
	// under it is three minutes of a twenty-five minute run, and what was checked
	// before that was called safe.
	Trim *TrimRecord

	// started, origin and window are the one time axis: our moments are counted
	// from started, the instant the replay scheduled the capture's origin to, and
	// the original's from that same origin, so an elapsed time means the same
	// moment of the same match on both sides. They are not printed as clock
	// times because nothing in the comparison is about the wall clock.
	started time.Time
	origin  time.Time
	window  Window
}

// OK reports whether the replay matched the original on everything the client
// depends on.
func (r *Report) OK() bool { return len(r.MustMatch) == 0 }

// mustMatch records a divergence that fails the test. The flag travels with the
// finding so the rendered report cannot label a must-match as informative: which
// list it is in and how it is described are the same fact.
func (r *Report) mustMatch(f Finding) {
	f.MustMatch = true
	r.MustMatch = append(r.MustMatch, f)
}

// FirstDivergence renders the first divergence, must-match first, with the two
// frames side by side.
func (r *Report) FirstDivergence() string {
	f := r.first()
	if f == nil {
		return "  (no divergence)"
	}
	level := "informative"
	if f.MustMatch {
		level = "MUST-MATCH"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "  [%s] %s | client port %d | index %d\n", level, f.Rule, f.ClientPort, f.Index)
	if when := whenLine(f.AtOurs, f.AtTheirs); when != "" {
		fmt.Fprintf(&b, "    %s\n", when)
	}
	fmt.Fprintf(&b, "    %s\n", f.Detail)
	for _, l := range f.Lines {
		fmt.Fprintf(&b, "    %s\n", l)
	}
	fmt.Fprintf(&b, "    original: %s\n", blankIfEmpty(f.Theirs))
	fmt.Fprintf(&b, "    ours    : %s", blankIfEmpty(f.Ours))
	return b.String()
}

// whenLine renders the two sides' moments. A side with no moment to show says
// so, because "t=0" would read as a message the original sent at the very start
// of the match, and a message that is absent and a message whose arrival was
// never recorded are both reported as the absence of a time, with the absence
// itself spelled out in the finding's Detail.
func whenLine(ours, theirs Elapsed) string {
	switch {
	case ours.Valid && theirs.Valid:
		return fmt.Sprintf("when: ours t=%s | original t=%s", ours, theirs)
	case ours.Valid:
		return fmt.Sprintf("when: ours t=%s | original: no time on this side", ours)
	case theirs.Valid:
		return fmt.Sprintf("when: original t=%s | ours: no time on this side", theirs)
	default:
		return ""
	}
}

func (r *Report) first() *Finding {
	if len(r.MustMatch) > 0 {
		return &r.MustMatch[0]
	}
	// A divergence about a connection is what a reader can act on: it carries
	// two frames. A capture-level finding (a config difference, a dropped flag)
	// only wins when there is nothing per-connection to show.
	for i := range r.Informative {
		if r.Informative[i].ClientPort != 0 {
			return &r.Informative[i]
		}
	}
	if len(r.Informative) > 0 {
		return &r.Informative[0]
	}
	return nil
}

func blankIfEmpty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(nothing: the other side sent no such frame)"
	}
	return s
}

// Render is the report a generated test logs: what the truth was, what the
// server ran with, what was ignored, and every divergence with its two frames.
func (r *Report) Render() string {
	var b strings.Builder
	fmt.Fprintf(&b, "sim: capture %s (commit %s, dirty=%s)\n", r.Capture, r.Commit, r.Dirty)
	fmt.Fprintf(&b, "  truth      : %s writer=%s\n", r.Reference, r.ReferenceWriter)
	fmt.Fprintf(&b, "  other side : %s\n", r.OtherSide)
	fmt.Fprintf(&b, "  our config : %s\n", strings.Join(r.CommandLine, " "))
	fmt.Fprintf(&b, "  truth's    : %s\n", strings.Join(r.TruthCommand, " "))
	if len(r.ConfigDiff) > 0 {
		for _, d := range r.ConfigDiff {
			fmt.Fprintf(&b, "  config diff: %s\n", d)
		}
	}
	if len(r.DroppedFlags) > 0 {
		fmt.Fprintf(&b, "  dropped    : %s (this parser does not know them; our default is the closest reproduction)\n",
			strings.Join(r.DroppedFlags, ", "))
	}
	if len(r.Running) == 0 {
		b.WriteString("  deviations : none: the server ran with exactly the recorded command line\n")
	}
	for _, d := range r.Running {
		fmt.Fprintf(&b, "  deviation  : %s recorded=%s running=%s\n               %s\n", d.Field, d.Recorded, d.Running, d.Why)
	}
	for _, m := range r.Masked {
		fmt.Fprintf(&b, "  masked     : 0x%02X body[%d:%d] %s\n", m.Type, m.From, m.To, m.Why)
	}
	// The axis, before any number that is read on it. Two facts a reader has to
	// have before "t=26.31s" means anything: what t counts from, and that the
	// original's is a whole-second stamp and ours is not.
	fmt.Fprintf(&b, "  time axis  : t is elapsed since the first message the client sent, which is the instant the replay\n")
	fmt.Fprintf(&b, "               started feeding the input. The original's t is its proxy's stamp, which is whole\n")
	fmt.Fprintf(&b, "               seconds (cmd/proxy/packets.go:47), so it is exact to the second; ours is the replay's clock.\n")
	if r.Trim != nil {
		r.renderTrim(&b)
	}
	fmt.Fprintf(&b, "  window     : the last input was fed at t=%s and the connection was closed at t=%s; a message of ours\n",
		r.window.Start, r.window.End)
	fmt.Fprintf(&b, "               inside that window was produced with no input left to cause it\n")
	for _, c := range r.Conns {
		state := "compared"
		if !c.Compared {
			state = "NOT COMPARED"
		}
		fmt.Fprintf(&b, "  conn %-6d %-10s %-12s sent=%d received ours=%d (%s) original=%d (%s)\n",
			c.ClientPort, c.Transport, state, c.Sent, c.Ours, endsAt(c.OursLast), c.Theirs, endsAt(c.TheirsLast))
		for _, n := range c.Notes {
			fmt.Fprintf(&b, "      note: %s\n", n)
		}
	}
	for _, n := range r.Notes {
		fmt.Fprintf(&b, "  note: %s\n", n)
	}
	fmt.Fprintf(&b, "  MUST-MATCH  : %d divergence(s)\n", len(r.MustMatch))
	fmt.Fprintf(&b, "  INFORMATIVE : %d divergence(s)\n", len(r.Informative))
	for _, f := range r.Informative {
		fmt.Fprintf(&b, "    [%s] conn %d #%d: %s\n", f.Rule, f.ClientPort, f.Index, f.Detail)
		if when := whenLine(f.AtOurs, f.AtTheirs); when != "" {
			fmt.Fprintf(&b, "        %s\n", when)
		}
		for _, l := range f.Lines {
			fmt.Fprintf(&b, "        %s\n", l)
		}
		if f.Ours != "" || f.Theirs != "" {
			fmt.Fprintf(&b, "        original: %s\n", blankIfEmpty(f.Theirs))
			fmt.Fprintf(&b, "        ours    : %s\n", blankIfEmpty(f.Ours))
		}
	}
	if len(r.MustMatch) > 1 {
		b.WriteString("  the remaining must-match divergences:\n")
		for _, f := range r.MustMatch[1:] {
			fmt.Fprintf(&b, "    [%s] conn %d #%d: %s\n", f.Rule, f.ClientPort, f.Index, f.Detail)
			if when := whenLine(f.AtOurs, f.AtTheirs); when != "" {
				fmt.Fprintf(&b, "        %s\n", when)
			}
		}
	}
	fmt.Fprintf(&b, "  first divergence:\n%s", r.FirstDivergence())
	return b.String()
}

// endsAt renders the end of a stream for the connection table, where a stream
// with no recorded time has to look different from one that ends at t=0.
func endsAt(e Elapsed) string {
	if !e.Valid {
		return "no time recorded"
	}
	return "last t=" + e.String()
}

// renderTrim prints the crop this capture is, in full, where a reader cannot miss
// it. The proof is printed whether or not it is there: a crop whose evidence is
// empty has to read as unchecked, because "no evidence" and "checked and fine" are
// different facts and a capture that looks like the second one when it is the first
// is worse than an uncropped one.
func (r *Report) renderTrim(b *strings.Builder) {
	t := r.Trim
	fmt.Fprintf(b, "  TRIMMED    : this capture is %s of the run, kept %s from %s to %s. %s\n",
		t.KeptSpan, t.KeptSpan, t.KeptFrom, t.KeptTo, t.Why)
	fmt.Fprintf(b, "               cut from %s by %s, %s. Nothing about the replay changes when a capture is\n",
		t.Source, t.Tool, t.Dropped)
	fmt.Fprintf(b, "               cropped: the capture is still replayed in real time, because the server emits\n")
	fmt.Fprintf(b, "               0x0A per tick and a faster replay falsifies the counts (Options.ReplaySpeed)\n")
	if len(t.Verified) == 0 {
		fmt.Fprintf(b, "               NOT VERIFIED: nothing has established that this crop preserved the verdict\n")
	}
	for _, l := range t.Verified {
		fmt.Fprintf(b, "               verified: %s\n", l)
	}
	if t.Leftover != "" {
		fmt.Fprintf(b, "               %s\n", t.Leftover)
	}
}

// compare runs every rule over the replay and fills the report.
func compare(rep *Report, cap *Capture, ours []*Conn) {
	byPort := make(map[int]*Conn, len(ours))
	for _, o := range ours {
		byPort[o.ClientPort] = o
	}

	if len(rep.ConfigDiff) > 0 {
		rep.Informative = append(rep.Informative, Finding{
			Rule: RuleConfigDiff,
			Detail: "the two sides of the capture were launched with different settings, so the replay reproduces ours " +
				"and the truth is the original's: " + strings.Join(rep.ConfigDiff, ", "),
		})
	}
	for _, f := range rep.DroppedFlags {
		rep.Informative = append(rep.Informative, Finding{
			Rule:   RuleDroppedFlag,
			Detail: fmt.Sprintf("%s is in the recorded command line but not in our flag set, so the replay used our default for it", f),
		})
	}
	for _, d := range rep.Running {
		if d.Field == "command line" {
			continue
		}
		rep.Informative = append(rep.Informative, Finding{
			Rule:   RuleDeviation,
			Detail: fmt.Sprintf("%s: recorded=%s running=%s (%s)", d.Field, d.Recorded, d.Running, d.Why),
		})
	}

	// The match's own ack stream, used to say which frame of the match a replay
	// frame is. Built before the per-connection rules because the observer
	// connections are compared against the player's, and which connection is the
	// player's is a fact about the capture rather than about any one of them.
	match := buildMatchIndex(cap.Conns, byPort)
	// The same stream taken apart by content, for the rule that has to work when
	// the two sides grouped the acks differently and no whole frame identifies.
	content := buildMatchContent(cap.Conns, byPort)

	compared := 0
	for _, want := range cap.Conns {
		got, ok := byPort[want.ClientPort]
		if !ok {
			got = &Conn{ClientPort: want.ClientPort, Transport: want.Transport,
				Notes: []string{"no replay result for this client port"}}
		}
		cr := ConnReport{
			ClientPort: want.ClientPort,
			Transport:  want.Transport,
			Sent:       len(want.Up),
			Ours:       len(got.Down),
			Theirs:     len(want.Down),
			OursLast:   lastAt(got.Down, rep.started),
			TheirsLast: lastAt(want.Down, rep.origin),
			Notes:      append(append([]string{}, want.Notes...), got.Notes...),
		}
		// cr is appended once its fate is decided: a report that says a
		// connection was compared has to be the same report that carries its
		// notes, and appending a copy first would print "NOT COMPARED" over a
		// connection that was in fact compared.
		checkInput(rep, want)
		if len(want.Down) == 0 {
			rep.Informative = append(rep.Informative, Finding{
				Rule:       RuleCoverage,
				ClientPort: want.ClientPort,
				Detail:     "the original never sent anything on this connection, so there is nothing to compare our answer against",
			})
			cr.Notes = append(cr.Notes, "no truth: the original said nothing here")
			rep.Conns = append(rep.Conns, cr)
			continue
		}
		if replayFailed(got) {
			rep.Informative = append(rep.Informative, Finding{
				Rule:       RuleCoverage,
				ClientPort: want.ClientPort,
				Detail:     "the replay of this connection did not complete, so its answer was not compared",
			})
			cr.Notes = append(cr.Notes, "replay did not complete: not compared")
			rep.Conns = append(rep.Conns, cr)
			continue
		}
		cr.Compared = true
		rep.Conns = append(rep.Conns, cr)
		// Measured for every connection that was compared, and the measurement is
		// the messages themselves reduced to digests: a crop is checked against
		// the untrimmed replay's, and that comparison must be able to see a
		// message that moved as well as one that is missing.
		rep.Measures = append(rep.Measures, measureConn(want, got, content))
		compared++
		compareConn(rep, want, got, match, content)
	}
	rep.Conns = sortedConns(rep.Conns)

	if compared == 0 {
		rep.Informative = append(rep.Informative, Finding{
			Rule:   RuleCoverage,
			Detail: "no connection of this capture could be compared against the original: the capture says nothing about what this server does",
		})
	}
}

func replayFailed(c *Conn) bool {
	for _, n := range c.Notes {
		if strings.HasPrefix(n, "replay failed") || strings.HasPrefix(n, "not replayed") {
			return true
		}
	}
	return false
}

func sortedConns(in []ConnReport) []ConnReport {
	sort.Slice(in, func(i, j int) bool { return in[i].ClientPort < in[j].ClientPort })
	return in
}

// checkInput cross-checks the two sides of the capture against each other. The
// proxy duplicates every client write to both children, so the client->server
// stream has to be identical in both logs; when it is not, the capture is the
// thing in doubt and neither server's answer can be blamed.
func checkInput(rep *Report, want *Conn) {
	if len(want.otherUp) == 0 {
		return
	}
	n := min(len(want.Up), len(want.otherUp))
	for i := range n {
		if want.Up[i].Type == want.otherUp[i].Type && bytes.Equal(want.Up[i].Body, want.otherUp[i].Body) {
			continue
		}
		rep.Informative = append(rep.Informative, Finding{
			Rule:       RuleInputAgreement,
			ClientPort: want.ClientPort,
			Index:      i,
			Detail:     "the two sides of the capture disagree about what the client sent, so this capture cannot decide anything",
			Ours:       want.Up[i].String(),
			Theirs:     want.otherUp[i].String(),
		})
		return
	}
	if len(want.Up) != len(want.otherUp) {
		rep.Informative = append(rep.Informative, Finding{
			Rule:       RuleInputAgreement,
			ClientPort: want.ClientPort,
			Index:      n,
			Detail:     fmt.Sprintf("the two sides of the capture saw %d and %d client messages", len(want.Up), len(want.otherUp)),
		})
	}
}

// compareConn runs the must-match rules and then the informative ones for one
// connection. The must-match rules each read a filtered stream, so a difference
// in one message type cannot hide a difference in another by shifting positions.
func compareConn(rep *Report, want, got *Conn, match *matchIndex, content *matchContent) {
	if want.Transport != got.Transport {
		rep.Informative = append(rep.Informative, Finding{
			Rule:       RuleFraming,
			ClientPort: want.ClientPort,
			Detail:     fmt.Sprintf("the capture carried this connection as %s and the replay framed it as %s", want.Transport, got.Transport),
		})
	}

	upgradeResponse(rep, want, got)
	startObserveStatus(rep, want, got)
	activateObserver(rep, want, got)
	replayPageOrder(rep, want, got, match)
	replayContent(rep, want, got, content)
	sequenceAndBodies(rep, want, got)
	counts(rep, want, got)
	pendingPeers(rep, want, got)
	tickCadence(rep, want, got)
	messageTail(rep, want, got)
}

// lastAt is when a stream stopped, which for two streams of different lengths is
// the first question and not a detail: an extra message at the end and an extra
// message in the middle both make the totals differ, and only the times say
// which one it is.
func lastAt(msgs []*Msg, zero time.Time) Elapsed {
	for i := len(msgs) - 1; i >= 0; i-- {
		if e := elapsedOf(msgs[i], zero); e.Valid {
			return e
		}
	}
	return Elapsed{}
}

// spanOf is the first and last moment of a stream, which is what a count
// finding has to say: a type whose totals differ is only located by where its
// messages start and stop.
func spanOf(msgs []*Msg, zero time.Time) (first, last Elapsed) {
	first = Elapsed{}
	for _, m := range msgs {
		if e := elapsedOf(m, zero); e.Valid {
			first = e
			break
		}
	}
	return first, lastAt(msgs, zero)
}

// startObserveStatus compares the sequence of 0x19 status bytes. Must-match: the
// client reads this byte to decide whether to go on to ActivateObserver, and a
// 3 there is what made _30 report a match with no start-of-game data.
func startObserveStatus(rep *Report, want, got *Conn) {
	ours := filtered(got.Down, protocol.MsgStartObserveResp)
	theirs := filtered(want.Down, protocol.MsgStartObserveResp)
	for i := range min(len(ours), len(theirs)) {
		if statusOf(ours[i]) == statusOf(theirs[i]) {
			continue
		}
		rep.mustMatch(Finding{
			Rule:       RuleStartObserveStatus,
			ClientPort: want.ClientPort,
			Index:      i,
			Detail: fmt.Sprintf("StartObserveResponse #%d: the original answered %d and we answered %d",
				i, statusOf(theirs[i]), statusOf(ours[i])),
			Ours:     ours[i].String(),
			Theirs:   theirs[i].String(),
			AtOurs:   elapsedOf(ours[i], rep.started),
			AtTheirs: elapsedOf(theirs[i], rep.origin),
		})
	}
	if len(ours) > len(theirs) {
		rep.mustMatch(Finding{
			Rule:       RuleStartObserveStatus,
			ClientPort: want.ClientPort,
			Index:      len(theirs),
			Detail: fmt.Sprintf("we answered StartObserveResponse %d time(s) and the original %d",
				len(ours), len(theirs)),
			Ours:   ours[len(theirs)].String(),
			Theirs: "",
			AtOurs: elapsedOf(ours[len(theirs)], rep.started),
		})
	}
	if len(ours) < len(theirs) {
		// Silence is not a safe answer to a request the client is blocked on: it
		// waits for this byte, and _30 is what staying silent costs. The rule
		// cannot stop at the first byte that differs either, so both directions
		// of a missing answer are must-match.
		rep.mustMatch(Finding{
			Rule:       RuleStartObserveStatus,
			ClientPort: want.ClientPort,
			Index:      len(ours),
			Detail: fmt.Sprintf("the original answered StartObserveResponse %d time(s) and we answered %d: the client is blocked on this byte",
				len(theirs), len(ours)),
			Ours:     "",
			Theirs:   theirs[len(ours)].String(),
			AtTheirs: elapsedOf(theirs[len(ours)], rep.origin),
		})
	}
}

func statusOf(m *Msg) int {
	if len(m.Body) == 0 {
		return -1
	}
	return int(m.Body[0])
}

// activateObserver compares whether a 0x1E arrives at all, and its status.
// Must-match: without it the observer never starts receiving the replay, which
// is exactly where _30 died.
func activateObserver(rep *Report, want, got *Conn) {
	ours := filtered(got.Down, protocol.MsgActivateObserverResp)
	theirs := filtered(want.Down, protocol.MsgActivateObserverResp)
	switch {
	case len(ours) == 0 && len(theirs) > 0:
		rep.mustMatch(Finding{
			Rule:       RuleActivateObserverResponse,
			ClientPort: want.ClientPort,
			Detail: fmt.Sprintf("the original answered ActivateObserver %d time(s) and we never did: the observer never starts, and the client then reports that the start of the match is missing",
				len(theirs)),
			Ours:     "",
			Theirs:   theirs[0].String(),
			AtTheirs: elapsedOf(theirs[0], rep.origin),
		})
	case len(ours) > 0 && len(theirs) == 0:
		rep.mustMatch(Finding{
			Rule:       RuleActivateObserverResponse,
			ClientPort: want.ClientPort,
			Detail:     "we answered ActivateObserver and the original never did",
			Ours:       ours[0].String(),
			Theirs:     "",
			AtOurs:     elapsedOf(ours[0], rep.started),
		})
	case len(ours) != len(theirs):
		rep.mustMatch(Finding{
			Rule:       RuleActivateObserverResponse,
			ClientPort: want.ClientPort,
			Detail:     fmt.Sprintf("ActivateObserverResponse: the original sent %d, we sent %d", len(theirs), len(ours)),
			Ours:       firstString(ours),
			Theirs:     firstString(theirs),
			AtOurs:     firstElapsed(ours, rep.started),
			AtTheirs:   firstElapsed(theirs, rep.origin),
		})
	}
	for i := range min(len(ours), len(theirs)) {
		if statusOf(ours[i]) == statusOf(theirs[i]) {
			continue
		}
		rep.mustMatch(Finding{
			Rule:       RuleActivateObserverResponse,
			ClientPort: want.ClientPort,
			Index:      i,
			Detail: fmt.Sprintf("ActivateObserverResponse #%d: the original answered %d and we answered %d",
				i, statusOf(theirs[i]), statusOf(ours[i])),
			Ours:     ours[i].String(),
			Theirs:   theirs[i].String(),
			AtOurs:   elapsedOf(ours[i], rep.started),
			AtTheirs: elapsedOf(theirs[i], rep.origin),
		})
	}
}

func firstString(msgs []*Msg) string {
	if len(msgs) == 0 {
		return ""
	}
	return msgs[0].String()
}

// firstElapsed is the moment of a stream's first message, for a finding that is
// about a whole stream rather than about one message in it.
func firstElapsed(msgs []*Msg, zero time.Time) Elapsed {
	first, _ := spanOf(msgs, zero)
	return first
}

// replayPageOrder checks the replay stream an observer is served: which point of
// the match its first frame is, and the order of its frames.
//
// The stream is the session's 0x0A ack frames, not the 0x09 the players upload.
// The original's two observer connections in 2026-09-29_012528_03cb67c7cf19
// received 1346 and 800 messages, every one of them a byte-identical copy of a
// message of the player's 0x0A stream and in strictly increasing order of it,
// from ack #0 (the match's first ack) up to ack #1345 and ack #799. Not one 0x09.
//
// Two things about the window are must-match, and both are read as frame identity
// against the match's own ack stream, not as bytes: the frames carry run-specific
// data, so what is comparable is each frame's index in that stream.
//
//   - the sequence. A reversed pair crashed the client in capture _5, which is why
//     TestGameDataAckOrder exists, so a stream that does not rise is a must-match
//     on its own, whatever the other side did.
//   - the alignment with the original. Two streams of the same length that are
//     the same frames in the same order are the same answer; a position where they
//     name different frames is not, and the case that matters is the first one:
//     the second observer of 2026-09-29_012528_03cb67c7cf19 was being put at
//     frame 8 of the match instead of frame 0, which is why it was served 70
//     messages where the original served 803.
//
// How MANY frames the observer got is informative, not must-match: where the
// capture was cut and where the replay's window ends both move the tail, so a
// count difference there is not evidence of anything. What is not left out is
// where each side's window started and ended, which is in the informative finding
// under every outcome.
func replayPageOrder(rep *Report, want, got *Conn, match *matchIndex) {
	if !requestedObservation(want) {
		// A player is sent the 0x0A ack stream directly, not out of the store, so
		// there is no observer window to talk about on this connection.
		return
	}
	ours, theirs := replayStream(got.Down), replayStream(want.Down)
	window := match.windowLines(ours, theirs)

	switch {
	case len(ours) == 0 && len(theirs) == 0:
		rep.Informative = append(rep.Informative, Finding{
			Rule:       RuleReplayPageOrder,
			ClientPort: want.ClientPort,
			Detail:     "neither side served this observer a single replay frame",
		})
		return
	case len(ours) == 0:
		rep.mustMatch(Finding{
			Rule:       RuleReplayPageOrder,
			ClientPort: want.ClientPort,
			Detail: fmt.Sprintf("the original served this observer %d replay frames and we served none: "+
				"the client presses observe and sits there waiting for the game to start", len(theirs)),
			Theirs: theirs[0].String(),
			Lines:  window,
		})
		return
	case len(theirs) == 0:
		rep.Informative = append(rep.Informative, Finding{
			Rule:       RuleReplayPageOrder,
			ClientPort: want.ClientPort,
			Detail:     fmt.Sprintf("we served this observer %d replay frames and the original served none", len(ours)),
			Ours:       ours[0].String(),
			Lines:      window,
		})
		return
	}

	ourIdx, ourOK := match.indexOf(ours)
	theirIdx, theirOK := match.indexOf(theirs)
	if !ourOK || !theirOK {
		rep.Informative = append(rep.Informative, Finding{
			Rule:       RuleReplayPageOrder,
			ClientPort: want.ClientPort,
			Detail: "the replay stream could not be located in the match's own ack stream, " +
				"so no order or position claim is made about it: a frame that identifies against neither " +
				"side's player stream would only be placed by a guess",
			Lines: window,
		})
		return
	}

	for i := 1; i < len(ourIdx); i++ {
		if ourIdx[i] > ourIdx[i-1] {
			continue
		}
		rep.mustMatch(Finding{
			Rule:       RuleReplayPageOrder,
			ClientPort: want.ClientPort,
			Index:      i,
			Detail: fmt.Sprintf("replay frame #%d is match frame %d, and the one before it was %d: "+
				"the stream has to rise (capture _5 crashed the client on a reversed pair)",
				i, ourIdx[i], ourIdx[i-1]),
			Ours:   ours[i].String(),
			Theirs: theirs[min(i, len(theirs)-1)].String(),
			AtOurs: elapsedOf(ours[i], rep.started),
		})
		return
	}

	for i := range min(len(ourIdx), len(theirIdx)) {
		if ourIdx[i] == theirIdx[i] {
			continue
		}
		detail := fmt.Sprintf("replay frame #%d is match frame %d and the original served match frame %d in that position: "+
			"the two observers are being put at different points of the match", i, ourIdx[i], theirIdx[i])
		if i == 0 {
			detail = fmt.Sprintf("we put this observer at match frame %d and the original put it at frame %d: "+
				"an observer that starts later is an observer with no start-of-match data, which is the _30 symptom",
				ourIdx[0], theirIdx[0])
		}
		rep.mustMatch(Finding{
			Rule:       RuleReplayPageOrder,
			ClientPort: want.ClientPort,
			Index:      i,
			Detail:     detail,
			Ours:       ours[i].String(),
			Theirs:     theirs[i].String(),
			AtOurs:     elapsedOf(ours[i], rep.started),
			AtTheirs:   elapsedOf(theirs[i], rep.origin),
			Lines:      window,
		})
		return
	}

	if len(ours) != len(theirs) {
		firstOurs, lastOurs := spanOf(ours, rep.started)
		firstTheirs, lastTheirs := spanOf(theirs, rep.origin)
		rep.Informative = append(rep.Informative, Finding{
			Rule:       RuleReplayPageOrder,
			ClientPort: want.ClientPort,
			Detail: fmt.Sprintf("replay stream: the original served %d frames and we served %d, "+
				"the same frames of the match in the same order", len(theirs), len(ours)),
			AtOurs:   lastOurs,
			AtTheirs: lastTheirs,
			Lines: append(window,
				fmt.Sprintf("window: ours from t=%s to t=%s | original from t=%s to t=%s",
					firstOurs, lastOurs, firstTheirs, lastTheirs)),
		})
	}
}

// requestedObservation reports whether the client asked to observe, which is what
// makes a connection an observer: it sent a StartObserveRequest (0x18 or 0x1C,
// and the dispatcher handles both).
func requestedObservation(c *Conn) bool {
	for _, m := range c.Up {
		if m.Type == 0x18 || m.Type == 0x1C {
			return true
		}
	}
	return false
}

// IsObserver is requestedObservation, for the tools outside this package that have
// to tell a player connection from an observer one to read a report: a capture's
// observer is the connection that asked to observe, and nothing else.
func IsObserver(c *Conn) bool { return requestedObservation(c) }

// replayStream is the stream of frames an observer is served: the session's 0x0A
// acks, which is what the frame store holds and what the original puts on that
// connection.
func replayStream(msgs []*Msg) []*Msg { return filtered(msgs, protocol.MsgEndOfGameData) }

// matchIndex locates frames in the match's own ack stream.
//
// A frame's identity is its body, and the body of a 0x0A ack repeats (the
// six-byte heartbeat is five hundred of them in a row), so a body alone does not
// say WHICH frame of the match it is. The index is therefore resolved against
// the player's own 0x0A stream of the same side, taking the first candidate at or
// after the previous one: that is what makes a run of identical heartbeats
// resolve to consecutive indices instead of all to the same one.
//
// When the bodies do not identify, nothing is claimed. A rule that guessed here
// would be reporting an alignment it invented, which is the failure mode
// SESSION-JOURNAL.md records for the 0x0A extras.
type matchIndex struct {
	ours   map[string][]int
	theirs map[string][]int
	oursN  int
	theirN int
}

// buildMatchIndex finds the player connection and indexes the 0x0A acks each side
// sent it. A connection is the player's when it did not ask to observe; the first
// one that did not is taken, and a capture with no such connection produces an
// empty index, which makes every order claim fall back to "not located".
func buildMatchIndex(conns []*Conn, ours map[int]*Conn) *matchIndex {
	m := &matchIndex{ours: map[string][]int{}, theirs: map[string][]int{}}
	for _, c := range conns {
		if requestedObservation(c) {
			continue
		}
		for _, side := range []struct {
			conn   *Conn
			into   map[string][]int
			isOurs bool
		}{
			{conn: c, into: m.theirs, isOurs: false},
			{conn: ours[c.ClientPort], into: m.ours, isOurs: true},
		} {
			if side.conn == nil {
				continue
			}
			for i, msg := range replayStream(side.conn.Down) {
				side.into[string(msg.Body)] = append(side.into[string(msg.Body)], i)
			}
			if side.isOurs {
				m.oursN = len(replayStream(side.conn.Down))
			} else {
				m.theirN = len(replayStream(side.conn.Down))
			}
		}
		break
	}
	return m
}

// buildMatchContent is buildMatchIndex read by content: the match's own ordered
// sub-message sequence on each side, so a replay window's first sub can be named
// as a match sub index. A window that starts on the match's sub 600 and one that
// starts on sub 0 differ in the game the observer reconstructs, and no amount of
// re-grouping the acks can hide that.
//
// The player connection is found the same way as in buildMatchIndex: the first
// connection that did not ask to observe, per side.
func buildMatchContent(conns []*Conn, ours map[int]*Conn) *matchContent {
	m := &matchContent{}
	for _, c := range conns {
		if requestedObservation(c) {
			continue
		}
		m.theirs = readAcks(replayStream(c.Down)).subs
		if o := ours[c.ClientPort]; o != nil {
			m.ours = readAcks(replayStream(o.Down)).subs
		}
		break
	}
	return m
}

// indexOf turns a replay stream into frame indices, and reports whether every
// frame identified. An unidentifiable frame is a fact about the capture or about
// the harness, not about the order, so the caller is told instead of being given
// a guess.
//
// The scan is monotone: the first candidate strictly after the previous index,
// which is what resolves a run of identical heartbeats to consecutive indices
// instead of all of them to the same one.
//
// When there is no candidate after the previous index the frame cannot advance,
// and the two things that used to be done here were both wrong. It took cands[0],
// the EARLIEST occurrence of that body in the match: with a body that repeats (the
// heartbeat does, five hundred times) that is the far end of a long list, so a
// window that legitimately starts late in the match was reported as falling off
// the back of it, and the rule read that as a reversed pair. And it then made that
// guess the cursor, so prev moved BACKWARDS and every frame after it was aligned
// against a position that no frame of the stream had reached: one unplaceable
// frame misaligned the whole rest of the window.
//
// What it does instead: report the frame at the NEAREST candidate at or below the
// cursor (the smallest claim the bytes support) and leave the cursor where it was.
// The reversal rule still fires, on the honest index, because a frame that cannot
// advance is a frame that does not rise whatever index it is given; what stops is
// the cascade, and with it the invented positions in the positional comparison
// against the original. This is the open flake SESSION-JOURNAL.md records: the
// window starting on match frame 503 instead of 0 after the store recycled.
func (m *matchIndex) indexOf(stream []*Msg) ([]int, bool) {
	if len(stream) == 0 {
		return nil, false
	}
	out := make([]int, 0, len(stream))
	prev := -1
	ourSide := m.ours
	if !m.identified(stream, ourSide) {
		ourSide = m.theirs
	}
	for _, msg := range stream {
		cands := ourSide[string(msg.Body)]
		if len(cands) == 0 {
			return nil, false
		}
		pick := -1
		for _, cand := range cands {
			if cand > prev {
				pick = cand
				break
			}
		}
		if pick < 0 {
			// No candidate advances: report the closest one that does not, and
			// keep the cursor. cands is ascending (buildMatchIndex appends in
			// stream order), so the last entry is the nearest one at or below
			// the cursor.
			pick = cands[len(cands)-1]
		}
		out = append(out, pick)
		if pick > prev {
			prev = pick
		}
	}
	return out, true
}

func (m *matchIndex) identified(stream []*Msg, idx map[string][]int) bool {
	if len(idx) == 0 {
		return false
	}
	for _, msg := range stream {
		if len(idx[string(msg.Body)]) == 0 {
			return false
		}
	}
	return true
}

// windowLines says where in the match each side's window starts and ends, which
// is the number a count alone cannot give: 800 frames served from frame 0 and 800
// served from frame 600 are the same count and not the same answer.
func (m *matchIndex) windowLines(ours, theirs []*Msg) []string {
	ourIdx, _ := m.indexOf(ours)
	theirIdx, _ := m.indexOf(theirs)
	span := func(idx []int, total int) string {
		if len(idx) == 0 {
			return "none"
		}
		return fmt.Sprintf("[%d..%d] of %d", idx[0], idx[len(idx)-1], total)
	}
	ourTotal, theirTotal := m.oursN, m.theirN
	if ourTotal == 0 {
		ourTotal = m.theirN
	}
	if theirTotal == 0 {
		theirTotal = m.oursN
	}
	return []string{fmt.Sprintf("match frames served: ours %s | original %s",
		span(ourIdx, ourTotal), span(theirIdx, theirTotal))}
}

// normalise returns a copy of the body with every masked range filled in, which
// is what the byte-for-byte comparison reads.
func normalise(m *Msg) []byte {
	out := append([]byte(nil), m.Body...)
	for _, f := range maskedFields {
		if f.Type != m.Type {
			continue
		}
		for i := f.From; i < f.To && i < len(out); i++ {
			out[i] = maskFill
		}
	}
	return out
}

// sequenceAndBodies reports the whole stream: the type sequence position by
// position, and any body that differs at an aligned position for a type no
// must-match rule already covers byte for byte. Informative, because the 0x0A
// tick cadence and the original's own inconsistencies live here. 0x1E is
// deliberately not in the covered set: its must-match rule reads the status
// byte only, so the rest of that body is compared here, with the start time
// masked out.
//
// Every finding carries the moment each side said it, because "the two streams
// diverge at message #17" says nothing about whether the difference is a
// message at the end or a message through the middle, and the byte comparison
// has to be the positional one to keep comparing the same message with itself.
// The alignment that locates an extra message in time is align(), used by the
// counts rule: reusing it here would pair two different messages after the first
// insertion and report their bodies as a mismatch, which is a different finding
// from the one being asked about.
//
// That is also why there is one body finding per TYPE and not one per position.
// The comparison is positional, so once the two streams have different lengths
// every position past the first insertion pairs a different message and reports a
// mismatch that is only a consequence of the offset. Printing six hundred of them
// buries the one that locates the difference, which is the whole point of the
// finding: the first one of each type is reported with both frames, and the total
// goes on its Lines so the report still says how many there were.
func sequenceAndBodies(rep *Report, want, got *Conn) {
	covered := map[byte]bool{
		protocol.MsgStartObserveResp: true,
	}
	// per type: how many positions differed, and where its first finding landed in
	// rep.Informative so the count can be added to it.
	mismatches := map[byte]int{}
	firstAt := map[byte]int{}
	for i := range min(len(want.Down), len(got.Down)) {
		theirs, ours := want.Down[i], got.Down[i]
		if theirs.Type != ours.Type {
			rep.Informative = append(rep.Informative, Finding{
				Rule:       RuleMessageSequence,
				ClientPort: want.ClientPort,
				Index:      i,
				Detail:     fmt.Sprintf("message #%d: the original sent %s and we sent %s", i, theirs, ours),
				Ours:       ours.String(),
				Theirs:     theirs.String(),
				AtOurs:     elapsedOf(ours, rep.started),
				AtTheirs:   elapsedOf(theirs, rep.origin),
			})
			break
		}
		if covered[ours.Type] {
			continue
		}
		if bytes.Equal(normalise(ours), normalise(theirs)) {
			continue
		}
		mismatches[ours.Type]++
		if _, reported := firstAt[ours.Type]; reported {
			continue
		}
		firstAt[ours.Type] = len(rep.Informative)
		rep.Informative = append(rep.Informative, Finding{
			Rule:       RuleBodyMismatch,
			ClientPort: want.ClientPort,
			Index:      i,
			Detail:     fmt.Sprintf("message #%d (%s): the bodies differ", i, ours),
			Ours:       ours.String(),
			Theirs:     theirs.String(),
			AtOurs:     elapsedOf(ours, rep.started),
			AtTheirs:   elapsedOf(theirs, rep.origin),
		})
	}
	for typ, n := range mismatches {
		if n < 2 {
			continue
		}
		idx := firstAt[typ]
		rep.Informative[idx].Lines = append(rep.Informative[idx].Lines, fmt.Sprintf(
			"%d position(s) of this connection have a different %s body. Only the first is shown: the two "+
				"streams have different lengths, so every position past the first insertion compares a "+
				"different message and the rest follow from the offset rather than from a difference of their own",
			n, protocol.MessageTypeName(typ)))
	}
	if len(want.Down) != len(got.Down) {
		oursLast, theirsLast := lastAt(got.Down, rep.started), lastAt(want.Down, rep.origin)
		f := Finding{
			Rule:       RuleMessageSequence,
			ClientPort: want.ClientPort,
			Index:      min(len(want.Down), len(got.Down)),
			Detail: fmt.Sprintf("the original sent %d messages and we sent %d",
				len(want.Down), len(got.Down)),
			AtOurs:   oursLast,
			AtTheirs: theirsLast,
		}
		// The totals are the same two numbers either way, so the report says
		// outright that this finding is not where the question is answered: the
		// [message-tail] finding below splits them by time.
		f.Lines = append(f.Lines, fmt.Sprintf(
			"the totals say nothing about where these %+d message(s) are; the [message-tail] finding below splits them by time",
			len(got.Down)-len(want.Down)))
		rep.Informative = append(rep.Informative, f)
	}
}

// counts reports how many messages of each type each side sent. It is the view
// that makes a structural difference visible at a glance, even when the order is
// right, and it now says where each type's difference sits: the span of the type
// on both sides, the split around the settle window, what the payload counts
// account for, and the seconds where the two counts disagree.
//
// Those are not decoration. A type of byte-identical messages (the 6-byte tick
// ack is exactly that, five hundred of them in a row) cannot be lined up one by
// one at all, and a line-up that guesses would put a time on something nobody can
// locate. See the alignment note in when.go: the counts are what is evidence.
func counts(rep *Report, want, got *Conn) {
	ours := histogram(got.Down)
	theirs := histogram(want.Down)
	types := make([]byte, 0, len(ours)+len(theirs))
	for t := range ours {
		types = append(types, t)
	}
	for t := range theirs {
		if _, ok := ours[t]; !ok {
			types = append(types, t)
		}
	}
	sort.Slice(types, func(i, j int) bool { return types[i] < types[j] })
	for _, t := range types {
		if ours[t] == theirs[t] {
			continue
		}
		rep.Informative = append(rep.Informative, Finding{
			Rule:       RuleMessageCount,
			ClientPort: want.ClientPort,
			Detail: fmt.Sprintf("%s: the original sent %d, we sent %d",
				protocol.MessageTypeName(t), theirs[t], ours[t]),
			Lines: countTiming(rep, want, got, t),
		})
	}
}

// countTiming is the where of one type's count difference: the span of the type
// on both sides, the split around the settle window, what the payload counts
// account for, and the seconds where the two counts disagree.
func countTiming(rep *Report, want, got *Conn, t byte) []string {
	oursMsgs, theirsMsgs := filtered(got.Down, t), filtered(want.Down, t)
	oursAxis, theirsAxis := onAxis(oursMsgs, rep.started), onAxis(theirsMsgs, rep.origin)
	firstOurs, lastOurs := spanOf(oursMsgs, rep.started)
	firstTheirs, lastTheirs := spanOf(theirsMsgs, rep.origin)
	if !firstOurs.Valid && !firstTheirs.Valid {
		return []string{"when: the capture carries no timestamp for these messages, so where the difference sits cannot be said"}
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("when: ours from t=%s to t=%s | original from t=%s to t=%s",
		firstOurs, lastOurs, firstTheirs, lastTheirs))

	oursIn, oursBefore := 0, 0
	for _, e := range oursAxis {
		if !e.Valid {
			continue
		}
		if rep.window.after(e) {
			oursIn++
		} else {
			oursBefore++
		}
	}
	theirsIn, theirsBefore := 0, 0
	for _, e := range theirsAxis {
		if !e.Valid {
			continue
		}
		if rep.window.after(e) {
			theirsIn++
		} else {
			theirsBefore++
		}
	}
	lines = append(lines, fmt.Sprintf(
		"window %s: %d of ours at or after it opened and %d before it, against %d of the original's and %d",
		rep.window, oursIn, oursBefore, theirsIn, theirsBefore))

	// The payload counts, which is what can be said about which messages are the
	// extra ones without trusting an order the two streams may not share. It only
	// localises anything when the payloads line up: a type whose payloads carry
	// run-specific bytes (the real tick acks do) is a different finding, and the
	// report says that instead of quoting a surplus nobody can use.
	extra := len(oursMsgs) - len(theirsMsgs)
	over, under := payloadDiff(oursMsgs, theirsMsgs, rep.started)
	switch {
	case len(over) == 0:
		lines = append(lines, fmt.Sprintf(
			"not placeable by payload: we sent no payload of this type more often than the original did, so the %d extra are spread "+
				"over payloads the two sides sent the same number of times and the bytes alone cannot say which are ours", abs(extra)))
	case under == 0 && total(over) == abs(extra):
		shown, dropped := over, 0
		if len(shown) > maxSurplusListed {
			dropped, shown = len(shown)-maxSurplusListed, shown[:maxSurplusListed]
		}
		parts := make([]string, 0, len(shown))
		for _, s := range shown {
			parts = append(parts, s.String())
		}
		line := fmt.Sprintf("by payload: we sent these %d payload(s) more often than the original, which accounts for all %d extra: %s",
			len(shown), abs(extra), strings.Join(parts, "; "))
		if dropped > 0 {
			line += fmt.Sprintf(" (+%d more payload(s), the same surplus split among them)", dropped)
		}
		lines = append(lines, line)
	default:
		lines = append(lines, fmt.Sprintf(
			"not placeable by payload: the payloads of this type are not the same bytes on the two sides (we sent %d more copies of "+
				"some and %d fewer of others than the original, so the surplus adds up to %d and not to the %d extra). Which %d of ours "+
				"the original did not send is not something the bytes can say. The seconds below are the evidence",
			total(over), under, total(over), abs(extra), abs(extra)))
	}
	if b := secondBuckets(oursAxis, theirsAxis, maxSecondBuckets); b != "" {
		lines = append(lines, "per second, original|ours, only the seconds that differ: "+b)
		lines = append(lines, "a differing second says where the difference is and not yet that it is a count: the two sides' traffic "+
			"can be up to a second apart, because the original's stamps are whole seconds and ours are the replay's clock")
	}
	return lines
}

// renderTimes lists moments, keeping the list readable when there are many: the
// first few and the last few with the count of what is in between, so a
// conclusion that depends on where the extremes are can still be read.
func renderTimes(in []Elapsed, max int) string {
	if len(in) <= max {
		return joinTimes(in)
	}
	half := max / 2
	return joinTimes(in[:half]) + fmt.Sprintf(" ... and %d more ... ", len(in)-max) + joinTimes(in[len(in)-half:])
}

func joinTimes(in []Elapsed) string {
	parts := make([]string, 0, len(in))
	for _, e := range in {
		parts = append(parts, "t="+e.String())
	}
	return strings.Join(parts, ", ")
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func histogram(msgs []*Msg) map[byte]int {
	out := make(map[byte]int)
	for _, m := range msgs {
		out[m.Type]++
	}
	return out
}

// messageTail answers the question the totals cannot: the messages we sent and
// the original did not are either produced during the match, or produced during
// the settle window the harness keeps reading for after the last captured input.
//
// A total of "eleven more" looks identical either way, which is why this is a
// rule and not a remark. It is informative, not must-match, and that is not a
// judgement about the two explanations: which of them it is depends on the
// harness's own settle and not on anything the client sees, and the settle is a
// harness choice, so a finding that could be an artifact of the harness must not
// be able to fail a test. What it does do is make the two impossible to confuse
// when somebody reads the report, and that is what the rule is for.
//
// It is emitted only when the two streams are not the same length, because with
// the same number of messages there is no extra message to place.
func messageTail(rep *Report, want, got *Conn) {
	v := judgeTail(onAxis(got.Down, rep.started), onAxis(want.Down, rep.origin), rep.window)
	if v.Extra == 0 {
		return
	}
	lines := []string{v.Reason}
	oursLast, theirsLast := lastAt(got.Down, rep.started), lastAt(want.Down, rep.origin)
	lines = append(lines, fmt.Sprintf("our last message at t=%s, the original's at t=%s: %d of our %d fall at or after the window opened, against %d of its %d",
		oursLast, theirsLast, v.InWindow, v.Ours, v.TheirsAfter, v.Theirs))
	if theirsLast.Valid {
		lines = append(lines, fmt.Sprintf("the original's stamps are whole seconds, so its last message really happened somewhere inside the second t=%s and "+
			"whether that instant was inside our %s is not something a whole-second stamp can say", theirsLast, rep.window))
	}
	lines = append(lines, fmt.Sprintf("the split is exact: %d before the window plus %d at or after it is the %d extra",
		v.ExtrasBeforeWindow, v.Extra-v.ExtrasBeforeWindow, v.Extra))
	rep.Informative = append(rep.Informative, Finding{
		Rule:       RuleMessageTail,
		ClientPort: want.ClientPort,
		Detail: fmt.Sprintf("we sent %d messages and the original %d, and this is where the %d extra ones are",
			v.Ours, v.Theirs, abs(v.Extra)),
		AtOurs:   oursLast,
		AtTheirs: theirsLast,
		Lines:    lines,
	})
}
