package server

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/luskaner/ageBattleServer/internal/profiletest"
	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// WHAT THIS FILE IS, AND WHY IT IS A SEPARATE FILE FROM THE TICK'S
//
// internal/session/tickprofile_test.go profiles the tick with the fixture's
// connection counting messages and keeping nothing (peersweep_test.go's sweepConn),
// so everything it measures stops at the transport: what a send costs is not in it.
// This file is the other half. It runs a REAL server on a REAL listener, with REAL
// TCP clients speaking the wire protocol, and profiles the process: the accept
// path, the read path, the per-connection goroutines, the main loop's tick, and the
// scheduler and allocator underneath all of it.
//
// It is also where the goroutine, mutex and block profiles come from, because those
// only mean anything with more than one goroutine. The session package starts none
// (SESSION-JOURNAL.md 35) and the tick fixture drove one, so a mutex profile taken
// there would read zero contention and would be measuring the absence of goroutines
// rather than the absence of contention.
//
// ---------------------------------------------------------------------------
// THE SCALE, and why it is this small
// ---------------------------------------------------------------------------
//
// The tick at the recorded simulation period is INVISIBLE to a wall-clock profile,
// and the arithmetic is the reason, not a guess. 125 ms is eight ticks a second
// (SESSION-JOURNAL.md 38.1) and a tick measured 20,932 ns at 32 peers, so a window of
// profileServerSeconds holds profileServerSeconds x 8 ticks, and at four peers that
// is:
//
//	3 s x 8 ticks x ~3.2 us = ~77 us of tick in 3 s of wall clock, which is 0,0026%
//
// At profileCPUHz that is 0,08 expected samples. So the tick does NOT appear in this
// profile and no reading of it from here is legitimate; what this file is for is the
// other 99,9974%, and for the goroutine, mutex and block profiles, which are about
// the connection path and not about the tick. The tick's own attribution is the
// session file's, and the two are reported side by side rather than merged.
//
// Four clients and three seconds is also the whole budget. This is NOT a load test:
// nothing here spawns a burner, nothing runs in a loop that matters to a person at
// the machine, and the message count is a named constant rather than a rate that
// drifts.

// profileServerClients is how many real TCP clients join. Four is a full AoE2DE game.
const profileServerClients = 4

// profileServerSeconds is the wall-clock length of the run. Three seconds is 24
// ticks at the recorded period.
const profileServerSeconds = 3 * time.Second

// profileServerUploadsPerTick is how many 0xFE game-data pages each client queues
// per client-side interval. One, so the tick's input is the shape the 37's sweep
// drives, which is what makes the tick's share computable above.
const profileServerUploadsPerTick = 1

// profileServerSimulationPeriod is the recorded -simulationPeriod. It is NOT
// shortened to make the tick visible in the profile: the point of this file is to
// measure the server as it runs, and a faster tick would be measuring a
// configuration no capture carries.
const profileServerSimulationPeriod = 125

// profileServerMaxFrames is the ceiling on frames one client's reader will drain, so
// a miscount upstream cannot turn into an unbounded loop. It is generous against the
// real count: at 8 ticks a second over profileServerSeconds each client receives
// about 24 acks plus its echoes, so this is two orders of magnitude of headroom.
const profileServerMaxFrames = 20000

// profileCPUHz is the sampling rate, as in the session profile: the window is
// seconds, and 100 Hz would carry about 300 samples. runtime.SetCPUProfileRate
// before StartCPUProfile, and the runtime prints "cannot set cpu profile rate until
// previous profile has finished" when StartCPUProfile's own call to set 100 is
// refused. That message is expected and is not a failure; what decides is the
// sample count in the file.
const profileCPUHz = 1000

// profileMinCPUSamples is the floor on samples, and it is LOW on purpose.
//
// This workload is nearly idle and the measurement says so: four clients at the
// recorded period, each uploading one page a tick, produce about 54 samples over
// profileServerSeconds, which is 54 ms of CPU in three seconds of wall clock, or
// 1,8% of one core. There is no honest way to raise that without turning this file
// into the load test that internal/server/load_test.go already is, and a load test
// is not what is being asked for.
//
// So the floor is set to catch the one failure that matters here -- a window that
// did not run at all, which reads 0 to 5 samples -- and NOT to make the ranking
// statistically comfortable. Measured on this machine over three runs: 54, 32 and
// 26 samples, so the floor is a quarter of the lowest of those and three times what
// a window that never ran reads. What the ranking is good for is naming the order of
// magnitude of the connection path against the tick, not for resolving two functions
// that are both under 2%.
const profileMinCPUSamples = 12

