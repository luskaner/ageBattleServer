package sim

import (
	"strings"
	"testing"
)

// The shapes the rise check of replayPageOrder has to tell apart.
//
// A replay keeps being served tick acks after the player's own recorded stream
// ends: the player log stops when the capture stops, while the observer's window
// is still being filled during the settle the harness keeps reading for. A frame
// that arrives there has no later copy of its body in the player's stream, so
// indexOf cannot advance the cursor for it, and two of those used to be reported
// as a must-match reversed pair. Neither is one:
//
//   - a frame past the end of the player's own stream, whatever body it repeats.
//     This is the CI failure: "replay frame #17 is match frame 16, and the one
//     before it was 16" -- eighteen frames served against a player's stream of
//     seventeen.
//   - a frame that lands on the same match frame as the one before it inside the
//     window, which is a repeat and not a frame that went backwards.
//
// The count of frames is not must-match (AGENTS.md: the order and the starting
// point are), so both are count differences and stay informative. A frame that
// lands on a match frame STRICTLY before the one the window had reached, inside
// the part of the match the player's stream covers, is still a must-match: that is
// the inverted pair capture _5 crashed the client on.

// pageOrderReport runs replayPageOrder over one synthetic pair of observers. The
// player's own ack stream is the same on both sides, so every body resolves
// against the match the test wrote; want is the original's observer window and
// got is ours.
func pageOrderReport(t *testing.T, match, want, got [][]byte) *Report {
	t.Helper()
	truthPlayer := playerConn(match...)
	ourPlayer := playerConn(match...)
	index := buildMatchIndex([]*Conn{truthPlayer}, map[int]*Conn{ourPlayer.ClientPort: ourPlayer})

	rep := &Report{}
	replayPageOrder(rep, observerConn(want...), observerConn(got...), index)
	return rep
}

func informativePageOrder(rep *Report, contains string) bool {
	for _, f := range rep.Informative {
		if f.Rule == RuleReplayPageOrder && strings.Contains(f.Detail, contains) {
			return true
		}
	}
	return false
}

func TestReplayPageOrderAllowsARepeatedTailFrame(t *testing.T) {
	// Four match frames, five served, the last one another heartbeat: the match
	// holds no later copy of it, so it cannot advance. This is the CI shape.
	match := [][]byte{[]byte("P0"), []byte("E"), []byte("E"), []byte("E")}
	want := [][]byte{[]byte("P0"), []byte("E"), []byte("E")}
	got := [][]byte{[]byte("P0"), []byte("E"), []byte("E"), []byte("E"), []byte("E")}
	rep := pageOrderReport(t, match, want, got)

	if len(rep.MustMatch) != 0 {
		t.Fatalf("a repeated tail frame must not be a must-match, got %v", rep.MustMatch)
	}
	if !informativePageOrder(rep, "past its end") {
		t.Fatalf("the repeated tail frame must be reported as informative, got %v", rep.Informative)
	}
}

func TestReplayPageOrderAllowsATailPastTheEndOfTheMatch(t *testing.T) {
	// The match ends on a game-data ack, so the extra tick ack the observer is
	// served during the settle repeats an EARLIER heartbeat and resolves to a
	// strictly smaller index. It is still past the end of the player's stream and
	// still a count difference, not a reversed pair.
	match := [][]byte{[]byte("P0"), []byte("E"), []byte("E"), []byte("S3")}
	want := [][]byte{[]byte("P0"), []byte("E"), []byte("E")}
	got := [][]byte{[]byte("P0"), []byte("E"), []byte("E"), []byte("S3"), []byte("E")}
	rep := pageOrderReport(t, match, want, got)

	if len(rep.MustMatch) != 0 {
		t.Fatalf("a frame past the end of the player's stream must not be a must-match, got %v", rep.MustMatch)
	}
	if !informativePageOrder(rep, "past its end") {
		t.Fatalf("the frame past the end must be reported as informative, got %v", rep.Informative)
	}
}

func TestReplayPageOrderAllowsARepeatedFrameTheOriginalAlsoServed(t *testing.T) {
	// Both observers repeat match frame 2 in position 3, so the two windows are
	// the same answer: the repeat is a count difference and the alignment agrees.
	match := [][]byte{[]byte("P0"), []byte("E"), []byte("E"), []byte("S3")}
	window := [][]byte{[]byte("P0"), []byte("E"), []byte("E"), []byte("E")}
	rep := pageOrderReport(t, match, window, window)

	if len(rep.MustMatch) != 0 {
		t.Fatalf("the same repeated window on both sides must not be a must-match, got %v", rep.MustMatch)
	}
	if !informativePageOrder(rep, "the same frame as the one before it") {
		t.Fatalf("the repeat must be reported as informative, got %v", rep.Informative)
	}
}

func TestReplayPageOrderStillRejectsAGenuineReversal(t *testing.T) {
	// P0 (match frame 0) after S3 (match frame 3), inside the four frames the
	// player's own stream holds: a real inverted pair.
	match := [][]byte{[]byte("P0"), []byte("E"), []byte("E"), []byte("S3")}
	want := [][]byte{[]byte("P0"), []byte("E"), []byte("E")}
	got := [][]byte{[]byte("P0"), []byte("E"), []byte("E"), []byte("P0")}
	rep := pageOrderReport(t, match, want, got)

	if len(rep.MustMatch) != 1 || rep.MustMatch[0].Rule != RuleReplayPageOrder {
		t.Fatalf("a frame that goes strictly backwards must stay a must-match, got %v", rep.MustMatch)
	}
	if !strings.Contains(rep.MustMatch[0].Detail, "goes backwards") {
		t.Fatalf("the finding must say the stream goes backwards, got %q", rep.MustMatch[0].Detail)
	}
}
