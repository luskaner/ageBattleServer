package sim_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/luskaner/ageBattleServer/internal/sim"
)

// Which loop ran, and how to be sure a passing run is not the wrong one.
//
// A `go test` of a package that passes prints one line, "ok <pkg> <time>", and
// nothing the test binary wrote: that was measured on this toolchain, with a test
// that prints to stdout, logs through t.Log and writes to stderr, and `go test`
// without -v showed none of it. So a banner alone cannot carry this, and the
// guarantees have to be built out of things that do survive:
//
//   - Every capture test says what it skipped, through t.Skipf, which -v and
//     -json both show (short.go).
//   - TestGeneratedTestForEveryCapture fails if a capture has no generated test,
//     in BOTH modes. The fast path therefore cannot quietly stop covering a
//     capture that is in testdata: the capture and its test are tied together by
//     a test, not only by cmd/gensim -check.
//   - TestFastPathIsNotTheFullSuite fails when the fast path is not actually
//     shorter than the full one, which is what stops ShortModeBudget from drifting
//     until "short" means "everything" and nobody notices for a release.
//
// TestMain prints the mode and the two lists, because the run is the only place
// both are known at once.

// runCall matches the one line a generated test is built from.
var runCall = regexp.MustCompile(`sim\.Run\(t, "(testdata/[^"]+)"\)`)

// testdataCaptures lists the capture directories this package can replay, sorted.
func testdataCaptures(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join("testdata", e.Name(), "metadata.json")); err != nil {
			continue
		}
		out = append(out, "testdata/"+e.Name())
	}
	if len(out) == 0 {
		t.Fatal("testdata holds no capture at all")
	}
	return out
}

// generatedTests maps every capture directory a generated test in this package
// replays, to that test's name. It reads the package's own source files, which is
// what makes the check independent of cmd/gensim: a test that a human edited by
// hand to cover nothing is a test this map does not have an entry for.
func generatedTests(t *testing.T) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	out := map[string]string{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, "_test.go") {
			continue
		}
		body, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, m := range runCall.FindAllStringSubmatch(string(body), -1) {
			out[m[1]] = name
		}
	}
	return out
}

// captureClock is a capture and the wall clock it recorded, which is what its
// replay costs.
type captureClock struct {
	dir  string
	span time.Duration
	err  error
}

func captureClocks(t *testing.T) []captureClock {
	t.Helper()
	dirs := testdataCaptures(t)
	out := make([]captureClock, 0, len(dirs))
	for _, dir := range dirs {
		span, err := sim.CaptureSpan(dir)
		out = append(out, captureClock{dir: dir, span: span, err: err})
	}
	return out
}

// TestMain prints what this run is before it starts, so a -v or -json log carries
// the answer to "did the captures run?" at the top rather than only as six SKIP
// lines at the bottom.
func TestMain(m *testing.M) {
	// testing.Short reads a flag, and it panics rather than answering before the
	// command line has been parsed. m.Run would parse it, but the banner is
	// printed before m.Run.
	flag.Parse()
	mode := "FULL: every capture in testdata is replayed in real time"
	if testing.Short() {
		mode = fmt.Sprintf("FAST (-short): captures recording more than %s are SKIPPED. "+
			"THIS IS NOT THE FULL SUITE: run without -short to replay all of them", sim.ShortModeBudget)
	}
	fmt.Printf("internal/sim: %s\n", mode)
	code := m.Run()
	os.Exit(code)
}

// TestGeneratedTestForEveryCapture is the tie between a capture and its test, and
// it runs in both modes so the fast path cannot lose one.
//
// Without it the only thing standing between a capture directory and a green run
// is `cmd/gensim -check`, which is a command somebody has to remember to type.
func TestGeneratedTestForEveryCapture(t *testing.T) {
	generated := generatedTests(t)
	for _, dir := range testdataCaptures(t) {
		if generated[dir] == "" {
			t.Errorf("%s has no generated test that replays it: a capture cannot be added without its test. "+
				"Run: go run ./cmd/gensim -out internal/sim internal/sim/%s", dir, strings.TrimPrefix(dir, "testdata/"))
		}
	}
	for dir := range generated {
		if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err != nil {
			t.Errorf("a generated test replays %s, which has no metadata.json: it is not a capture this harness can read", dir)
		}
	}
}

