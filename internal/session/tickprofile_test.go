package session

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/luskaner/ageBattleServer/internal/profiletest"
)

// WHAT THIS FILE IS
//
// tickcost_test.go measures HOW MUCH a tick costs (nanoseconds, allocations,
// bytes) and splits it by STAGE, by measuring the same fixture with pieces of the
// tick taken away. Both are counters and differences, so neither can name a
// FUNCTION: "one more sub a peer costs 494 ns" says what, and this file says
// where.
//
// So this is the pprof profile of the same fixture, over the same costTick, at the
// same peer counts and the same tick shapes the 38.2 decomposition used. It is a
// TEST for three reasons: it needs the fixture, which is a _test.go type of this
// package; it has to run in a process that is doing nothing else, which is what a
// test binary is; and a profile nobody regenerates is one that goes stale.
//
// It is SKIPPED unless BS_PROFILE_DIR is set. Nothing here runs in the suite, and
// that is the point: a profile is a measurement and not an assertion. What IS
// asserted, and always, is that the driver really drove the tick and really
// collected samples (profileAssertSamples), because a profile of an empty window
// is the one failure that produces a clean-looking report.
//
// ---------------------------------------------------------------------------
// WHAT IS AND IS NOT IN THESE PROFILES, because it decides what they may be used
// for
// ---------------------------------------------------------------------------
//
// IN: PopAckTick and AckState.PopTick (the pending list, the sort, the body),
// QueueGameData, QueueChecksumEntries, ReconcileChecksums, Session.Tick,
// closeExpiredPeers, closeSyncTimedOutPeers, updateObserversLocked, and the
// per-tick send closure the fixture builds.
//
// OUT: the socket. sweepConn counts and keeps nothing (peersweep_test.go), so
// every write to a peer is a counter increment. That is deliberate for the heap
// (a mock that recorded every message would dwarf the thing being measured) and
// it means these profiles say NOTHING about what a send costs. That half is
// internal/server/serverprofile_test.go's, and neither file answers the other's
// question.

// ---------------------------------------------------------------------------
// THE AXES, every one of them a named constant or a named slice
// ---------------------------------------------------------------------------

// profileCPUPeers and profileCPUModes are the points and the shapes. Each mode is
// profiled for profileCPUSeconds of WALL time rather than for a tick count, which
// is what makes the shares comparable: a flat share is a share of CPU SECONDS, so
// equal wall time is equal samples.
//
// Profiling the mode list is what turns the 38.2 table's differences into names.
// costModeNoUploads is the mode that "the whole flush" was measured by
// subtraction, so a function present in the full profile and absent from that one
// is in the flush. costModeFourUploads adds three more subs a peer, so whatever
// GROWS against the reference is the marginal cost of a sub. And the two peer
// counts are what separates a per-SESSION cost from a per-PEER one: a function
// whose share is the same at 8 and 32 peers runs once a tick whatever the peer
// count, and one whose share grows with the peers is in the per-peer walk.
var (
	profileCPUModes = []costMode{costModeFull, costModeOneChecksum, costModeNoUploads, costModeFourUploads, costModeObservers}
	profileCPUPeers = []int{8, 32}
)

// profileCPUSeconds is the wall-clock length of ONE cpu profile, and with it the
// whole CPU budget of this file: five modes at 32 peers and one at 8, so about
// seven seconds of profiling plus the fixtures.
//
// It is short on purpose, and the sample rate is raised to make it readable rather
// than to make it longer: at the default 100 Hz this window carries about 120
// samples, which is a top-5 with the top three fighting over the noise.
const profileCPUSeconds = 1200 * time.Millisecond

// profileCPUMaxTicks caps the ticks in ONE cpu window. profileCPUSeconds is the
// target and this is the ceiling: the window ends at whichever comes first, and
// the driver reports which, because a window that hit the cap is a measurement of
// the cap and not of the seconds.
//
// The number is the slowest point of 38.1 (44,947 ns a tick at 32 peers) times
// profileCPUSeconds, times four for the four-uploads mode, times two so a machine
// half as fast as the one those figures came from still finishes. It is a ceiling
// on a LOOP, which is what AGENTS.md rule 7 asks to be a named constant in the
// code; nothing in this file scales without one.
const profileCPUMaxTicks = 400_000

// profileCPUHz is the sampling rate of the tick profiles, and it is 1000 rather
// than the default 100 because the windows are seconds and not minutes: at 100 Hz
// this window carries about 120 samples, which cannot rank a top-10. The overhead
// is on the order of a percent of the work being measured.
//
// runtime.SetCPUProfileRate before pprof.StartCPUProfile is the documented way to
// set a rate, and StartCPUProfile's own internal call to set 100 is then REFUSED
// by the runtime, which prints "cannot set cpu profile rate until previous profile
// has finished" on stderr. That message is not a failure and does not stop the
// profile. What decides is the sample count, read back out of the file by
// profileAssertSamples, and the file above says the message is expected so nobody
// reads it as one.
const profileCPUHz = 1000

