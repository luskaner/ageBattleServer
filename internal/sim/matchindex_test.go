package sim

import (
	"testing"
)

// TestIndexOfDoesNotWalkTheCursorBackwards pins the fallback in matchIndex.indexOf.
//
// The scan resolves each observer frame against the player's own 0x0A stream by
// taking the first candidate strictly after the previous index. A body repeats (the
// six-byte heartbeat is five hundred of them in a row), so a frame can run out of
// candidates after the cursor: nothing in the match's stream is that body at a
// later position than where the window already is.
//
// Two things used to happen there and both were wrong. The frame was reported at
// cands[0], the EARLIEST occurrence of that body in the match, which for a
// repeating body is the far end of the list, so a window that legitimately starts
// late in the match was reported as falling off the back of it; and that guess
// became the cursor, so prev moved BACKWARDS and every frame after it was aligned
// against a position no frame of the window had reached. One unplaceable frame
// then misaligned the rest of the window, which is what produced the spurious
// "reversed pair" of the open flake (the window starting on match frame 503
// instead of 0 after the store block recycled before the observer's first visit).
//
// The fix reports the NEAREST candidate at or below the cursor and leaves the
// cursor where it was. Falsifiable both ways: with the defect reintroduced, the two
// positions asserted below come out as 0 and 1 where they require 3 and 7.
func TestIndexOfDoesNotWalkTheCursorBackwards(t *testing.T) {
	// The match's own ack stream, as (index, body). A repeats at 0 and 3 and G at
	// 1 and 7, so both have candidates on both sides of the window's path.
	match := []string{"A", "G", "B", "A", "D", "E", "F", "G"}
	index := map[string][]int{}
	for i, body := range match {
		index[body] = append(index[body], i)
	}
	m := &matchIndex{ours: index, oursN: len(match)}

	// The window the observer was served: the first seven frames are the match in
	// order, and then one more A. The match holds A at 0 and 3 and the cursor is
	// at 6, so this frame cannot advance. The ninth frame is another G, which the
	// match holds at 1 and 7.
	window := []string{"A", "G", "B", "A", "D", "E", "F", "A", "G"}
	stream := make([]*Msg, 0, len(window))
	for _, body := range window {
		stream = append(stream, &Msg{Type: 0x0A, Body: []byte(body)})
	}

	got, ok := m.indexOf(stream)
	if !ok {
		t.Fatal("indexOf reported the window as unidentifiable, but every body it carries is in the match's stream")
	}
	if len(got) != len(window) {
		t.Fatalf("indexOf returned %d positions for %d frames", len(got), len(window))
	}
	// The first seven frames are an increasing embedding of the match and must be
	// placed where they belong.
	for i, w := range []int{0, 1, 2, 3, 4, 5, 6} {
		if got[i] != w {
			t.Fatalf("frame #%d (%s) placed at match frame %d, want %d (full: %v)", i, window[i], got[i], w, got)
		}
	}
	// The eighth frame has no candidate after 6. It is the A the match last held at
	// 3, and cands[0] (0) would claim the observer fell back to the very start of
	// the match, seven frames before where it had reached.
	if got[7] != 3 {
		t.Errorf("unplaceable frame #7 (%s) reported at match frame %d, want the nearest candidate 3 "+
			"(cands[0] is the earliest occurrence in the match and invents a reversed pair); full: %v",
			window[7], got[7], got)
	}
	// And because the cursor did not move back to 0, the ninth frame is placed at
	// its own forward occurrence, 7, and not dragged back to 1 with it.
	if got[8] != 7 {
		t.Errorf("frame #8 (%s) placed at match frame %d, want 7: one unplaceable frame must not move the "+
			"cursor backwards and misalign the rest of the window; full: %v", window[8], got[8], got)
	}
}

// TestIndexOfStillReportsAGenuineReversal is the guard against the fix above being
// a way of making the must-match stop firing. A window that really does go
// backwards has to come out as a frame that does not rise, at the index the bytes
// support, which is what the rule turns into a reversed pair.
func TestIndexOfStillReportsAGenuineReversal(t *testing.T) {
	match := []string{"A", "B", "C", "D", "E"}
	index := map[string][]int{}
	for i, body := range match {
		index[body] = append(index[body], i)
	}
	m := &matchIndex{ours: index, oursN: len(match)}
	// D (match frame 3) followed by B (match frame 1): the stream does not rise.
	stream := []*Msg{
		{Type: 0x0A, Body: []byte("A")},
		{Type: 0x0A, Body: []byte("D")},
		{Type: 0x0A, Body: []byte("B")},
	}
	got, ok := m.indexOf(stream)
	if !ok {
		t.Fatal("indexOf reported the window as unidentifiable")
	}
	if len(got) != 3 || got[0] != 0 || got[1] != 3 || got[2] != 1 {
		t.Fatalf("positions %v, want [0 3 1]", got)
	}
	if got[2] > got[1] {
		t.Fatalf("frame #2 is placed after frame #1 (%v): a reversed pair would no longer be visible", got)
	}
}