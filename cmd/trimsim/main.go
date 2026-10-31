// Command trimsim inspects a capture, crops its wall clock, and proves the crop
// changed nothing the verdict depends on.
//
// The harness replays every capture in real time, because the server emits 0x0A per
// simulation tick and not per input, so a faster replay measures the harness and not
// the server (internal/sim.Options.ReplaySpeed). The consequence is that a test
// costs what the recorded run cost: the 24 minute capture is a 25 minute test. This
// tool is the way out that keeps the comparison honest, because it shortens the DATA
// instead of the clock.
//
//	trimsim -inspect captures/2026-09-29_101831_xxx     what is in the capture, by second
//	trimsim -trim <src> -out <dst> -keep 4m -why "..."  crop it
//	trimsim -verify <src> <dst>                        replay both and compare
//	trimsim -verify -keep 3m -out <dst> <src>           all three, in one go
//
// -verify is the part that matters. It replays the full capture and the cropped one
// through exactly the rules every capture goes through and then checks that the
// verdict, the recorded answers and the observer's replay content are the same
// thing with a suffix missing (internal/sim.CheckTrim). A crop that changes any of
// them is not written, or is deleted: the default here is to leave the suite with
// the long capture rather than with a short one that quietly stopped testing
// something.
//
// Nothing here speeds a replay up, ever, and -verify is refused rather than made
// approximate when the two replays would not fit a wall-clock ceiling.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/sim"
)

func main() {
	var (
		inspect = flag.String("inspect", "", "print what a capture holds, bucketed by elapsed time, and exit")
		trim    = flag.String("trim", "", "crop the capture at this directory")
		out     = flag.String("out", "", "where the crop is written")
		keep    = flag.Duration("keep", 0, "how much of the capture's own timeline to keep")
		why     = flag.String("why", "", "the reason this crop point was chosen, recorded in the capture")
		verify  = flag.Bool("verify", false, "replay the full capture and the crop and compare the verdicts")
		report  = flag.Bool("report", false, "print the full report of each replay")
		bucket  = flag.Duration("bucket", 10*time.Second, "bucket size for -inspect")
	)
	flag.Parse()
	args := flag.Args()

	if err := run(*inspect, *trim, *out, *keep, *why, *verify, *report, *bucket, args); err != nil {
		fmt.Fprintln(os.Stderr, "trimsim:", err)
		os.Exit(1)
	}
}

func run(inspect, trim, out string, keep time.Duration, why string, verify, report bool, bucket time.Duration, args []string) error {
	switch {
	case inspect != "":
		return inspectCapture(inspect, bucket)
	case trim != "":
		if out == "" {
			return fmt.Errorf("-trim needs -out: the crop is a new capture and never overwrites one")
		}
		if err := crop(trim, out, keep, why); err != nil {
			return err
		}
		if !verify {
			return nil
		}
		return prove(trim, out, report)
	case verify:
		if len(args) != 2 {
			return fmt.Errorf("-verify takes the full capture and the crop: trimsim -verify <src> <dst>")
		}
		return prove(args[0], args[1], report)
	default:
		return fmt.Errorf("nothing to do: pass -inspect <dir>, -trim <dir> -out <dir> -keep <d>, or -verify <src> <dst>")
	}
}

func crop(src, dst string, keep time.Duration, why string) error {
	if keep <= 0 {
		return fmt.Errorf("-keep is the window to keep and it is %s: a crop of nothing is not a crop", keep)
	}
	if why == "" {
		// A crop whose reason is not written down is a crop nobody can argue with
		// later, and "we measured that it is enough" is the whole justification.
		return fmt.Errorf("-why is required: the crop point is a judgement and the judgement has to be readable in the capture")
	}
	rec, err := sim.Trim(src, dst, sim.TrimOptions{Keep: keep, Source: filepath.Base(src), Why: why})
	if err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", dst)
	fmt.Printf("  kept    : %s, from %s to %s\n", rec.KeptSpan, rec.KeptFrom, rec.KeptTo)
	fmt.Printf("  dropped : %s\n", rec.Dropped)
	fmt.Printf("  verified: NOT YET. The crop is not safe until -verify has replayed both.\n")
	return nil
}

