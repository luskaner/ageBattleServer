package sim

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

// replay-content is what the observer's replay stream CARRIES, as opposed to
// which frames of the match it is built from: the ordered game-data
// sub-messages and the ordered checksum entries, read out of the acks and
// compared with no regard to how each side grouped them into frames.
//
// RuleReplayPageOrder identifies a replay frame by its whole body, against the
// player's own ack stream. That works while both sides group the acks the same
// way, and it stops working the moment they do not: indexOf gives up when a body
// is in neither index, and the rule then claims nothing at all. On
// 2026-09-30_010531_a553ef8fd63c that is exactly what happened — the report said
// "the replay stream could not be located in the match's own ack stream" for the
// original while it was serving 5363 frames, so the one question the capture was
// taken to answer (does the observer get the right game data?) was left with no
// answer at all. The grouping is not the client's business: two servers that put
// the same 12312 sub-messages in a different number of acks hand the observer the
// same game, and this rule is what says so.
//
// What it compares, and why each is must-match:
//
//   - the starting point. An observer put at frame 600 of the match is an
//     observer with no start-of-match data, which is the _30 symptom
//     (SESSION-JOURNAL.md "Correction of the 22"). Read as the match frame the
//     window's first sub-message is, which no amount of re-grouping can move.
//   - the sub-message sequence. A dropped, duplicated or reordered sub is the
//     observer reconstructing a different game.
//   - the checksum-entry sequence. Same argument: the entries are the
//     sync-state page hashes the observer reconciles against.
//
// How MANY subs or entries each side served stays informative, exactly as the
// frame count does in RuleReplayPageOrder: where the capture was cut and where
// the replay's settle window ends both move the tail.
//
// ackMarker is the first byte of every EndOfGameData body the session builds. A
// body that does not start with it is not an ack this decoder understands, and
// the rule says so rather than reading a length out of whatever follows.
const ackMarker = 0xFE

// ackSub is one player's game data inside an ack: the slot it came from and the
// bytes, which are the chunk the client uploaded in its 0x09.
type ackSub struct {
	slot    byte
	payload []byte
}

func (s ackSub) String() string {
	return fmt.Sprintf("slot %d len=%d %s", s.slot, len(s.payload), hexPreview(s.payload, 16))
}

// ackFrame is one decoded EndOfGameData body.
//
// The shape is the one internal/session/ack.go builds and documents, and it was
// read off the wire of 2026-09-30_010531_a553ef8fd63c on both sides before it
// was written here: 25,856 acks, 10,588 of them carrying subs, every one of them
// parsing to the last byte and none with a trailing byte.
//
//	flush = [FE][subCount] + subs + tail
//	sub   = [slot u8][len u16 LE][payload]
//	tail  = [pending count][pending slots] + block
//	block = [len u16 LE][02][count u16 LE][count * 8 bytes]  or  [00 00]
type ackFrame struct {
	subs    []ackSub
	pending []byte
	entries [][]byte
}

// decodeAck reads one ack body. It is strict on purpose: a body whose lengths do
// not add up, or that carries anything after the block, is reported as not
// understood instead of being read as far as it parses, because a decoder that
// stops halfway produces a sub-message sequence that looks real and is not.
func decodeAck(body []byte) (ackFrame, bool) {
	var f ackFrame
	if len(body) < 2 || body[0] != ackMarker {
		return f, false
	}
	i := 2
	for n := int(body[1]); n > 0; n-- {
		if i+3 > len(body) {
			return ackFrame{}, false
		}
		l := int(binary.LittleEndian.Uint16(body[i+1 : i+3]))
		if i+3+l > len(body) {
			return ackFrame{}, false
		}
		f.subs = append(f.subs, ackSub{slot: body[i], payload: body[i+3 : i+3+l]})
		i += 3 + l
	}
	if i >= len(body) {
		return ackFrame{}, false
	}
	k := int(body[i])
	i++
	if i+k > len(body) {
		return ackFrame{}, false
	}
	f.pending = append([]byte(nil), body[i:i+k]...)
	i += k
	if i+2 > len(body) {
		return ackFrame{}, false
	}
	blockLen := int(binary.LittleEndian.Uint16(body[i : i+2]))
	i += 2
	if blockLen == 0 {
		// No entries: the length is the whole block and nothing follows it.
		if i != len(body) {
			return ackFrame{}, false
		}
		return f, true
	}
	if i+blockLen > len(body) {
		return ackFrame{}, false
	}
	block := body[i : i+blockLen]
	if block[0] != 0x02 || len(block) < 3 {
		return ackFrame{}, false
	}
	count := int(binary.LittleEndian.Uint16(block[1:3]))
	if 3+8*count != len(block) {
		return ackFrame{}, false
	}
	for e := range count {
		f.entries = append(f.entries, block[3+8*e:3+8*(e+1)])
	}
	i += blockLen
	if i != len(body) {
		return ackFrame{}, false
	}
	return f, true
}