// TestFastPathIsNotTheFullSuite says what the fast path left out, and fails if it
// left nothing out.
//
// The failure mode it exists for: ShortModeBudget is raised, or a capture is
// cropped until every one of them fits under it, and then -short silently becomes
// the full suite and the budget stops buying anything. Here that is a failure with
// a number in the message instead of a slow suite nobody questions.
//
// It also checks the opposite mistake, which is quieter: a classification that
// depends on the mode would be a rule choosing what to RUN rather than what to
// SKIP, and that is how a capture falls out of the fast loop without anybody
// choosing it. The split is by recorded clock and nothing else, so the same
// captures are long in both modes and every one is in exactly one list.
func TestFastPathIsNotTheFullSuite(t *testing.T) {
	clocks := captureClocks(t)
	short := testing.Short()
	var ran, skipped []string
	var ranSeconds, skippedSeconds time.Duration
	for _, c := range clocks {
		if c.err != nil {
			t.Errorf("%s cannot be read, so it would run in the fast path and fail there: %v", c.dir, c.err)
			continue
		}
		long := c.span > sim.ShortModeBudget
		if !short || !long {
			ran = append(ran, c.dir)
			ranSeconds += c.span
			continue
		}
		skipped = append(skipped, c.dir)
		skippedSeconds += c.span
	}
	if got, want := len(ran)+len(skipped), len(clocks); got != want {
		t.Errorf("%d capture(s) are in neither list out of %d: every capture is either replayed or skipped for being "+
			"long, and one that is in neither is being dropped by accident", got, want)
	}
	t.Logf("mode: %s\n  replayed (%s): %s\n  skipped as long (%s): %s",
		modeName(), ranSeconds.Round(time.Second), list(ran), skippedSeconds.Round(time.Second), list(skipped))

	if !short {
		if len(ran) != len(clocks) {
			t.Errorf("full mode replays %d of %d captures: the full suite is the one that skips nothing", len(ran), len(clocks))
		}
		return
	}
	if len(skipped) == 0 {
		t.Fatalf("the fast path is not shorter than the full suite: every capture in testdata records at most %s, "+
			"so -short and no flag cost the same %.0fs and the split has stopped buying anything. "+
			"Either the captures got cheap enough to run always, or ShortModeBudget is wrong",
			sim.ShortModeBudget, ranSeconds.Seconds())
	}
	if len(ran) == 0 {
		t.Fatalf("the fast path replays no capture at all: %s of skipped recorded clock and nothing to show for it",
			skippedSeconds.Round(time.Second))
	}
	// The capture that needs no capture on disk is the one coverage that has to
	// survive into the fast path: it is the only one that runs on a machine with
	// no captures, and it has already caught two defects that no capture replay
	// did (SESSION-JOURNAL.md, sections 23 and 25).
	found := false
	for _, dir := range ran {
		if strings.HasSuffix(dir, "/synthetic") {
			found = true
		}
	}
	if !found {
		t.Errorf("the synthetic fixture is not in the fast path (it records more than %s): "+
			"it is the only coverage that needs no capture, so it may not be the thing the budget drops. Replayed: %s",
			sim.ShortModeBudget, list(ran))
	}
	if ranSeconds > sim.ShortModeBudget {
		t.Errorf("the fast path budgets %s of recorded clock and plans %.0fs: %s",
			sim.ShortModeBudget, ranSeconds.Seconds(), list(ran))
	}
}

func modeName() string {
	if testing.Short() {
		return "FAST (-short)"
	}
	return "FULL"
}

func list(in []string) string {
	if len(in) == 0 {
		return "(none)"
	}
	return strings.Join(in, ", ")
}