// profileMemTicks is the window of the memory profile, at every peer count and in
// the reference mode. It is 4096 because that is where the ranking stops moving:
// the fixed part of a tick is 4 to 6 objects (38.1), so the window is 25k to 35k
// allocation samples, two orders of magnitude more than there are distinct sites.
//
// Its byte cost is churn and not live heap: at the measured 11,016 B a tick (38.1)
// 4096 ticks are 45 MB through a store that holds 2 MiB, and the collector takes
// it back inside the window. profileMemory checks the live heap against the 37's
// own ceiling, so a window that stopped being transient FAILS instead of reporting
// itself as garbage.
const profileMemTicks = 4096

// profileMemRate is runtime.MemProfileRate for the memory window: 1, which records
// EVERY allocation, against Go's default of 512 KiB per sample.
//
// The default is a sampled profile and it is useless for this question, because the
// 38.3 pooling took away exactly the allocations that were big enough to be sampled:
// at 512 KiB the window's whole 48,75 MB ranked down to TWO sites (the ack body and
// the store's copy of it, 54,6% and 45,4%) and the 4 to 6 objects a tick that remain
// did not appear at all, because 6 KiB a tick against a 512 KiB sampling interval is
// a 1% chance. So the profile agreed with the conclusion "there is nothing left" for
// the wrong reason: it could not see the thing.
//
// profileMemTicks is what bounds the cost of recording all of them. It is also
// restored afterwards, because MemProfileRate is process-global and the rest of the
// suite must not pay for it.
const profileMemRate = 1

// profileMinCPUSamples is the fewest samples a cpu profile may carry and still be
// read. At profileCPUHz over profileCPUSeconds the expectation is about 1100 (the
// runtime coalesces some); 300 is a quarter of that and is the floor below which
// a top-N is a lottery. It is a named constant because it is the assertion that
// replaces "have a look at the profile".
const profileMinCPUSamples = 300

// profileMinTicks is the fewest ticks a window may drive for its per-tick figures
// to mean anything: one tick allocates six objects, and a profile cannot rank six
// objects into stages.
const profileMinTicks = 1024

// profileDirEnv turns this file on, and its value is the parent directory the
// artifacts go into. Each package writes into its own subdirectory of it, so the
// session artifacts and the server ones cannot be compared against each other by
// accident.
const profileDirEnv = "BS_PROFILE_DIR"

// profilePackage is this file's subdirectory of BS_PROFILE_DIR.
const profilePackage = "session"

// ---------------------------------------------------------------------------
// THE DRIVER
// ---------------------------------------------------------------------------

// TestTickProfile is the whole file: for every peer count, one cpu profile per
// mode and one memory window, all bounded by the constants above.
//
// It fails, rather than quietly writing a report, when a window did not do its
// work: too few samples in the profile, too few ticks in the window, a live heap
// over the 37's ceiling, or a goroutine that a tick left behind. Each of those
// produces a beautiful, empty, entirely fictional top-10.
func TestTickProfile(t *testing.T) {
	dir := profileDir(t)
	for _, peers := range profileCPUPeers {
		for _, m := range profileModesAt(peers) {
			name := fmt.Sprintf("cpu-%02dpeers-%s", peers, profileModeSlug(m))
			d := profileFixture(t, peers, m)
			before := runtime.NumGoroutine()
			ticks, elapsed, path := profileCPU(t, dir, name, d, m)
			after := runtime.NumGoroutine()
			t.Logf("%-28s %7d ticks in %6.0f ms (%6.0f ns/tick), goroutines %d -> %d",
				name, ticks, float64(elapsed.Microseconds())/1000,
				float64(elapsed.Nanoseconds())/float64(ticks), before, after)
			profileAssertSamples(t, path)
			if ticks < profileMinTicks {
				t.Errorf("%s: the window drove %d ticks, under the %d that make a per-tick figure mean anything, so "+
					"every share in it is a share of about %d objects", name, ticks, profileMinTicks, ticks*6)
			}
			if after != before {
				t.Errorf("%s: the window went from %d goroutines to %d. The session package starts none "+
					"(SESSION-JOURNAL.md 35) and a tick may not, so a difference here is a goroutine a tick leaked, "+
					"and the profile above is mostly measuring it", name, before, after)
			}
		}
		profileMemory(t, dir, peers, costModeFull)
	}
}