// ackStream is one connection's acks, taken apart.
type ackStream struct {
	frames int
	subs   []ackSub
	// entries keeps every 8-byte record in arrival order, and entryFrames how
	// many acks carried at least one, because the same entries spread over more
	// frames is a grouping difference and has to read as one.
	entries     [][]byte
	entryFrames int
	// unreadable counts the acks this decoder did not understand, and unread is
	// the first of them, so a report can name the frame instead of only counting.
	unreadable int
	unread     *Msg
	// subsPerFrame is the grouping, kept apart from the payload so a difference
	// in it is never mistaken for a difference in the data.
	subsPerFrame map[int]int
}

// readAcks takes a connection's 0x0A stream apart.
func readAcks(msgs []*Msg) ackStream {
	s := ackStream{subsPerFrame: map[int]int{}}
	for _, m := range msgs {
		s.frames++
		f, ok := decodeAck(m.Body)
		if !ok {
			if s.unreadable == 0 {
				s.unread = m
			}
			s.unreadable++
			continue
		}
		s.subsPerFrame[len(f.subs)]++
		s.subs = append(s.subs, f.subs...)
		if len(f.entries) > 0 {
			s.entryFrames++
			s.entries = append(s.entries, f.entries...)
		}
	}
	return s
}

// matchContent locates a window in the match by CONTENT: the match's own ordered
// sub-message sequence, per side, so a window's first sub can be named as a
// match sub index even when the two sides grouped the acks differently and no
// whole frame identifies.
type matchContent struct {
	ours   []ackSub
	theirs []ackSub
}

// subIndex returns where sub sits in the match's sequence, or -1.
func subIndex(seq []ackSub, sub ackSub) int {
	for i, s := range seq {
		if s.slot == sub.slot && string(s.payload) == string(sub.payload) {
			return i
		}
	}
	return -1
}

// replayContent compares the payload the observer was served, without regard to
// the framing, and says where in the match the window starts.
func replayContent(rep *Report, want, got *Conn, content *matchContent) {
	if !requestedObservation(want) {
		// A player is sent the session's acks directly. There is no observer
		// window here, and no store in the path.
		return
	}
	ourStream := readAcks(replayStream(got.Down))
	theirStream := readAcks(replayStream(want.Down))
	if ourStream.frames == 0 && theirStream.frames == 0 {
		// RuleReplayPageOrder already says this in its own words.
		return
	}

	lines := []string{
		fmt.Sprintf("ack frames: ours %d | original %d", ourStream.frames, theirStream.frames),
		fmt.Sprintf("game-data subs served: ours %d | original %d", len(ourStream.subs), len(theirStream.subs)),
		fmt.Sprintf("checksum entries served: ours %d | original %d", len(ourStream.entries), len(theirStream.entries)),
		fmt.Sprintf("acks carrying subs, by count: ours %s | original %s",
			groupHistogram(ourStream.subsPerFrame), groupHistogram(theirStream.subsPerFrame)),
	}
	if ourStream.unreadable > 0 || theirStream.unreadable > 0 {
		lines = append(lines, fmt.Sprintf(
			"NOT COMPARED for the acks this decoder does not understand: ours %d | original %d",
			ourStream.unreadable, theirStream.unreadable))
		if theirStream.unread != nil {
			lines = append(lines, "first one, the original's: "+theirStream.unread.String())
		}
		if ourStream.unread != nil {
			lines = append(lines, "first one, ours: "+ourStream.unread.String())
		}
	}
	lines = append(lines, content.line(ourStream, theirStream))

	// The starting point, named in the match's own subs so that re-grouping the
	// acks cannot move it.
	if len(ourStream.subs) > 0 && len(theirStream.subs) > 0 {
		ours, theirs := ourStream.subs[0], theirStream.subs[0]
		if ours.slot != theirs.slot || string(ours.payload) != string(theirs.payload) {
			rep.mustMatch(Finding{
				Rule:       RuleReplayContent,
				ClientPort: want.ClientPort,
				Detail: fmt.Sprintf("the window starts on a different sub-message of the match: ours is %s, the original's is %s. "+
					"An observer put past the start of the match has no start-of-match data, which is the _30 symptom",
					ours, theirs),
				Ours:      ours.String(),
				Theirs:    theirs.String(),
				AtOurs:    elapsedOf(firstSubMsg(got.Down), rep.started),
				AtTheirs:  elapsedOf(firstSubMsg(want.Down), rep.origin),
				Lines:     lines,
				Index:     0,
				MustMatch: true,
			})
			return
		}
	}

	if f, ok := firstSubDiff(ourStream.subs, theirStream.subs); ok {
		rep.mustMatch(Finding{
			Rule:       RuleReplayContent,
			ClientPort: want.ClientPort,
			Index:      f.i,
			Detail: fmt.Sprintf("replay sub-message #%d is %s and the original served %s in that position: "+
				"the observer is reconstructing a different game", f.i, f.ours, f.theirs),
			Ours:      f.ours.String(),
			Theirs:    f.theirs.String(),
			Lines:     lines,
			MustMatch: true,
		})
		return
	}

	if f, ok := firstEntryDiff(ourStream.entries, theirStream.entries); ok {
		rep.mustMatch(Finding{
			Rule:       RuleReplayContent,
			ClientPort: want.ClientPort,
			Index:      f.i,
			Detail: fmt.Sprintf("replay checksum entry #%d differs from the one the original served in that position",
				f.i),
			Ours:      hexPreview(f.ours, 16),
			Theirs:    hexPreview(f.theirs, 16),
			Lines:     lines,
			MustMatch: true,
		})
		return
	}

	rep.Informative = append(rep.Informative, Finding{
		Rule:       RuleReplayContent,
		ClientPort: want.ClientPort,
		Detail: fmt.Sprintf("the observer's replay carries the same game data on both sides: %d sub-messages and %d checksum entries, "+
			"in the same order, from the same sub-message of the match. The acks are grouped differently (%d frames against %d), "+
			"which is what RuleReplayPageOrder cannot see and the client does not read",
			len(ourStream.subs), len(ourStream.entries), ourStream.frames, theirStream.frames),
		Lines: lines,
	})
}

