package session

import (
	"strings"
	"testing"
)

// WHAT THIS FILE IS
//
// 39 turned the stage decomposition's tolerance from a fixed FRACTION of the
// reference into max(that fraction, the spread this run measured between windows of
// one mode), because the tick got 42% cheaper when AckState's maps became arrays
// and a fraction of the reference shrank with it while the noise it has to absorb
// did not. Measured on a machine with an unrelated game and an IDE on it: 2
// failures in 6 runs before, 0 in 8 after.
//
// A tolerance that can GROW is only safe if growing it does not make the check
// meaningless, and that cannot be shown from the code: it has to be shown with
// numbers. So this file drives the sign rule directly, with no clock anywhere in
// it, and pins the two things that matter:
//
//   - a difference comfortably larger than the tolerance STILL fails, at any noise
//     level, so a real "these two modes are not the same tick" cannot hide behind a
//     wide tolerance (TestStageSignVerdictStillBites);
//   - and the measured floor is the WIDEST spread and not the mean or the median,
//     because the stage that fails is the one that lands on the worst window
//     (TestStageNoiseFloorIsTheWidestSpread).
//
// Every test here calls costStageTolerance and costStageSignVerdict, the two functions
// the real test calls. That is not tidiness: with max() written out in the tests as
// well, putting the defect back (the proportional term alone, no measured floor) left
// this whole file green, because the tests were asserting their own arithmetic.
//
// The falsification of the whole change is the first test with the noise term set
// large: if the rule had been written to accept anything inside the noise, that row
// would come back "".

// TestStageSignVerdictStillBites is the falsification. Every "must fail" row is a
// difference several times the tolerance, including the rows where the tolerance is
// the widest this measurement can produce.
func TestStageSignVerdictStillBites(t *testing.T) {
	const ref = 20_000.0
	for _, tol := range []float64{0, 500, 4_000, 16_000} {
		// A difference of the WHOLE reference is a violation at every tolerance here,
		// because none of them reaches the reference. This is the row that cannot be
		// argued with: if the rule accepts a difference the size of the thing it is
		// measuring, no tolerance is doing any work.
		for _, c := range []struct {
			name    string
			delta   float64
			removes bool
		}{
			{"removes, dearer by the whole reference", ref, true},
			{"adds, cheaper by the whole reference", -ref, false},
		} {
			if why := costStageSignVerdict(ref, tol, c.delta, c.removes); why == "" {
				t.Errorf("tolerance %.0f ns, %s: the rule returned no complaint, so a difference of %.0f ns against a "+
					"reference of %.0f passed. A tolerance that can grow has swallowed the check",
					tol, c.name, c.delta, ref)
			}
		}

		// And a difference of three times whatever the tolerance is, which is the row
		// that matters once the tolerance is the MEASURED noise: the claim is that
		// noise moves the tolerance and not the check. It is skipped at a tolerance of
		// zero because "three times zero" is a zero difference, and a zero difference
		// is not a violation of anything.
		if tol > 0 {
			for _, c := range []struct {
				name    string
				delta   float64
				removes bool
			}{
				{"removes, dearer by 3x the tolerance", +3 * tol, true},
				{"adds, cheaper by 3x the tolerance", -3 * tol, false},
			} {
				why := costStageSignVerdict(ref, tol, c.delta, c.removes)
				if why == "" {
					t.Errorf("tolerance %.0f ns, %s: the rule returned no complaint for a difference of %.0f ns, so "+
						"moving the tolerance moved the check", tol, c.name, c.delta)
					continue
				}
				if c.removes && !strings.Contains(why, "REMOVES work") {
					t.Errorf("tolerance %.0f ns, %s: the complaint does not say the mode REMOVES work: %q", tol, c.name, why)
				}
				if !c.removes && !strings.Contains(why, "ADDS work") {
					t.Errorf("tolerance %.0f ns, %s: the complaint does not say the mode ADDS work: %q", tol, c.name, why)
				}
			}
		}

		// And the other side: a difference INSIDE the tolerance is not a complaint,
		// which is the property the tolerance exists for.
		if tol > 0 {
			for _, c := range []struct {
				name  string
				delta float64
			}{
				{"removes, dearer by half the tolerance", +tol / 2},
				{"adds, cheaper by half the tolerance", -tol / 2},
			} {
				if why := costStageSignVerdict(ref, tol, c.delta, c.delta > 0); why != "" {
					t.Errorf("tolerance %.0f ns, %s: the rule complained about a difference of %.0f ns that is inside "+
						"the tolerance, which is the one thing the tolerance is for: %q", tol, c.name, c.delta, why)
				}
			}
		}
	}
}