// profileModesAt is the mode list for one peer count: the 38.2 decomposition's
// whole list at the peer count it used, and only the reference at the count that
// is here to show what a share looks like when the peer count is small.
func profileModesAt(peers int) []costMode {
	if peers == profileCPUPeers[len(profileCPUPeers)-1] {
		return profileCPUModes
	}
	return []costMode{costModeFull}
}

// profileModeSlug is the mode's name as a file name: no spaces and nothing the
// shell would read. It is matched on the shape, not on the name, because the names
// are prose meant for a report.
func profileModeSlug(m costMode) string {
	switch {
	case m.observers > 0:
		return "observers"
	case m.uploads == costModeNoUploads.uploads:
		return "nouploads"
	case m.uploads == costModeFourUploads.uploads:
		return "4uploads"
	case m.checksums == costModeOneChecksum.checksums:
		return "1checksum"
	default:
		return "full"
	}
}

// profileFixture brings one fixture to the steady state its mode describes, which
// is costReady's whole job (tickcost_test.go): the store full, the observers
// seated and served, and costWarmTicks of the mode itself so the pools and the
// maps are in the state the figures describe and not the state the fixture was
// built in.
func profileFixture(t *testing.T, peers int, m costMode) *sweepFixture {
	t.Helper()
	d, ok := costReady(peers, m)
	if !ok {
		t.Fatalf("%d peers, %s: the fixture did not reach its steady state, so a profile of it is a profile of a "+
			"session that has not started and of a store with room to spare", peers, m.name)
	}
	return d
}

// profileCPU drives the mode under a CPU profile until profileCPUSeconds or
// profileCPUMaxTicks, whichever comes first, and returns the ticks it drove, how
// long that took and where the profile went.
//
// The rate is set once and never reset: Go's CPU profiler is process-global and
// there is one at a time, so this has to be the only thing in the process that
// starts one. The stderr line from the refused second call is expected and is
// named in profileCPUHz.
func profileCPU(t *testing.T, dir, name string, d *sweepFixture, m costMode) (int, time.Duration, string) {
	t.Helper()
	path := filepath.Join(dir, name+".prof")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	runtime.SetCPUProfileRate(profileCPUHz)
	if err := pprof.StartCPUProfile(f); err != nil {
		t.Fatalf("StartCPUProfile for %s: %v", name, err)
	}
	start := time.Now()
	ticks := 0
	for time.Since(start) < profileCPUSeconds && ticks < profileCPUMaxTicks {
		costTick(d, d.ticks, m.uploads, m.checksums)
		ticks++
	}
	elapsed := time.Since(start)
	pprof.StopCPUProfile()
	runtime.KeepAlive(d)
	return ticks, elapsed, path
}