// prove is the check the whole tool exists for: two replays, the same rules, and a
// comparison that fails loudly. On success the evidence is written into the crop's
// own metadata, so the claim travels with the data; on failure the crop is left on
// disk with its record saying it is unverified, and the tool exits non-zero so
// nothing gets committed believing it.
func prove(src, dst string, report bool) error {
	fmt.Printf("replaying the full capture %s (in real time; this is as long as the run was)\n", filepath.Base(src))
	full, err := sim.Check(src, sim.Options{})
	if err != nil {
		return fmt.Errorf("full capture: %w", err)
	}
	if report {
		fmt.Println(full.Render())
	}
	fmt.Printf("replaying the crop %s\n", filepath.Base(dst))
	trimmed, err := sim.Check(dst, sim.Options{})
	if err != nil {
		return fmt.Errorf("crop: %w", err)
	}
	if report {
		fmt.Println(trimmed.Render())
	}

	check := sim.CheckTrim(full, trimmed)
	fmt.Printf("kept %s, dropped %s\n", check.Kept, check.Dropped)
	for _, l := range check.Lines {
		fmt.Printf("  verified: %s\n", l)
	}
	leftover := leftoverNote(trimmed)
	if !check.OK() {
		for _, p := range check.Problems {
			fmt.Printf("  PROBLEM [%s] conn %d: %s\n", p.Rule, p.Conn, p.Detail)
		}
		fmt.Printf("REJECTED: the crop of %s is not safe, and %s has been left unverified\n", filepath.Base(src), filepath.Base(dst))
		return fmt.Errorf("the crop changes what the harness concludes; it is not a crop of the same run")
	}
	lines := append([]string{}, check.Lines...)
	lines = append(lines, fmt.Sprintf("the verdict is unchanged: %s", verdict(trimmed)))
	if err := sim.SetVerified(dst, lines, leftover); err != nil {
		return fmt.Errorf("recording the proof in the crop: %w", err)
	}
	fmt.Printf("ACCEPTED: the crop of %s preserves the verdict; the proof is in %s/metadata.json\n", filepath.Base(src), dst)
	return nil
}

func verdict(rep *sim.Report) string {
	if rep.OK() {
		return fmt.Sprintf("no must-match divergence against the original (%d informative)", len(rep.Informative))
	}
	return fmt.Sprintf("%d must-match divergence(s) against the original", len(rep.MustMatch))
}

// leftoverNote is how many bytes of an incomplete frame the truth ends with, which
// happens when the crop point falls inside a message. It is a fact about the crop,
// not a defect in the capture, and it is written down instead of left for a reader
// to infer from a "N trailing bytes" note.
func leftoverNote(rep *sim.Report) string {
	var notes []string
	for _, c := range rep.Conns {
		for _, n := range c.Notes {
			if strings.Contains(n, "trailing bytes") {
				notes = append(notes, fmt.Sprintf("conn %d: %s", c.ClientPort, n))
			}
		}
	}
	if len(notes) == 0 {
		return ""
	}
	return "the crop point fell inside a message: " + strings.Join(notes, "; ")
}