// profileDirEnv turns this file on; its value is the parent directory, and this
// package writes into its own subdirectory of it so the two sets of artifacts are
// never compared against each other.
const profileDirEnv = "BS_PROFILE_DIR"

// profilePackage is this file's subdirectory of BS_PROFILE_DIR.
const profilePackage = "server"

// TestServerRunningProfile runs a real server under every profile the runtime
// offers that means something with more than one goroutine, and asserts the run
// actually did the thing: the clients joined, the server ticked, and the CPU
// profile carries samples.
//
// The assertions are the reason this is a test. A server that never accepted, or a
// window that closed before the first tick, still produces a perfectly readable and
// completely fictional profile, and the failure that would be visible downstream is
// a byte comparison in internal/sim that says "the original sent something".
func TestServerRunningProfile(t *testing.T) {
	dir := profileDir(t)
	silenceLogger()

	// Both profilers are process-global and are switched back off afterwards,
	// because the rest of the suite must not run with block and mutex profiling on:
	// every blocking event would carry its own bookkeeping.
	//
	// Neither of these two returns the old value (SetBlockProfileRate takes a rate in
	// nanoseconds and returns nothing; SetMutexProfileFraction returns the old
	// fraction), so "off" is written as the zero both of them document rather than as
	// a value read back. Both are set BEFORE the run and read once at the end, which
	// is the only order in which the samples they collect belong to the window.
	runtime.SetBlockProfileRate(1)
	defer runtime.SetBlockProfileRate(0)
	runtime.SetMutexProfileFraction(1)
	defer runtime.SetMutexProfileFraction(0)

	srv := startServer(t, &Config{
		EphemeralListenPorts: EphemeralBS,
		SimulationPeriod:    profileServerSimulationPeriod,
		LogLevel:            "none",
	})
	port := srv.bsListener.GetPort()
	if port == 0 {
		t.Fatal("bsPort 0 after Start: the listener is not listening on a port")
	}

	// Join first, OUTSIDE the profile: the handshake is real work and it belongs to
	// the connection path, which is what the session file cannot see, but the
	// accept storm of N simultaneous dials is not representative of anything and
	// would put its own cost in the window.
	connect := lobbyConnectBody(1)
	// gorZero is the process before any client exists. The three counts it brackets
	// are what says the server runs a goroutine per connection and that the window
	// left none behind: measured against the count AFTER the readers were started,
	// a growing count would be indistinguishable from the readers arriving.
	gorZero := runtime.NumGoroutine()
	clients := make([]*profileClient, 0, profileServerClients)
	for range profileServerClients {
		c := newProfileClient(t, port)
		c.send(protocol.MsgLobbyConnect, connect)
		c.expect(protocol.MsgLobbyConnect, connect)
		clients = append(clients, c)
	}
	// The readers start HERE and not at dial time, because the handshake above reads
	// frames on this goroutine and a bufio.Reader has exactly one reader: two
	// goroutines sharing one of them is a data race AND it starves the other, which
	// is what the first version of this file did (the handshake took the frames the
	// reader was meant to count, and the window reported zero drained).
	for _, c := range clients {
		c.startReader()
	}
	gorWithClients := runtime.NumGoroutine()
	if n := srv.clientCount.Load(); n < profileServerClients {
		t.Fatalf("the server is holding %d clients after %d joined it. Everything below would then be a profile of a "+
			"server that turned the connections away", n, profileServerClients)
	}

	runtime.GC()
	runtime.GC()
	profileWrite(t, dir, "heap-before", "heap")
	var framesBefore int64
	for _, c := range clients {
		framesBefore += c.frames.Load()
	}
	// The readers were started above, so their goroutines are already in the count:
	// waiting for them here would make "did the window leak" and "did the reader
	// start" the same question.
	gorBefore := gorWithClients
	_ = gorBefore

	cpuPath := profileServerCPU(t, dir, clients)

	var framesAfter int64
	for _, c := range clients {
		framesAfter += c.frames.Load()
	}
	gorAfter := runtime.NumGoroutine()

	afterPath := profileWrite(t, dir, "heap-after", "heap")
	profileWrite(t, dir, "goroutine", "goroutine")
	profileWrite(t, dir, "mutex", "mutex")
	profileWrite(t, dir, "block", "block")

	t.Logf("%d clients, %d ms at -simulationPeriod %d: %d frames drained, goroutines %d with no client, %d with "+
		"them, %d after the window; %s, %s and the goroutine, mutex and block profiles written",
		len(clients), profileServerSeconds.Milliseconds(), profileServerSimulationPeriod,
		framesAfter-framesBefore, gorZero, gorBefore, gorAfter, filepath.Base(cpuPath), filepath.Base(afterPath))

	// The run has to have done the three things that make its numbers mean anything.
	if framesAfter-framesBefore < int64(profileServerClients) {
		t.Errorf("the clients drained %d frames in %d ms, under the %d that %d clients at %d ticks a second cannot "+
			"avoid. The server either never ticked or never answered, and both the cpu profile and the goroutine "+
			"profile above are then of a server that sat there",
			framesAfter-framesBefore, profileServerSeconds.Milliseconds(), profileServerClients,
			profileServerClients, 1000/profileServerSimulationPeriod)
	}
	profileAssertSamples(t, cpuPath)
	// One goroutine per connection on the server's side, plus one per reader here.
	// Measured against gorWithClients and not against the window's own start, so a
	// reader arriving late cannot be read as the server not running one.
	if grew := gorBefore - gorZero; grew < profileServerClients {
		t.Errorf("%d clients took the process from %d goroutines to %d, under the %d that %d clients at one "+
			"goroutine each need. The profile above is then not of the server this server is",
			profileServerClients, gorZero, gorBefore, profileServerClients, profileServerClients)
	}
	if gorAfter < gorBefore {
		t.Errorf("the window went from %d goroutines to %d, so it ended with fewer than it started with. A tick "+
			"that closed goroutines would show up here as a tick that closed the wrong ones", gorBefore, gorAfter)
	}
}