// profileMemory is the allocation side of the same fixture: two heap profiles
// around a window of ticks, which is what turns the window's allocations into a
// DIFF rather than a total.
//
//	go tool pprof -base mem-before.prof mem-after.prof -alloc_space   what the window allocated
//	go tool pprof -base mem-before.prof mem-after.prof -alloc_objects how many objects
//	go tool pprof -inuse_space mem-after.prof                          what is still live
//	go tool pprof -inuse_objects mem-after.prof                        how many objects are live
//
// The BEFORE profile is taken with the collector OFF, so nothing the window
// allocated can be collected out from under the diff; the AFTER profile is taken
// after the collector is back on and after two forced collections, which is what
// puts the window's allocations into the profile AT ALL (Go flushes the bucket
// table into the readable profile only at a collection) and what makes its inuse
// figures the long-lived half of the answer instead of "whatever the last cycle
// had not reached".
//
// What the diff charges to the tick is then all four of the 38.1 counters, and
// what it does NOT charge is the collector's own work inside those two
// collections, which is stated here because it is the one thing in these numbers
// that is not the tick.
func profileMemory(t *testing.T, dir string, peers int, m costMode) {
	t.Helper()
	name := fmt.Sprintf("mem-%02dpeers-%s", peers, profileModeSlug(m))
	d := profileFixture(t, peers, m)

	runtime.GC()
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	oldRate := runtime.MemProfileRate
	runtime.MemProfileRate = profileMemRate
	profileWriteHeap(t, dir, name+"-before")

	old := debug.SetGCPercent(-1)
	for range profileMemTicks {
		costTick(d, d.ticks, m.uploads, m.checksums)
	}
	debug.SetGCPercent(old)

	var mid runtime.MemStats
	runtime.ReadMemStats(&mid)
	// The two collections have to happen BEFORE the after profile and not after it.
	// Go's memory profile is a bucket table that is flushed into the readable
	// profile AT A COLLECTION, so a window that runs with the collector off has
	// allocated 45 MB that no profile can see yet: measured, the -base diff of the
	// window came out as 0 of 0 total while MemStats said 45,000 KiB. The same two
	// collections give the long-lived half of the answer, so one call does both.
	live := liveHeapAfterGC()
	// The fixture has to be alive HERE and not merely before the window: runtime.KeepAlive
	// placed any earlier releases it to the collector before this reading, and the
	// 2 MiB store goes with it, so the "live heap" came out at 370 KiB instead of the
	// 2,4 MiB the same fixture holds (SESSION-JOURNAL.md 38.1). The long-lived half of
	// this measurement is the whole fixture, and a GC that cannot see it measures the
	// wrong program.
	runtime.KeepAlive(d)
	profileWriteHeap(t, dir, name+"-after")
	runtime.MemProfileRate = oldRate

	allocated := mid.TotalAlloc - before.TotalAlloc
	objects := mid.Mallocs - before.Mallocs
	t.Logf("%-28s %d ticks: %d objects (%.2f a tick), %d KiB allocated (%.0f B a tick); live after two collections "+
		"%d KiB; %d collections ran inside the window",
		name, profileMemTicks, objects, float64(objects)/float64(profileMemTicks), allocated>>10,
		float64(allocated)/float64(profileMemTicks), live>>10, mid.NumGC-before.NumGC)

	// The live heap is the LONG-LIVED half of the answer, and this window is meant
	// to leave nothing behind, so a live heap over the 37's ceiling means part of
	// what it allocated is long-lived and the whole "this is garbage" reading of
	// the alloc_space figures is wrong.
	if int(live) > peerSweepHeapCeiling {
		t.Errorf("%s: the live heap is %d bytes after two collections, over the %d the 37's ceiling allows. This window "+
			"is meant to be entirely transient, so over the ceiling means something it allocated is long-lived and "+
			"its alloc_space figures cannot be read as garbage", name, live, peerSweepHeapCeiling)
	}
	// And the SHORT-LIVED half has to exist at all: a window that allocated nothing
	// would rank an empty top-10 just as beautifully as a full one.
	if objects < profileMinTicks {
		t.Errorf("%s: the window allocated %d objects over %d ticks, under the %d the ranking needs, so the sites "+
			"below it are a list of code that never ran", name, objects, profileMemTicks, profileMinTicks)
	}
}

// profileWriteHeap writes a heap profile and returns its path. Go's heap profile
// carries all four sample types (alloc_space, alloc_objects, inuse_space,
// inuse_objects), so one file per boundary answers the long-lived and the
// short-lived question together.
func profileWriteHeap(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name+".prof")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()
	if err := pprof.Lookup("heap").WriteTo(f, 0); err != nil {
		t.Fatalf("write heap profile %s: %v", path, err)
	}
	return path
}

// profileAssertSamples reads the cpu profile back and counts its samples, which is
// the one check that says the file is a profile and not an empty one. It parses
// the protobuf rather than shelling out to pprof, so the check does not depend on
// pprof's own text output staying the same.
func profileAssertSamples(t *testing.T, path string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back %s: %v", path, err)
	}
	n := profiletest.SampleCount(raw)
	t.Logf("    %-28s %5d samples in %d KiB", filepath.Base(path), n, len(raw)>>10)
	if n < profileMinCPUSamples {
		t.Errorf("%s: the cpu profile carries %d samples, under the %d this file insists on. At profileCPUHz over "+
			"profileCPUSeconds the expectation is about %d, so either the window did not run or the rate did not "+
			"take, and the top functions of this file would be the top functions of nothing",
			filepath.Base(path), n, profileMinCPUSamples, profileCPUHz*int(profileCPUSeconds/time.Second))
	}
	// The assertion a sample count cannot make: costTick runs in this same process,
	// so a positive count only says the profiler was on. This says it was looking at
	// the tick, and it is what catches a window that drove the fixture's setup and
	// nothing else.
	if !profiletest.Contains(raw, "internal/session.(*AckState).PopTick") {
		t.Errorf("%s: the profile carries %d samples and not one of them names AckState.PopTick, which is the whole "+
			"subject of this file. A window that did not tick looks exactly like this", filepath.Base(path), n)
	}
}

// profileDir is this package's artifact directory, or a skip.
//
// A skip and not a failure: this file has no business being a gate on the suite,
// and a test that fails because an environment variable is not set is a test
// about the environment.
func profileDir(t *testing.T) string {
	t.Helper()
	base := os.Getenv(profileDirEnv)
	if base == "" {
		t.Skipf("%s is not set, so there is nowhere to write the profiles; set it to a directory to regenerate "+
			"them. See SESSION-JOURNAL.md 39", profileDirEnv)
	}
	dir := filepath.Join(base, profilePackage)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("%s=%s: %v", profileDirEnv, base, err)
	}
	return dir
}