// inspectCapture prints the traffic of one capture bucketed by elapsed time, which
// is how a crop point is chosen: the answer is where the last thing the capture was
// taken to show happens, and that is only visible with the whole timeline in front
// of you.
func inspectCapture(dir string, bucket time.Duration) error {
	if bucket <= 0 {
		bucket = 10 * time.Second
	}
	cap, err := sim.Load(dir)
	if err != nil {
		return err
	}
	origin := cap.Origin()
	fmt.Printf("capture %s\n", filepath.Base(dir))
	fmt.Printf("  truth      : %s writer=%s\n", cap.Reference.File, cap.Reference.Writer)
	fmt.Printf("  other side : %s writer=%s\n", cap.Other.File, cap.Other.Writer)
	fmt.Printf("  commands   : ours   %s\n", strings.Join(cap.Args(), " "))
	fmt.Printf("               truth  %s\n", strings.Join(cap.Reference.Args, " "))
	if t := cap.Trim(); t != nil {
		fmt.Printf("  TRIMMED    : %s kept from %s (%s dropped), verified=%d\n", t.KeptSpan, t.KeptFrom, t.Dropped, len(t.Verified))
	}
	for _, n := range cap.Notes {
		fmt.Printf("  note: %s\n", n)
	}
	if origin.IsZero() {
		return fmt.Errorf("%s has no client message with a timestamp, so it has no timeline to inspect", dir)
	}
	span := lastMoment(cap.Conns).Sub(origin)
	fmt.Printf("  timeline   : %s long, from %s\n", span.Round(time.Second), origin.Format(time.RFC3339))

	// The control messages first, at their own moments: the crop point has to be
	// after the last of them, and they are the ones a crop is not allowed to lose.
	fmt.Println("  control messages (the ones no crop may lose):")
	for _, c := range cap.Conns {
		role := "player"
		if sim.IsObserver(c) {
			role = "OBSERVER"
		}
		for _, m := range c.Up {
			switch m.Type {
			case 0x15, 0x18, 0x1C, 0x1A:
				fmt.Printf("    t=%-10s conn %-6d %-8s up   0x%02X %s\n", rel(origin, m.At), c.ClientPort, role, m.Type, protocol.MessageTypeName(m.Type))
			}
		}
		for _, m := range c.Down {
			switch m.Type {
			case 0x19, 0x1E:
				fmt.Printf("    t=%-10s conn %-6d %-8s down 0x%02X %s\n", rel(origin, m.At), c.ClientPort, role, m.Type, protocol.MessageTypeName(m.Type))
			}
		}
	}

	type cell struct{ up, down map[byte]int }
	for _, c := range cap.Conns {
		role := "player"
		if sim.IsObserver(c) {
			role = "OBSERVER"
		}
		var upTotal, downTotal int
		buckets := map[int64]*cell{}
		for _, m := range c.Up {
			upTotal++
			k := int64(m.At.Sub(origin) / bucket)
			if buckets[k] == nil {
				buckets[k] = &cell{up: map[byte]int{}, down: map[byte]int{}}
			}
			buckets[k].up[m.Type]++
		}
		for _, m := range c.Down {
			downTotal++
			k := int64(m.At.Sub(origin) / bucket)
			if buckets[k] == nil {
				buckets[k] = &cell{up: map[byte]int{}, down: map[byte]int{}}
			}
			buckets[k].down[m.Type]++
		}
		fmt.Printf("  conn %d (%s, %s): %d up, %d down\n", c.ClientPort, c.Transport, role, upTotal, downTotal)
		var keys []int64
		for k := range buckets {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		for _, k := range keys {
			cu := buckets[k]
			if len(cu.up) == 0 && len(cu.down) == 0 {
				continue
			}
			fmt.Printf("    t=%-10s up %-28s down %s\n",
				(time.Duration(k) * bucket).Round(time.Second), renderCounts(cu.up), renderCounts(cu.down))
		}
		for _, n := range c.Notes {
			fmt.Printf("    note: %s\n", n)
		}
	}
	return nil
}

func renderCounts(m map[byte]int) string {
	if len(m) == 0 {
		return "-"
	}
	keys := make([]byte, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("0x%02X x%d", k, m[k]))
	}
	return strings.Join(parts, " ")
}

func rel(origin, at time.Time) string {
	if at.IsZero() {
		return "no time"
	}
	return at.Sub(origin).Round(time.Second).String()
}

func lastMoment(conns []*sim.Conn) time.Time {
	var last time.Time
	for _, c := range conns {
		for _, m := range c.Up {
			if !m.At.IsZero() && m.At.After(last) {
				last = m.At
			}
		}
	}
	return last
}