// profileServerCPU drives the clients under a CPU profile for profileServerSeconds
// and returns the profile's path.
//
// Each client is driven from its OWN goroutine on purpose: with one goroutine
// uploading for all of them the mutex profile would read zero contention and the
// answer would be about the shape of the test rather than about the server.
func profileServerCPU(t *testing.T, dir string, clients []*profileClient) string {
	t.Helper()
	path := filepath.Join(dir, "cpu.prof")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	runtime.SetCPUProfileRate(profileCPUHz)
	if err := pprof.StartCPUProfile(f); err != nil {
		t.Fatalf("StartCPUProfile: %v", err)
	}
	deadline := time.Now().Add(profileServerSeconds)
	interval := time.Duration(profileServerSimulationPeriod) * time.Millisecond
	var wg sync.WaitGroup
	for _, c := range clients {
		wg.Add(1)
		go func(c *profileClient) {
			defer wg.Done()
			page := gameDataPage(make([]byte, profileServerUploadBytes))
			for round := 0; time.Now().Before(deadline); round++ {
				for range profileServerUploadsPerTick {
					// The payload carries the round so the server's checksums differ
					// tick to tick, which is what a real client does and what keeps
					// the reconcile from short-circuiting on identical input.
					binary.LittleEndian.PutUint16(page[2:], uint16(round))
					if err := c.trySend(protocol.MsgGameData, page); err != nil {
						return
					}
				}
				time.Sleep(interval)
			}
		}(c)
	}
	wg.Wait()
	pprof.StopCPUProfile()
	return path
}

// profileServerUploadBytes is the payload of one 0xFE page. It is the 37's fixture
// size (peerSweepUploadBytes, 150) so the tick's input is the shape the sweep's
// figures describe and the arithmetic above holds.
const profileServerUploadBytes = 150

// profileClient is one real TCP client: the wire protocol from the outside, plus a
// reader goroutine so the server's writes are always drained.
//
// The reader exists so the numbers are of a working server: a client that stops
// reading fills the socket buffer, the server's write blocks, and the block profile
// then measures the test's own backpressure instead of the server.
type profileClient struct {
	conn   net.Conn
	rd     *bufio.Reader
	frames atomic.Int64
	once   sync.Once
}

// newProfileClient dials the listener and registers the close. It does NOT start the
// reader: the handshake reads on the test's own goroutine and a bufio.Reader has one
// reader. Call startReader once the handshake is done.
func newProfileClient(t *testing.T, port int) *profileClient {
	t.Helper()
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return &profileClient{conn: conn, rd: bufio.NewReader(conn)}
}

// startReader drains the connection until it closes or the ceiling is reached, once.
func (c *profileClient) startReader() {
	c.once.Do(func() { go c.read() })
}