type subDiff struct {
	i      int
	ours   ackSub
	theirs ackSub
}

type entryDiff struct {
	i      int
	ours   []byte
	theirs []byte
}

// firstSubDiff returns the first position where the two sequences name different
// sub-messages, over the length both have. A length difference is not a
// difference here: the tail of a window is decided by where the capture was cut
// and by the settle window, and saying so is the count rule's job.
func firstSubDiff(ours, theirs []ackSub) (subDiff, bool) {
	for i := range min(len(ours), len(theirs)) {
		if ours[i].slot == theirs[i].slot && string(ours[i].payload) == string(theirs[i].payload) {
			continue
		}
		return subDiff{i: i, ours: ours[i], theirs: theirs[i]}, true
	}
	return subDiff{}, false
}

func firstEntryDiff(ours, theirs [][]byte) (entryDiff, bool) {
	for i := range min(len(ours), len(theirs)) {
		if string(ours[i]) == string(theirs[i]) {
			continue
		}
		return entryDiff{i: i, ours: ours[i], theirs: theirs[i]}, true
	}
	return entryDiff{}, false
}

// firstSubMsg is the first ack of a stream, which is where a window that starts
// on the match's first sub-message puts its first frame.
func firstSubMsg(msgs []*Msg) *Msg {
	stream := replayStream(msgs)
	if len(stream) == 0 {
		return nil
	}
	return stream[0]
}

// line names where in the match each side's window starts, in subs. It is the
// answer replay-page-order could not give on 2026-09-30_010531_a553ef8fd63c,
// where no whole frame of the original's stream was found in its own player
// stream and the rule said so instead of naming a position.
func (m *matchContent) line(ours, theirs ackStream) string {
	at := func(seq []ackSub, s ackStream) string {
		if len(s.subs) == 0 {
			return "none"
		}
		i := subIndex(seq, s.subs[0])
		if i < 0 {
			return "not a sub-message of the match"
		}
		return fmt.Sprintf("sub %d of %d", i, len(seq))
	}
	return fmt.Sprintf("window starts at: ours %s | original %s", at(m.ours, ours), at(m.theirs, theirs))
}

// groupHistogram renders "how many acks carried N subs" compactly, so a
// difference in grouping is one readable line and not a map dump.
func groupHistogram(hist map[int]int) string {
	keys := make([]int, 0, len(hist))
	for k := range hist {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%dx%d", hist[k], k))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " ")
}
