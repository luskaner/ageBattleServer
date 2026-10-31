package sim

import (
	"strings"
	"testing"

	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// The rules of replay-content, driven directly.
//
// 2026-09-30_010531_a553ef8fd63c is the reason this file exists. The report said
// "the replay stream could not be located in the match's own ack stream" for the
// original, on a connection where the original was serving 5363 replay frames, so
// the one question the capture was taken to answer — does the observer get the
// right game data — was left with no answer at all. The whole-frame comparison
// cannot answer it once the two sides group the same acks differently, and these
// tests pin the rule that can.

// ack builds one ack body in the shape decodeAck reads: [0xfe][sub count] + subs
// + [pending count][pending slots] + [u16 LE len] block. It is written out here
// rather than taken from the code under test, because a fixture built from the
// implementation cannot fail when the implementation is wrong.
//
// Each sub is a slot and a payload, and the payload carries its own u16 LE length
// because that is what the client uploaded and the server queues verbatim.
func ack(subs ...[]byte) []byte {
	out := []byte{ackMarker, byte(len(subs))}
	for _, p := range subs {
		out = append(out, 1, byte(len(p)), byte(len(p)>>8))
		out = append(out, p...)
	}
	return append(out, 0x00, 0x00, 0x00)
}

// ackWithEntry is ack plus a block carrying one checksum entry, the 8 bytes
// [crc u32 LE][page u32 LE] the reconciliation appends.
func ackWithEntry(sub []byte, crc, page uint32) []byte {
	out := ack(sub)
	// Replace the empty block at the tail with a real one: 3 + 8 bytes.
	out = out[:len(out)-2]
	out = append(out, 0x0B, 0x00, 0x02, 0x01, 0x00)
	for i := range 4 {
		out = append(out, byte(crc>>(8*uint(i))))
	}
	for i := range 4 {
		out = append(out, byte(page>>(8*uint(i))))
	}
	return out
}

func ackMsg(body []byte) *Msg {
	return &Msg{Type: protocol.MsgEndOfGameData, Body: body, Framing: FramingWS}
}

// observerConn is a connection that asked to observe, which is what makes it an
// observer for the rules.
func observerConn(down ...[]byte) *Conn {
	c := &Conn{ClientPort: 5002, Transport: FramingWS}
	c.Up = []*Msg{{Type: 0x1C, Body: make([]byte, 24)}}
	for _, b := range down {
		c.Down = append(c.Down, ackMsg(b))
	}
	return c
}

func playerConn(down ...[]byte) *Conn {
	c := &Conn{ClientPort: 5001, Transport: FramingWS}
	for _, b := range down {
		c.Down = append(c.Down, ackMsg(b))
	}
	return c
}

func contentOf(t *testing.T, match ...[]byte) *matchContent {
	t.Helper()
	return &matchContent{ours: readAcks(acksOf(match)).subs, theirs: readAcks(acksOf(match)).subs}
}

func acksOf(bodies [][]byte) []*Msg {
	out := make([]*Msg, 0, len(bodies))
	for _, b := range bodies {
		out = append(out, ackMsg(b))
	}
	return out
}

func findingFor(rep *Report, rule string) *Finding {
	for i := range rep.MustMatch {
		if rep.MustMatch[i].Rule == rule {
			return &rep.MustMatch[i]
		}
	}
	return nil
}

func informativeFor(rep *Report, rule string) *Finding {
	for i := range rep.Informative {
		if rep.Informative[i].Rule == rule {
			return &rep.Informative[i]
		}
	}
	return nil
}

// TestReplayContentPassesWhenTheGroupingDiffers is the property the real capture
// rests on: the same subs in a different number of acks is the same game, and the
// rule has to say so instead of failing or, worse, claiming nothing.
func TestReplayContentPassesWhenTheGroupingDiffers(t *testing.T) {
	// The match: three acks carrying three uploads.
	match := [][]byte{ack([]byte{0xAA}), ack([]byte{0xBB}), ack([]byte{0xCC})}
	// The original served one sub per ack. We packed the first two together and
	// let the third ride alone, which is a different number of frames and the
	// same 0x0A bytes' worth of game data.
	theirs := observerConn(match[0], match[1], match[2])
	ours := observerConn(ack([]byte{0xAA}, []byte{0xBB}), match[2])

	rep := &Report{}
	replayContent(rep, theirs, ours, contentOf(t, match...))

	if !rep.OK() {
		t.Fatalf("re-grouping the acks is not a divergence, got %v", mustMatchNames(rep))
	}
	f := informativeFor(rep, RuleReplayContent)
	if f == nil {
		t.Fatal("the rule must report what it compared, or a reader cannot tell an empty comparison from a passing one")
	}
	for _, want := range []string{
		"ack frames: ours 2 | original 3",
		"game-data subs served: ours 3 | original 3",
		"window starts at: ours sub 0 of 3 | original sub 0 of 3",
		"acks carrying subs, by count: ours 1x1 1x2 | original 3x1",
	} {
		if !strings.Contains(strings.Join(f.Lines, "\n"), want) {
			t.Errorf("the report must say %q, got %v", want, f.Lines)
		}
	}
}

// TestReplayContentDetectsADroppedSub: a sub the original served and we did not is
// the observer reconstructing a different game, whatever the framing is.
func TestReplayContentDetectsADroppedSub(t *testing.T) {
	match := [][]byte{ack([]byte{0xAA}), ack([]byte{0xBB}), ack([]byte{0xCC})}
	theirs := observerConn(match[0], match[1], match[2])
	ours := observerConn(match[0], match[2]) // BB never made it

	rep := &Report{}
	replayContent(rep, theirs, ours, contentOf(t, match...))

	f := findingFor(rep, RuleReplayContent)
	if f == nil {
		t.Fatalf("a dropped sub must be a must-match divergence, got %v", mustMatchNames(rep))
	}
	if f.Index != 1 {
		t.Errorf("the finding must name the position of the missing sub, got #%d", f.Index)
	}
	if !strings.Contains(f.Ours, "cc") || !strings.Contains(f.Theirs, "bb") {
		t.Errorf("the finding must show the two sub-messages, got ours=%q theirs=%q", f.Ours, f.Theirs)
	}
	t.Log("\n" + f.Detail)
}

// TestReplayContentDetectsAReorderedSub is the same defect the other way round: a
// reversed pair crashed the client in capture _5, and the body comparison could
// only see it while the framing matched.
func TestReplayContentDetectsAReorderedSub(t *testing.T) {
	match := [][]byte{ack([]byte{0xAA}), ack([]byte{0xBB})}
	theirs := observerConn(match[0], match[1])
	ours := observerConn(match[1], match[0])

	rep := &Report{}
	replayContent(rep, theirs, ours, contentOf(t, match...))

	if findingFor(rep, RuleReplayContent) == nil {
		t.Fatalf("a reversed pair must be a must-match divergence, got %v", mustMatchNames(rep))
	}
}

// TestReplayContentDetectsALateWindow is the _30 symptom expressed in subs: the
// window starts on the match's second sub, so the observer has no start-of-match
// data. The first sub differs, so the rule must fire on the starting point and not
// wait for a later position to disagree.
func TestReplayContentDetectsALateWindow(t *testing.T) {
	match := [][]byte{ack([]byte{0xAA}), ack([]byte{0xBB}), ack([]byte{0xCC})}
	theirs := observerConn(match[0], match[1], match[2])
	ours := observerConn(match[1], match[2]) // put at frame 1 of the match

	rep := &Report{}
	replayContent(rep, theirs, ours, contentOf(t, match...))

	f := findingFor(rep, RuleReplayContent)
	if f == nil {
		t.Fatalf("a window that starts past the first sub must be a must-match divergence, got %v", mustMatchNames(rep))
	}
	if !strings.Contains(f.Detail, "different sub-message of the match") {
		t.Errorf("the finding must be about the starting point, got %q", f.Detail)
	}
	if !strings.Contains(strings.Join(f.Lines, "\n"), "window starts at: ours sub 1 of 3 | original sub 0 of 3") {
		t.Errorf("the report must locate both windows in the match, got %v", f.Lines)
	}
	t.Log("\n" + f.Detail)
}

// TestReplayContentDetectsAChecksumEntryDifference: the entries are the sync-state
// page hashes the observer reconciles against, so one that is not the original's
// is a different game state, and nothing about the framing says so.
func TestReplayContentDetectsAChecksumEntryDifference(t *testing.T) {
	first := ackWithEntry([]byte{0xAA}, 0x11111111, 7)
	second := ackWithEntry([]byte{0xBB}, 0x22222222, 8)
	match := [][]byte{first, second}
	theirs := observerConn(first, second)
	ours := observerConn(first, ackWithEntry([]byte{0xBB}, 0x22222222, 9))

	rep := &Report{}
	replayContent(rep, theirs, ours, contentOf(t, match...))

	f := findingFor(rep, RuleReplayContent)
	if f == nil {
		t.Fatalf("a differing checksum entry must be a must-match divergence, got %v", mustMatchNames(rep))
	}
	if f.Index != 1 {
		t.Errorf("the finding must name the entry, got #%d", f.Index)
	}
	if !strings.Contains(f.Ours, "09 00 00 00") || !strings.Contains(f.Theirs, "08 00 00 00") {
		t.Errorf("the finding must show both entries, got ours=%q theirs=%q", f.Ours, f.Theirs)
	}
	t.Log("\n" + f.Detail)
}

// TestReplayContentSaysWhatItCouldNotRead: a body that is not an ack this decoder
// understands must be counted and named, never read as far as it parses. A
// sub-message sequence that looks real and is not would be worse than none.
func TestReplayContentSaysWhatItCouldNotRead(t *testing.T) {
	readable := ack([]byte{0xAA})
	match := [][]byte{readable, ack([]byte{0xBB})}
	// The sub length runs past the end of the body.
	truncated := []byte{ackMarker, 0x01, 0x01, 0x40, 0x00, 0xAA}
	theirs := observerConn(readable, match[1])
	ours := observerConn(readable, truncated)

	rep := &Report{}
	replayContent(rep, theirs, ours, contentOf(t, match...))

	f := informativeFor(rep, RuleReplayContent)
	if f == nil {
		t.Fatal("the rule must report the acks it could not read")
	}
	lines := strings.Join(f.Lines, "\n")
	if !strings.Contains(lines, "ours 1 | original 0") || !strings.Contains(lines, "first one, ours: 0x0A") {
		t.Errorf("the report must count and name the unreadable acks, got %v", f.Lines)
	}
	// The one sub both sides carried is still compared, and it matched.
	if !rep.OK() {
		t.Errorf("an unreadable ack is a fact about the capture, not a divergence: %v", mustMatchNames(rep))
	}
}

// TestDecodeAckRefusesWhatItCannotRead covers the decoder itself, so a shape that
// slips through the rule cannot slip through unnoticed.
func TestDecodeAckRefusesWhatItCannotRead(t *testing.T) {
	good := ack([]byte{0xAA, 0xBB})
	if _, ok := decodeAck(good); !ok {
		t.Fatalf("decodeAck must read %x", good)
	}
	for name, body := range map[string][]byte{
		"empty":              {},
		"no marker":          append([]byte{0x00, 0x00, 0x00, 0x00, 0x00}, 0x00, 0x00, 0x00),
		"sub count too high": {ackMarker, 0x02, 0x01, 0x01, 0x00, 0xAA, 0x00, 0x00, 0x00},
		"sub past the end":   {ackMarker, 0x01, 0x01, 0x40, 0x00, 0xAA, 0x00, 0x00, 0x00},
		"no pending count":   {ackMarker, 0x01, 0x01, 0x01, 0x00, 0xAA, 0x00, 0x00},
		"block past the end": {ackMarker, 0x00, 0x00, 0x00, 0x40, 0x00},
		"block marker wrong": {ackMarker, 0x00, 0x00, 0x00, 0x03, 0x00, 0x09, 0x00, 0x00},
		"trailing bytes":     append(append([]byte{}, good...), 0xFF),
	} {
		if _, ok := decodeAck(body); ok {
			t.Errorf("decodeAck must refuse a body with %s: %x", name, body)
		}
	}
}

func mustMatchNames(rep *Report) []string {
	out := make([]string, 0, len(rep.MustMatch))
	for _, f := range rep.MustMatch {
		out = append(out, f.Rule)
	}
	return out
}