// read never fails the test: a reader goroutine must not call t.Fatalf (that is a
// test-harness rule, and a failure raised here would be raised about the wrong
// goroutine on a connection the harness is closing anyway). Every outcome is
// counted, and the counts are what the driver asserts on.
func (c *profileClient) read() {
	for c.frames.Load() < profileServerMaxFrames {
		_ = c.conn.SetReadDeadline(time.Now().Add(2 * profileServerSeconds))
		var lenBuf [2]byte
		if _, err := io.ReadFull(c.rd, lenBuf[:]); err != nil {
			return
		}
		length := int(binary.BigEndian.Uint16(lenBuf[:]))
		if length < 1 {
			return
		}
		payload := make([]byte, length)
		if _, err := io.ReadFull(c.rd, payload); err != nil {
			return
		}
		c.frames.Add(1)
	}
}

func (c *profileClient) send(msgType byte, body []byte) {
	_ = c.trySend(msgType, body)
}

// trySend writes one framed message and reports the error instead of failing, so
// the per-client goroutine can keep its own deadline.
func (c *profileClient) trySend(msgType byte, body []byte) error {
	frame := wireFrame(msgType, body)
	_, err := c.conn.Write(frame)
	return err
}

// expect reads frames until one of msgType arrives, or the connection ends. It is
// only used for the handshake, which runs on the test's own goroutine.
func (c *profileClient) expect(msgType byte, body []byte) {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		typ, got, ok := c.tryReadFrame(200 * time.Millisecond)
		if !ok {
			continue
		}
		if typ == msgType && string(got) == string(body) {
			return
		}
	}
}

// tryReadFrame reads one framed message with a deadline, reporting a timeout as
// "not yet" rather than as an error.
func (c *profileClient) tryReadFrame(timeout time.Duration) (byte, []byte, bool) {
	if err := c.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return 0, nil, false
	}
	var lenBuf [2]byte
	if _, err := io.ReadFull(c.rd, lenBuf[:]); err != nil {
		return 0, nil, false
	}
	length := int(binary.BigEndian.Uint16(lenBuf[:]))
	if length < 1 {
		return 0, nil, false
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(c.rd, payload); err != nil {
		return 0, nil, false
	}
	return payload[0], payload[1:], true
}

// profileWrite writes one named runtime profile into dir.
func profileWrite(t *testing.T, dir, name, lookup string) string {
	t.Helper()
	p := pprof.Lookup(lookup)
	if p == nil {
		t.Fatalf("runtime has no %q profile on this platform", lookup)
	}
	path := filepath.Join(dir, name+".prof")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()
	// debug 0 for heap (the sampled protobuf, so pprof can diff it) and debug 1 for
	// the rest, whose protobuf form is not useful to read by hand.
	debug := 0
	if lookup != "heap" {
		debug = 1
	}
	if err := p.WriteTo(f, debug); err != nil {
		t.Fatalf("write %s profile: %v", lookup, err)
	}
	return path
}

// profileAssertSamples is the same check the session profile makes, on the same
// shared counter, because both files have to answer the same question before they
// may report anything: does the file carry samples, or is it an empty window with a
// beautiful top-10 waiting to be read out of it. See internal/profiletest for why
// counting Sample records is not the same number.
func profileAssertSamples(t *testing.T, path string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back %s: %v", path, err)
	}
	n := profiletest.SampleCount(raw)
	t.Logf("    %-14s %5d samples in %d KiB", filepath.Base(path), n, len(raw)>>10)
	if n < profileMinCPUSamples {
		t.Errorf("%s: the cpu profile carries %d samples, under the %d this file insists on. A window that did not "+
			"run reads 0 to 5, and the measured range for one that did is 26 to 54 (profileMinCPUSamples), so this "+
			"is a window that did not run and the top functions of this file would be the top functions of nothing",
			filepath.Base(path), n, profileMinCPUSamples)
	}
	// The stronger of the two: the fixtures run in this same process, so a positive
	// sample count only says the profiler was on. This says it was looking at the
	// server's own connection path and not only at the test harness that drives it.
	if !profiletest.Contains(raw, "internal/server.") {
		t.Errorf("%s: the profile carries %d samples and not one of them names any function of this package. The "+
			"clients are in this process too, so a well-formed profile of the harness alone is exactly what a window "+
			"in which the server never ran looks like, and nothing above it would say so",
			filepath.Base(path), n)
	}
}

// profileDir is this package's artifact directory, or a skip. A skip and not a
// failure: this file is not a gate on the suite.
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
