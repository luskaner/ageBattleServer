package sim

import (
	"testing"
	"time"
)

// The two loops, and the one number that separates them.
//
// A capture test costs what its capture cost: the replay is in real time on
// purpose, because the server emits 0x0A per simulation tick and a faster replay
// measures the harness instead of the server (Options.ReplaySpeed). Five captures
// at 47 s, 1m29s, 1m30s, 3m20s and 4m10s of recorded clock are therefore about
// 700 s of test, which is more than the budget the project gives the suite
// (AGENTS.md, 10 minutes) and it is paid on every single change. The captures
// cannot be cropped any further: `CheckTrim` already cuts each one down to the
// handshake of its observer, and a suffix does not move the must-match facts.
//
// So the clock of the recording is the thing to split on, and the split is a
// number in this file rather than a list of capture names: a name list goes stale
// the day a capture is added, and a stale list either drops a capture from the
// fast path or leaves it in it forever.
//
// This file exists instead of a `if testing.Short()` inside Run because three
// separate things have to agree on it: Run, the banner TestMain prints, and the
// test that says the fast path is not the full suite. Three copies of a threshold
// is how a loop ends up skipping something nobody chose to skip.

// ShortModeBudget is the recorded wall clock above which a capture is a long
// capture. It is an explicit constant in the code and not a per-capture decision,
// for the same reason AGENTS.md gives for synthetic data: a budget held in the
// reader's head is not a budget.
//
// 20 s is where it belongs. The committed synthetic fixture records about one
// second, so it stays in the fast path with room to spare, and the shortest real
// capture records 47 s, so every real capture is out. The fast path is therefore
// "the capture that needs no capture plus everything that is not a capture",
// which is exactly the coverage that is affordable to run on every keystroke.
const ShortModeBudget = 20 * time.Second

// CaptureSpan is the wall clock a capture's recorded run covered: from the first
// message the client sent to the last moment its own side recorded anything, in
// the capture's own clock. That is the duration its replay will take, which is
// why it is what the short-mode decision is made on.
//
// A capture that cannot be read reports the error instead of a number: the caller
// then runs it, and a broken capture has to fail loudly rather than be skipped
// for being unmeasurable.
func CaptureSpan(dir string) (time.Duration, error) {
	cap, err := Load(dir)
	if err != nil {
		return 0, err
	}
	first, last := cap.origin(), cap.lastTruthMoment()
	if first.IsZero() || last.IsZero() || last.Before(first) {
		return 0, nil
	}
	return last.Sub(first), nil
}

// skipLongCapture is called at the top of Run. It returns true when the test was
// skipped, and false when it has to go on.
//
// The skip reason names the capture's own recorded clock and the command that
// replays it, because a skip that says only "skipped" is indistinguishable from a
// capture nobody wrote a test for.
func skipLongCapture(t testing.TB, dir string) bool {
	if !testing.Short() {
		return false
	}
	span, err := CaptureSpan(dir)
	if err != nil || span <= ShortModeBudget {
		// An unreadable capture is not a long capture. Let it run: refusing to
		// skip it is what keeps "the fast path skipped something because it could
		// not be measured" from being a quiet way to hide a broken capture.
		return false
	}
	t.Skipf("long capture: %s records %s of wall clock, over the %s the fast path budgets, "+
		"and a capture is replayed in real time so a test costs what the run cost. "+
		"Run it with: go test ./internal/sim -run TestSim_<name>",
		dir, span.Round(time.Millisecond), ShortModeBudget)
	return true
}