// TestStageNoiseFloorIsTheWidestSpread pins what the floor is, because the choice
// between widest, mean and median is the difference between covering the window a
// failing stage landed on and covering the average one.
//
// The failing row is the mean: with spreads of 0, 0 and 3.000 the mean is 1.000, so
// a stage that came out 2.000 ns the wrong way would be reported as noise when it is
// twice the average spread and a sixth of the worst one. The widest is 3.000 and that
// difference is correctly inside it.
func TestStageNoiseFloorIsTheWidestSpread(t *testing.T) {
	readings := []costReading{
		{slowestNs: 1_000, nsPerTick: 1_000},
		{slowestNs: 2_500, nsPerTick: 2_500},
		{slowestNs: 9_000, nsPerTick: 6_000},
	}
	if got, want := costStageNoiseFloor(readings), 3_000.0; got != want {
		t.Errorf("costStageNoiseFloor = %.0f ns, want %.0f: it is the widest gap between a mode's own fastest and "+
			"slowest window, not the mean (%.0f) and not the median", got, want, 1_000.0)
	}

	// And the consequence, which is the point: with that floor a difference of
	// 2.000 ns is inside the noise and 4.000 ns is outside it.
	const tol = 3_000.0
	if why := costStageSignVerdict(20_000, tol, +2_000, true); why != "" {
		t.Errorf("a difference of 2.000 ns against a measured spread of 3.000 ns was reported: %q. That is the whole "+
			"reason the floor is the widest spread", why)
	}
	if why := costStageSignVerdict(20_000, tol, +4_000, true); why == "" {
		t.Error("a difference of 4.000 ns against a measured spread of 3.000 ns was accepted, so the floor does not " +
			"stop being a floor")
	}
}

// TestStageNoiseFloorIsZeroWhenEveryWindowAgrees: a quiet machine produces windows
// that differ only by the clock's last tick, and then the tolerance must fall back
// to the fraction, or the check would be permanently toothless on an idle box.
func TestStageNoiseFloorIsZeroWhenEveryWindowAgrees(t *testing.T) {
	readings := make([]costReading, 5)
	for i := range readings {
		readings[i] = costReading{slowestNs: 10_275, nsPerTick: 10_275}
	}
	if got := costStageNoiseFloor(readings); got != 0 {
		t.Errorf("costStageNoiseFloor = %.0f ns on five modes whose windows all agreed to the nanosecond, want 0", got)
	}
	ref := 10_275.0
	if got, want := costStageTolerance(ref, readings), ref*costStageNoiseShare; got != want {
		t.Errorf("with no measured noise the tolerance is %.0f ns, want the proportional %.0f", got, want)
	}
}

// TestStageToleranceNeverShrinksBelowTheNoise is the property the change is for,
// stated as a test: whatever the reference costs, the tolerance is at least the
// noise that was measured. Without it a future edit that halves costStageNoiseShare,
// or a tick that gets cheaper again, silently re-arms the flake.
func TestStageToleranceNeverShrinksBelowTheNoise(t *testing.T) {
	for _, ref := range []float64{1_000, 10_275, 20_932, 36_070, 100_000} {
		for _, noise := range []float64{0, 7_182, 16_118, 40_000} {
			readings := []costReading{{slowestNs: noise, nsPerTick: 0}}
			tol := costStageTolerance(ref, readings)
			if tol < noise {
				t.Errorf("reference %.0f ns, noise %.0f ns: costStageTolerance returned %.0f ns, which is under the noise it has to absorb",
					ref, noise, tol)
			}
			if tol < ref*costStageNoiseShare {
				t.Errorf("reference %.0f ns, noise %.0f ns: costStageTolerance returned %.0f ns, which is under the proportional term %.0f ns",
					ref, noise, tol, ref*costStageNoiseShare)
			}
		}
	}
	// The two figures this change was made for, so the property is pinned to the
	// numbers it exists for and not only to the algebra: the noise measured on the
	// loaded machine where the flake appeared, against the reference the cheaper tick
	// produced.
	loaded := []costReading{{slowestNs: 7_182, nsPerTick: 0}}
	if tol := costStageTolerance(10_275, loaded); tol != 7_182.0 {
		t.Errorf("with the loaded machine's 7.182 ns of spread the tolerance is %.0f ns, want 7.182: the "+
			"proportional term alone was 1.541 ns and that is what let the flake through", tol)
	}
}
