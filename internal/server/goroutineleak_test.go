package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/lxzan/gws"

	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/session"
)

// WHAT THIS FILE IS
//
// Go 1.27 added the `goroutineleak` profile (go1.27.md, "Goroutine leak
// profile"). runtime/pprof's own Profile doc lists it among the predefined
// profiles, pprof.Profiles() returns it, and net/http/pprof serves it at
// /debug/pprof/goroutineleak. A goroutine counts as leaked when it is blocked on
// a concurrency primitive that no runnable goroutine can reach, so it can never
// wake; the runtime decides that with the garbage collector, which is why a
// collection has to run before the profile says anything.
//
// This file drives the paths that OWN goroutines (both listeners, the
// per-connection read and write loops, the main loop, the reconnect swap, the
// observer list, the sync-timeout reap), tears them down, and asserts that
// nothing is left behind. It is the complement of SESSION-JOURNAL.md 34: that
// work measured the heap and the allocations per tick over two hours of match
// and found them flat, and it was right about everything it measured, but a
// goroutine per connection blocked on a channel nobody will ever signal does
// not move either number, because its stack is small and nothing per-tick grows.
//
// THREE MEASUREMENTS DECIDED THE SHAPE OF THESE ASSERTIONS, and all three are
// re-checked at runtime by TestLeakDetectorIsArmed, in a child process:
//   - The count is the "goroutineleak profile: total N" header of the debug=1
//     text, cross-checked against Profile.Count() read AFTER the write. Reading
//     Count() first compares against the previous sample: a forced runtime.GC()
//     does not refresh it (measured 0/1/2 against a header of 1/2/3 with 1/2/3
//     leaks planted), so the write has to come first.
//   - A collection has to happen after the leak appears: detection is the
//     collector's job, so the profile describes what the last collection saw.
//     leakedGoroutineReport forces leakProfileForceGC of them.
//   - A goroutine blocked on a primitive reachable from a RUNNABLE goroutine is
//     not reported, and is reported as soon as that reference is dropped. That
//     is the runtime's own documented caveat and it is why the cycles below drop
//     the server before they ask: with a *Server still in hand, none of the
//     paths in this file could fail this assertion.
//
// WHAT IT CANNOT SEE WITHOUT HELP, SAID UP FRONT
//
// The detector is reachability-based, and the runtime says plainly that it "may
// fail to identify leaks caused by blocking on concurrency primitives reachable
// through global variables or the local variables of runnable goroutines".
// Everything this server blocks on hangs off the Server object, so a goroutine
// that leaked while the test still held a *Server would NOT be reported. That
// is why every cycle below runs inside its own closure (runServerCycle): on
// return the server's object graph is unreachable, and a goroutine still parked
// on one of its channels is then a leak by the runtime's own definition. It is
// also why the primary assertion is the goroutine count with a bound and a
// retry -- that does not care about reachability -- and the profile is the
// second, GC-visible property.
//
// WHERE THE GOROUTINES ACTUALLY ARE
//
// All of them are in this package and in internal/net, one grep of `go ` over the
// non-test sources: the main loop and the munin loop (server.go:298, :301), the
// http.Server goroutine and one gws ReadLoop per WebSocket connection
// (ws_listener.go:162, :176), the TCPListener's accept loop and its context
// watcher (internal/net/tcp.go:72-73), the accept handler, and the read and
// write loop of every bsPort connection (client.go:87-88).
//
// internal/session starts NO goroutine at all -- session.go, manager.go, ack.go
// and replay.go have not a single `go` statement -- so its timeout path has no
// goroutine to leak of its own. What it does own is a clock, the sync-data
// timeout, and that one is exercised here through a real server, because that is
// where the closure of the connection happens.
//
// NOT THE SAME AS TestMainLoopAdaptiveTickerClientChurn: that one watches the
// client REGISTRY drain across WS connections, and says nothing about goroutines
// or about the runtime's view of reachability.

// ---------------------------------------------------------------------------
// THE FACILITY
// ---------------------------------------------------------------------------

// leakProfileName is the predefined profile Go 1.27 added. It is looked up
// rather than assumed: runtime/pprof.Lookup returns nil for a name the runtime
// does not have, and a nil *Profile would panic on WriteTo, which is a much
// worse failure than an explicit one.
const leakProfileName = "goroutineleak"

// leakProfileForceGC is how many forced collections run before a profile
// sample.
const leakProfileForceGC = 1

// leakedGoroutineReport forces the collection the detector needs and then
// samples the profile: the number of leaked goroutines and the profile's own
// text, so a failure can print the stacks instead of a bare number.
//
// The count is parsed out of the debug=1 text ("goroutineleak profile: total N",
// a line the profile always writes, even at zero) and Profile.Count() is read
// afterwards as a cross-check that the two agree.
//
// THE ORDER IS THE MEASURED PART. A forced runtime.GC() does not refresh
// Profile.Count() by itself: with 1, 2 and 3 leaks planted, Count() read right
// after runtime.GC() was 0, 1 and 2 while the text said 1, 2 and 3, i.e. Count()
// lagged exactly one refresh behind. Writing the profile is what refreshes it,
// so reading Count() BEFORE the write compares against the previous sample. The
// count therefore comes from the text, and Count() is read after the write and
// required to agree -- which is what turns "the profile quietly stopped counting"
// from a green test into a red one.
//
// The forced collection is not optional either: detection happens IN the
// collector, so the profile describes what the last collection saw, and a
// goroutine that parked a moment ago can be invisible because nothing collected
// after it.
func leakedGoroutineReport() (int, string, error) {
	p := pprof.Lookup(leakProfileName)
	if p == nil {
		return 0, "", fmt.Errorf("runtime/pprof has no %q profile; this file needs Go 1.27's goroutine leak profile", leakProfileName)
	}
	for range leakProfileForceGC {
		runtime.GC()
	}
	var buf bytes.Buffer
	if err := p.WriteTo(&buf, 1); err != nil {
		return 0, "", err
	}
	text := buf.String()
	viaHeader := -1
	for line := range strings.SplitSeq(text, "\n") {
		rest, ok := strings.CutPrefix(line, leakProfileName+" profile: total ")
		if !ok {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(rest))
		if err != nil {
			return 0, text, fmt.Errorf("parsing %q: %w", rest, err)
		}
		viaHeader = n
		break
	}
	if viaHeader < 0 {
		return 0, text, fmt.Errorf("no %q header line in the profile", leakProfileName)
	}
	if viaCount := p.Count(); viaCount != viaHeader {
		return 0, text, fmt.Errorf("the %q profile disagrees with itself: header says %d, "+
			"Profile.Count() says %d", leakProfileName, viaHeader, viaCount)
	}
	return viaHeader, text, nil
}

// leakedGoroutinesWithin retries until the profile reports nothing or the window
// is up, and returns the last count and report. It retries because the profile
// describes what the last collection saw and a teardown is still in flight on
// the first sample.
//
// The assertions are built on this rather than on the profile directly so that
// TestLeakDetectorIsArmed can run the SAME logic against a planted leak and
// require it to fire.
func leakedGoroutinesWithin(window time.Duration) (int, string, error) {
	var n int
	var text string
	deadline := time.Now().Add(window)
	for {
		count, report, err := leakedGoroutineReport()
		if err != nil {
			return 0, "", err
		}
		if count == 0 {
			return 0, "", nil
		}
		n, text = count, report
		if time.Now().After(deadline) {
			return n, text, nil
		}
		time.Sleep(leakSettlePoll)
	}
}

// ---------------------------------------------------------------------------
// THE BOUNDS, each named after what it bounds
// ---------------------------------------------------------------------------

// leakSettleWindow is how long an assertion waits for goroutines to wind down
// before giving up. It is a ceiling and not a cost: both loops below return as
// soon as the property holds, and in a healthy run they hold on the first poll.
// A goroutine's exit is asynchronous with respect to Stop() -- Stop() closes
// the sockets and returns, and the read and write loops unwind after it -- so
// the retry has to exist at all.
const leakSettleWindow = 2 * time.Second

// leakSettlePoll is the gap between two settle samples.
const leakSettlePoll = 10 * time.Millisecond

// leakSettleQuietSamples is how many samples in a row must show the SAME
// goroutine count before the baseline is considered settled. That is what makes
// a baseline a baseline instead of a snapshot taken between two goroutines. Five
// samples is 50 ms: long enough for the runtime's own lazy starts, short enough
// that it is not a sleep.
const leakSettleQuietSamples = 5

// leakGoroutineSlack is how many goroutines above the settled baseline one cycle
// may leave behind. It is not zero because the runtime hands out goroutines for
// things that are neither per-connection nor per-server (the resolver behind
// the first dial, gws's own), and because the count is a process-wide property
// rather than this server's. What it must not do is scale with the number of
// cycles, which is why it is a constant and not a fraction of the count.
const leakGoroutineSlack = 2

// leakCycleRepeats is how many cycles each path runs against its baseline, past
// the warm-up. Three is the smallest number in which a leak of one goroutine a
// cycle is a clear breach of leakGoroutineSlack (1+2+3 = 6 above a bound of 2)
// and the whole path still costs a fraction of a second.
const leakCycleRepeats = 3

// leakObserverCycles is how many join/leave rounds the observer path runs. It
// has its own number because the observer list is the one structure this
// repository deliberately empties where the original does not (SESSION-JOURNAL.md
// 32), so it is worth more than one round.
const leakObserverCycles = 2

// leakClientsPerCycle is how many clients one connect/disconnect cycle opens
// and drops. It is above 1 so that a per-connection goroutine which survives
// shows up as a multiple of the leak rather than as a rounding difference.
const leakClientsPerCycle = 4

// leakReconnectSwaps is how many socket swaps one cycle of the reconnect path
// performs. Each swap moves the peer's socket onto the new connection and closes
// the connection it replaced, so N swaps close N connections and the only one
// left is the last.
const leakReconnectSwaps = 4

// leakSyncTimeoutMS is the sync-timeout deadline the client announces with its
// LobbyConnect. The default is 60 s and no test may pay that;
// -allowTimeoutOverride exists precisely so the client can publish a shorter
// one. 60 ms is four of the 15 ms simulation periods below, so the reap cannot
// race the handshake that announces it.
const leakSyncTimeoutMS = 60

// leakSimulationPeriodMS is the -simulationPeriod of every server here. The
// captures record 125 ms; waiting for four ticks of a reap at that rate would
// cost half a second for the privilege of using the real number, so the tick is
// shortened instead and the assertions are written in ticks, not in seconds.
const leakSimulationPeriodMS = 15

// leakWaitDeadline bounds every wait in this file. A frame that never arrives
// fails the test instead of hanging it.
const leakWaitDeadline = 3 * time.Second

// leakChildWindow is the retry window TestLeakDetectorIsArmed's child runs with.
// The child has a planted leak that can never go away, so the retry exists only
// to exercise the same code the five lifecycle tests do; 200 ms is ample for
// that and keeps the falsification from costing the fast path two seconds.
const leakChildWindow = 200 * time.Millisecond

// leakTitleID and leakSessionID key the session the client and reconnect paths
// create. TitleID 5 is inside the signed 16-bit range the original's session
// tree comparator reads it as.
const (
	leakTitleID   = 5
	leakSessionID = 4242
)

// leakKickSessionID keys the second session, the one with no game data in it.
// The tick's observer kick is gated on the session being neither recording nor
// reconnectable, and recording is set the moment the frame store is created, so a
// session that has carried a page can
// never be kicked. Two sessions are the price of exercising that arm.
const leakKickSessionID = 4243

// leakObserverCap is -observerMax for these servers: StartObserve answers
// status 4 when it is 0.
const leakObserverCap = 8

// leakSecret is the LobbyConnect secret the reconnect path looks the peer up by.
// It has to be non-zero: a zero secret matches nothing.
const leakSecret = uint64(0x5EC2E7)

// leakWSTapFrames bounds the WebSocket tap's queue. An observer in these cycles
// receives tens of frames; an unbounded channel inside a test whose subject is
// goroutines would be absurd.
const leakWSTapFrames = 256

// ---------------------------------------------------------------------------
// THE FIXTURE
// ---------------------------------------------------------------------------

// leakServerConfig is the server every cycle below builds. Both transports are
// on, because the two own different goroutines: bsPort owns the accept loop, the
// context watcher and the per-connection read and write loops (client.go:87-88),
// and webSocketPort owns the http.Server goroutine plus one gws ReadLoop per
// connection (ws_listener.go:162 and :176).
func leakServerConfig(t testing.TB) *Config {
	t.Helper()
	silenceLogger()
	return &Config{
		EphemeralListenPorts: EphemeralBoth,
		SimulationPeriod:     leakSimulationPeriodMS,
		ObserverMax:          leakObserverCap,
		AllowTimeoutOverride: true,
		PageSize:             0x10000,
		MaxPagesPerMatch:     0,
		LogLevel:             "none",
	}
}

// runServerCycle builds, starts, drives and stops a whole server inside this
// call, and drops every reference to it before returning. That is not tidiness:
// it is what makes the goroutineleak profile say anything (see the header).
//
// The two teardowns run in the production order, Stop() first and cancel()
// after, which is what Run does when its context is done. The cancel is what
// releases the TCPListener's watchCtx (internal/net/tcp.go:80).
func runServerCycle(t *testing.T, cfg *Config, drive func(t *testing.T, srv *Server)) {
	t.Helper()
	func() {
		srv, err := New(cfg)
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err := srv.Start(ctx); err != nil {
			t.Fatalf("Start: %v", err)
		}
		defer func() { _ = srv.Stop() }()
		drive(t, srv)
	}()
}

// settleGoroutineCount returns runtime.NumGoroutine() once it has stopped
// falling, or the last reading before leakSettleWindow expires.
func settleGoroutineCount() int {
	quiet := 0
	last := runtime.NumGoroutine()
	deadline := time.Now().Add(leakSettleWindow)
	for time.Now().Before(deadline) {
		time.Sleep(leakSettlePoll)
		n := runtime.NumGoroutine()
		if n == last {
			quiet++
			if quiet >= leakSettleQuietSamples {
				return n
			}
			continue
		}
		quiet, last = 0, n
	}
	return last
}

// settleGoroutines waits for the goroutine count to come down to within slack of
// baseline and returns the reading it settled at.
func settleGoroutines(baseline, slack int) int {
	want := baseline + slack
	n := runtime.NumGoroutine()
	deadline := time.Now().Add(leakSettleWindow)
	for n > want && time.Now().Before(deadline) {
		time.Sleep(leakSettlePoll)
		n = runtime.NumGoroutine()
	}
	return n
}

// assertGoroutineBudget fails unless the count is within slack of the baseline
// after a forced collection, and prints what is still running.
func assertGoroutineBudget(t *testing.T, label string, baseline, slack int) {
	t.Helper()
	runtime.GC()
	after := settleGoroutines(baseline, slack)
	if after > baseline+slack {
		t.Errorf("%s left %d goroutine(s) above the baseline of %d (now %d, budget %d).\n%s",
			label, after-baseline, baseline, after, slack, goroutineDump())
	}
}

// assertNoLeakedGoroutines fails unless the Go 1.27 goroutineleak profile
// reports nothing.
func assertNoLeakedGoroutines(t *testing.T) {
	t.Helper()
	n, text, err := leakedGoroutinesWithin(leakSettleWindow)
	if err != nil {
		t.Fatalf("goroutineleak profile: %v", err)
	}
	if n != 0 {
		t.Errorf("the goroutineleak profile reports %d leaked goroutine(s):\n%s\nall goroutines:\n%s",
			n, text, goroutineDump())
	}
}

// assertNoGoroutineLeaks registers a cleanup that runs after the server's Stop
// cleanup (cleanups run LIFO, so it must be registered before the Stop one) and
// fails the test if the Go 1.27 goroutineleak profile reports any leaked
// goroutines: goroutines blocked on a concurrency primitive that can never
// become unblocked because the primitive is unreachable from any runnable
// goroutine. A short grace period lets worker goroutines wind down after Stop
// before the profile is sampled.
func assertNoGoroutineLeaks(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { assertNoLeakedGoroutines(t) })
}

// goroutineDump is the whole goroutine profile, printed with a failing
// assertion: "there is a leak" is not actionable and "this goroutine is still
// parked here" is.
func goroutineDump() string {
	var buf bytes.Buffer
	if err := pprof.Lookup("goroutine").WriteTo(&buf, 2); err != nil {
		return fmt.Sprintf("goroutine profile: %v", err)
	}
	return buf.String()
}

// assertCycleBudget is what every path below ends with: warm up, take the
// baseline, repeat the cycle, then assert the count and the profile.
func assertCycleBudget(t *testing.T, label string, cycles int, drive func(t *testing.T, srv *Server)) {
	t.Helper()
	// Warm-up first. The runtime starts goroutines lazily (the resolver behind
	// the first dial, gws's own), so a baseline taken before them describes a
	// colder process than the measured cycles see.
	runServerCycle(t, leakServerConfig(t), drive)
	baseline := settleGoroutineCount()
	t.Logf("settled baseline after one warm-up cycle: %d goroutines, budget %d above it", baseline, leakGoroutineSlack)

	for i := range cycles {
		runServerCycle(t, leakServerConfig(t), drive)
		t.Logf("cycle %d of %d done", i+1, cycles)
	}

	assertGoroutineBudget(t, label, baseline, leakGoroutineSlack)
	assertNoLeakedGoroutines(t)
}

// ---------------------------------------------------------------------------
// THE PATHS THAT OWN GOROUTINES
// ---------------------------------------------------------------------------

// TestServerStartStopReleasesEveryGoroutine builds, starts, drives and stops a
// whole server, three times over, and asserts nothing survives the cycle.
//
// What would make it fail: Stop() not releasing something it owns. It stops the
// bs listener, the ws listener and the http server, clears the client
// registries and clears `running` -- but it does not WAIT for the main loop, the
// accept loop, the context watcher, or any per-connection read and write loop.
// Those have to unwind on their own, and that is the property under test.
func TestServerStartStopReleasesEveryGoroutine(t *testing.T) {
	drive := func(t *testing.T, srv *Server) {
		// Both listeners really bound and a real connection really reached them,
		// so the test cannot pass on a server that never accepted anything.
		if srv.bsListener.GetPort() == 0 || srv.wsListener.GetPort() == 0 {
			t.Fatal("the listeners did not bind")
		}
		c := dialLeakTCP(t, srv.bsListener.GetPort())
		defer c.close()
		c.send(t, protocol.MsgLobbyPing, []byte{0x5A})
		if body := c.expect(t, protocol.MsgLobbyPing, leakWaitDeadline); len(body) != 1 || body[0] != 0x5A {
			t.Fatalf("ping echo body = %x, want the single byte 5A", body)
		}
	}
	assertCycleBudget(t, "a started/stopped server", leakCycleRepeats, drive)
}

// TestClientConnectDisconnectReleasesEveryGoroutine has real TCP clients
// complete the handshake, upload a replay page so the session's tick really
// runs, and go away: four at a time, three cycles deep.
//
// A client that connects and drops without joining anything exercises only the
// accept loop. Going through the handshake adds the peer, the ack tick, the
// frame store and the RemoveClient path on top, which is where a goroutine per
// connection would actually be created.
func TestClientConnectDisconnectReleasesEveryGoroutine(t *testing.T) {
	drive := func(t *testing.T, srv *Server) {
		connect := protocol.BuildLobbyConnect(leakTitleID, leakSessionID, leakSecret, 0, 0, 5, 0, 0)
		clients := make([]*leakConn, 0, leakClientsPerCycle)
		for i := range leakClientsPerCycle {
			c := dialLeakTCP(t, srv.bsListener.GetPort())
			clients = append(clients, c)
			c.send(t, protocol.MsgLobbyConnect, connect)
			c.expect(t, protocol.MsgLobbyConnect, leakWaitDeadline)
			c.send(t, protocol.MsgGameData, gameDataPage([]byte{0xAA, byte(i)}))
			// An ack only comes out of the session's per-tick flush, so
			// receiving one is what proves the tick ran.
			c.expectAck(t, leakWaitDeadline)
		}
		// All four handshakes name the same session, so there must be exactly
		// one session with four peers. Without this the handshake above proves
		// nothing and the teardown it leads to is not the one under test.
		sess := srv.SessionManager().GetSession(session.SessionKey{TitleID: leakTitleID, SessionID: leakSessionID})
		if sess == nil {
			t.Fatalf("no session after %d handshakes", leakClientsPerCycle)
		}
		if n := sess.PeerCount(); n != leakClientsPerCycle {
			t.Errorf("peers after %d handshakes = %d, want %d", leakClientsPerCycle, n, leakClientsPerCycle)
		}
		for _, c := range clients {
			c.close()
		}
	}
	assertCycleBudget(t, "a client connect/disconnect cycle", leakCycleRepeats, drive)
}

// TestReconnectSwapReleasesEveryGoroutine is the reconnect path, which the audit
// calls the most suspicious shape in the codebase and the one with the least
// coverage: the client reconnects with the secret its peer published, the server
// MOVES the peer's socket onto the new connection and CLOSES the connection it
// replaced (handlers.go). The replaced connection is
// left holding the old socket, so its read loop is blocked on a socket nobody is
// going to write to again and its write loop is parked on its own closeCh. Four
// swaps in a row are exactly the shape a never-released per-connection goroutine
// would survive.
//
// The swap is asserted from the client side, not merely measured afterwards: the
// replaced connection must see EOF, and the session must still hold exactly one
// peer. Without those the test would pass even if the reconnect branch never ran.
func TestReconnectSwapReleasesEveryGoroutine(t *testing.T) {
	drive := func(t *testing.T, srv *Server) {
		first := protocol.BuildLobbyConnect(leakTitleID, leakSessionID, leakSecret, 0, 0, 5, 0, 0)
		state1 := protocol.BuildLobbyConnect(leakTitleID, leakSessionID, leakSecret, 1, 0, 5, 0, 0)
		// State 2 is the arm that also resends stored frames, and no capture
		// contains a reconnect, so it gets its turn here too.
		state2 := protocol.BuildLobbyConnect(leakTitleID, leakSessionID, leakSecret, 2, 0, 5, 0, 0)

		current := dialLeakTCP(t, srv.bsListener.GetPort())
		current.send(t, protocol.MsgLobbyConnect, first)
		current.expect(t, protocol.MsgLobbyConnect, leakWaitDeadline)

		replaced := make([]*leakConn, 0, leakReconnectSwaps)
		for i := range leakReconnectSwaps {
			next := dialLeakTCP(t, srv.bsListener.GetPort())
			body := state1
			if i == leakReconnectSwaps-1 {
				body = state2
			}
			next.send(t, protocol.MsgLobbyConnect, body)
			next.expect(t, protocol.MsgLobbyConnect, leakWaitDeadline)
			// The connection just replaced must be closed BY THE SERVER. Nothing
			// of this test's closes it, which is what makes the EOF proof.
			current.expectEOF(t, leakWaitDeadline)
			replaced = append(replaced, current)
			current = next
		}

		sess := srv.SessionManager().GetSession(session.SessionKey{TitleID: leakTitleID, SessionID: leakSessionID})
		if sess == nil {
			t.Fatal("the reconnect branch must not have dropped the session")
		}
		if n := sess.PeerCount(); n != 1 {
			t.Errorf("peers after %d swaps = %d, want 1: a reconnecting client took a second slot",
				leakReconnectSwaps, n)
		}
		current.close()
	}
	assertCycleBudget(t, "a reconnect swap", leakCycleRepeats, drive)
}

// TestObserverJoinLeaveReleasesEveryGoroutine covers the observer list, the one
// structure this repository deliberately EMPTIES where the original does not:
// RemoveClient removes the entry on the disconnect, handleStopObserve does it
// on the 0x1B, and Tick() kicks every observer and empties the list when the
// session is neither recording nor reconnectable. The audit found the original
// leaves its entries behind; this asserts what ours does instead, and that the
// session's tick loop really served an observer out of the frame store before it
// went.
//
// All three ways out of the list are exercised: the StopObserve message, an
// abrupt disconnect, and the tick's kick.
func TestObserverJoinLeaveReleasesEveryGoroutine(t *testing.T) {
	drive := func(t *testing.T, srv *Server) {
		wsPort := srv.wsListener.GetPort()
		bsPort := srv.bsListener.GetPort()

		// A host on bsPort creates the session; one page gives it a frame store,
		// without which no observer can be served anything.
		host := dialLeakTCP(t, bsPort)
		host.send(t, protocol.MsgLobbyConnect, protocol.BuildLobbyConnect(leakTitleID, leakSessionID, 0xB0B, 0, 0, 5, 0, 0))
		host.expect(t, protocol.MsgLobbyConnect, leakWaitDeadline)
		host.send(t, protocol.MsgGameData, gameDataPage([]byte{0x01, 0x02}))
		host.expectAck(t, leakWaitDeadline)

		live := srv.SessionManager().GetSession(session.SessionKey{TitleID: leakTitleID, SessionID: leakSessionID})
		if live == nil {
			t.Fatal("no session to observe")
		}
		if live.ReplayPageCount() == 0 {
			t.Fatal("the frame store is empty, so no observer could have been served")
		}

		// 1. StopObserve, after the session's tick served it a 0x0A out of the
		// store: the observer path this file is here to watch.
		stopped, stopTap, _ := dialLeakWS(t, wsPort)
		observe(t, stopped, stopTap, leakSessionID, 1, true)
		stopped.WriteMessage(gws.OpcodeBinary, wsMessage(protocol.MsgStopObserve, nil))
		waitObserverCount(t, live, 0)
		closeLeakWS(stopped)

		// 2. Abrupt disconnect: RemoveClient is what takes the entry off.
		dropped, dropTap, _ := dialLeakWS(t, wsPort)
		observe(t, dropped, dropTap, leakSessionID, 2, true)
		closeLeakWS(dropped)
		waitObserverCount(t, live, 0)

		// 3. The tick's kick, on a session with no pages in it so it is not
		// recording (see leakKickSessionID).
		host2 := dialLeakTCP(t, bsPort)
		host2.send(t, protocol.MsgLobbyConnect, protocol.BuildLobbyConnect(leakTitleID, leakKickSessionID, 0xB0C, 0, 0, 5, 0, 0))
		host2.expect(t, protocol.MsgLobbyConnect, leakWaitDeadline)
		kick := srv.SessionManager().GetSession(session.SessionKey{TitleID: leakTitleID, SessionID: leakKickSessionID})
		if kick == nil {
			t.Fatal("no session to kick observers out of")
		}
		kicked, kickTap, kickDone := dialLeakWS(t, wsPort)
		observe(t, kicked, kickTap, leakKickSessionID, 3, false)
		kick.SetReconnectable(false)
		waitObserverCount(t, kick, 0)
		if body := kickTap.take(protocol.MsgStartObserveResp, leakWaitDeadline); len(body) != 1 || body[0] != 4 {
			t.Errorf("the kicked observer got 0x19 body %x, want the single byte 4", body)
		}
		// And the connection goes down, because the original's kick schedules its
		// close with a zero delay and logs that the observer session is
		// unavailable, which the socket loop tears down on its next pass. Nothing in this test
		// closes it, so the client's read loop ending is the server's doing and
		// nothing else.
		expectWSClosed(t, kickDone, "the kicked observer")
		closeLeakWS(kicked)

		// Nothing of ours is left holding an observer entry. The tick's kick
		// empties the session's list but not the server's routing map; the entry
		// there goes when the client disconnects, which is what the three
		// teardowns above did -- asynchronously, because the server learns of a
		// closed socket from its own read loop.
		waitObserverSessions(t, srv, 0)
		host.close()
		host2.close()
	}
	assertCycleBudget(t, "an observer join/leave cycle", leakObserverCycles, drive)
}

// TestSyncTimeoutReapReleasesEveryGoroutine covers the one clock in the session
// that is not the tick: the sync-data timeout. A peer that announces a short
// deadline with -allowTimeoutOverride, starts the match (which is what marks the
// peer as participating and sets the session's started latch) and then goes quiet
// is scheduled for closure by closeSyncTimedOutPeers and closed by
// closeExpiredPeers on a later tick, both off the tick's wall clock.
//
// It is the one path where the server itself creates the teardown rather than
// the client, so it is worth its own assertion: the reap must happen, and it
// must take the connection's goroutines with it.
func TestSyncTimeoutReapReleasesEveryGoroutine(t *testing.T) {
	drive := func(t *testing.T, srv *Server) {
		quiet := dialLeakTCP(t, srv.bsListener.GetPort())
		quiet.send(t, protocol.MsgLobbyConnect, protocol.BuildLobbyConnect(
			leakTitleID, leakSessionID, 0x71E, 0, 0, 5, leakSyncTimeoutMS, 0))
		quiet.expect(t, protocol.MsgLobbyConnect, leakWaitDeadline)
		// One page starts the match. Without it the peer is not participating and
		// the timeout arm never runs.
		quiet.send(t, protocol.MsgGameData, gameDataPage([]byte{0x01}))
		quiet.expectAck(t, leakWaitDeadline)
		// Now say nothing. The server must close this socket itself, and the
		// test never closes it.
		quiet.expectEOF(t, leakWaitDeadline)

		// The reaped peer was the session's only one, so the session goes with
		// it (RemoveClient -> HandleDisconnect -> eraseIfVacantLocked).
		waitSessionCount(t, srv, 0)
	}
	assertCycleBudget(t, "a sync-timeout reap", leakCycleRepeats, drive)
}

// ---------------------------------------------------------------------------
// THE DETECTOR HAS TO BE ARMED, IN A CHILD PROCESS
// ---------------------------------------------------------------------------

// leakChildEnv is how TestLeakDetectorIsArmed tells its own re-execution that it
// is the child. The falsification plants a goroutine that can never wake, and
// such a goroutine is reported by EVERY assertion in this package that samples
// the profile -- assertNoGoroutineLeaks hangs off startServer and
// startBlackBoxServer, so a planted leak left in this process would turn the
// rest of the package red. The child is a whole test binary, so the leak dies
// with it.
const leakChildEnv = "BATTLESERVER_LEAK_CHILD"

// TestLeakDetectorIsArmed is the falsification, in the suite and not only on the
// console: it plants the leak the other five tests are supposed to be blind to
// and requires this file's own measurement to catch it.
//
// The planted leak is a goroutine blocked on a channel that exists only as a
// local of the goroutine itself and that nobody will ever signal, which is the
// shape every one of the paths above would take if it stopped releasing its
// goroutines. A leak reachable through the caller's own frame would NOT be
// reported: the runtime documents that it "may fail to identify leaks caused by
// blocking on concurrency primitives reachable through global variables or the
// local variables of runnable goroutines", so the channel is created inside the
// leaking function and never leaves it.
//
// As a parent it re-executes this same test binary with leakChildEnv set; as the
// child it does the planting and the measuring, and its own exit status is the
// verdict.
func TestLeakDetectorIsArmed(t *testing.T) {
	if os.Getenv(leakChildEnv) == "1" {
		armLeakDetectorChild(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestLeakDetectorIsArmed$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(), leakChildEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("the child process did not confirm the detector: %v\n%s", err, out)
	}
	t.Logf("child process, planted leak, out:\n%s", out)
}

// armLeakDetectorChild is the half of TestLeakDetectorIsArmed that runs in the
// child. Everything it measures is printed, so the parent's log carries the
// evidence rather than just a pass.
//
// It answers three questions, in order: does the facility exist, does it catch a
// leak that cannot wake, and is its blindness the documented one (nothing behind
// a reference a runnable goroutine still holds).
func armLeakDetectorChild(t *testing.T) {
	t.Helper()

	p := pprof.Lookup(leakProfileName)
	if p == nil {
		t.Fatalf("runtime/pprof has no %q profile, so this whole file asserts nothing. "+
			"Toolchain is %s.", leakProfileName, runtime.Version())
	}
	found := false
	for _, prof := range pprof.Profiles() {
		if prof.Name() == leakProfileName {
			found = true
		}
	}
	if !found {
		t.Fatalf("pprof.Profiles() does not list %q", leakProfileName)
	}
	t.Logf("runtime %s lists the %q profile among %d predefined ones",
		runtime.Version(), leakProfileName, len(pprof.Profiles()))

	// A goroutine that has parked, so nothing below races its arrival.
	leakReady := make(chan struct{})
	go leakForever(leakReady)
	<-leakReady

	// A collection has to have happened after it parked: detection is the
	// collector's job, and the profile describes what the last one saw. A GC of
	// its own accord may already have run, so this is logged and not asserted;
	// what is asserted is that the count is right once one is forced.
	before := leakedCountNow(t)
	t.Logf("a goroutine is parked and no collection was forced: total %d", before)

	n, text, err := leakedGoroutinesWithin(leakChildWindow)
	if err != nil {
		t.Fatalf("goroutineleak profile: %v", err)
	}
	if n == 0 {
		t.Fatalf("a goroutine is blocked on a channel nobody will ever signal, and after "+
			"%d forced GC(s) the %q profile still reports none. The detector is not armed and "+
			"the five lifecycle tests in this file are a green over a vacuum.\n%s",
			leakProfileForceGC, leakProfileName, goroutineDump())
	}
	if !strings.Contains(text, "leakForever") {
		t.Fatalf("the profile reported %d leaked goroutine(s) but did not name leakForever's "+
			"stack, so it caught something else:\n%s", n, text)
	}
	t.Logf("after %d forced GC(s): total %d, and the stacks name leakForever:\n%s",
		leakProfileForceGC, n, text)

	// The documented blindness, measured. A goroutine parked on a channel that
	// this (runnable) frame still holds is NOT reported, and appears as soon as
	// the reference is dropped. This is the whole reason runServerCycle drops
	// the server before asking: with a *Server still in hand, none of the paths
	// in this file could fail this assertion.
	parked := make(chan struct{})
	reachable := make(chan struct{})
	go leakForeverOn(parked, reachable)
	<-parked
	withReference := leakedCountNow(t)
	runtime.KeepAlive(reachable)
	if withReference != n {
		t.Fatalf("with %d planted leaks, a goroutine parked on a channel this runnable frame "+
			"still references took the total to %d, want %d: the profile does not have the "+
			"reachability caveat the release notes describe, and this file's design rests on it",
			n, withReference, n)
	}
	t.Logf("a goroutine parked on a channel this frame references: total %d, not reported", withReference)

	reachable = nil
	runtime.KeepAlive(reachable)
	afterDrop := leakedCountNow(t)
	if afterDrop != n+1 {
		t.Fatalf("after dropping the only reference to the channel, the total is %d, want %d: "+
			"the detector did not pick the goroutine up once its primitive became unreachable",
			afterDrop, n+1)
	}
	t.Logf("after dropping that reference: total %d, the parked goroutine is now reported", afterDrop)
}

// leakedCountNow forces one collection and reads the profile once, without the
// retry. It is for measurements that do not want to wait.
func leakedCountNow(t *testing.T) int {
	t.Helper()
	n, _, err := leakedGoroutineReport()
	if err != nil {
		t.Fatalf("goroutineleak profile: %v", err)
	}
	return n
}

// leakForever blocks on a channel it owns and nobody else can reach. It is
// //go:noinline so the frame holding the channel is a real frame the runtime
// will not inline away.
func leakForever(ready chan<- struct{}) {
	ch := make(chan struct{})
	close(ready)
	<-ch
}

// leakForeverOn is leakForever with the channel handed in, which is what makes
// the primitive reachable from whoever holds it.
func leakForeverOn(ready chan<- struct{}, ch chan struct{}) {
	close(ready)
	<-ch
}

// ---------------------------------------------------------------------------
// SMALL WAITERS
// ---------------------------------------------------------------------------

// waitObserverCount polls a session's observer list until it reaches want.
func waitObserverCount(t *testing.T, sess *session.Session, want int) {
	t.Helper()
	deadline := time.Now().Add(leakWaitDeadline)
	for time.Now().Before(deadline) {
		if sess.ObserverCount() == want {
			return
		}
		time.Sleep(leakSettlePoll)
	}
	t.Fatalf("observer count = %d, want %d", sess.ObserverCount(), want)
}

// waitSessionCount polls the server's session map until it reaches want.
func waitSessionCount(t *testing.T, srv *Server, want int) {
	t.Helper()
	deadline := time.Now().Add(leakWaitDeadline)
	for time.Now().Before(deadline) {
		if n := srv.SessionManager().SessionCount(); n == want {
			return
		}
		time.Sleep(leakSettlePoll)
	}
	t.Fatalf("session count = %d, want %d", srv.SessionManager().SessionCount(), want)
}

// waitObserverSessions polls the server's observer routing map until it reaches
// want. It is a wait and not a single read because the entry goes when the
// server's own read loop notices the closed socket, which is after the test has
// closed it.
func waitObserverSessions(t *testing.T, srv *Server, want int) {
	t.Helper()
	deadline := time.Now().Add(leakWaitDeadline)
	for time.Now().Before(deadline) {
		srv.mu.Lock()
		n := len(srv.observerSessions)
		srv.mu.Unlock()
		if n == want {
			return
		}
		time.Sleep(leakSettlePoll)
	}
	srv.mu.Lock()
	n := len(srv.observerSessions)
	srv.mu.Unlock()
	t.Fatalf("observerSessions still holds %d connection(s), want %d", n, want)
}

// observe runs an observer's lifecycle up to the point where the session has
// registered and activated it, and requires both answers. peerIdentifier is what
// the observer list is keyed on, so two observers in the same cycle must not share
// it. wantServe asks for one 0x0A out of the frame store, which is the tick's
// observer update having really run; the kicked observer's session has no pages,
// so for that one there is nothing to serve.
func observe(t *testing.T, c *gws.Conn, tap *leakWSTap, sessionID int64, peerIdentifier uint64, wantServe bool) {
	t.Helper()
	req := make([]byte, 24)
	binary.LittleEndian.PutUint64(req[0:8], uint64(leakTitleID)<<48|uint64(sessionID))
	binary.LittleEndian.PutUint64(req[8:16], peerIdentifier)
	binary.LittleEndian.PutUint64(req[16:24], 0)
	if err := c.WriteMessage(gws.OpcodeBinary, wsMessage(0x1C, req)); err != nil {
		t.Fatalf("StartObserveRequest: %v", err)
	}
	if body := tap.take(protocol.MsgStartObserveResp, leakWaitDeadline); len(body) != 1 || body[0] != 0 {
		t.Fatalf("0x19 body = %x, want the single byte 0 (the arm that REGISTERS; SESSION-JOURNAL.md 22)", body)
	}
	if err := c.WriteMessage(gws.OpcodeBinary, wsMessage(protocol.MsgActivateObserver, nil)); err != nil {
		t.Fatalf("ActivateObserver: %v", err)
	}
	if body := tap.take(protocol.MsgActivateObserverResp, leakWaitDeadline); len(body) != 9 || body[0] != 0 {
		t.Fatalf("0x1E body = %x, want [status 0][eight bytes of time]", body)
	}
	if wantServe {
		if body := tap.take(protocol.MsgEndOfGameData, leakWaitDeadline); len(body) < 2 || body[0] != 0xFE {
			t.Fatalf("the observer got 0x0A body %x, want a FE ack out of the frame store", body)
		}
	}
}

// ---------------------------------------------------------------------------
// THE BLACK-BOX CLIENTS
// ---------------------------------------------------------------------------

// leakConn is a wire client for the bsPort transport that owns nothing: no
// goroutine of its own and no t.Cleanup, so every reference to it dies with the
// cycle that made it. A client whose sockets the test framework held on to until
// the end of the test would keep this file's picture blurred.
type leakConn struct {
	conn net.Conn
}

func dialLeakTCP(t *testing.T, port int) *leakConn {
	t.Helper()
	c, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("dial bsPort: %v", err)
	}
	return &leakConn{conn: c}
}

func (c *leakConn) send(t *testing.T, msgType byte, body []byte) {
	t.Helper()
	if _, err := c.conn.Write(wireFrame(msgType, body)); err != nil {
		t.Fatalf("write 0x%02x: %v", msgType, err)
	}
}

// readFrame reads the next [u16 length][type][body] frame, with a deadline.
func (c *leakConn) readFrame(t *testing.T, timeout time.Duration) (byte, []byte) {
	t.Helper()
	_ = c.conn.SetReadDeadline(time.Now().Add(timeout))
	var ln [2]byte
	if _, err := io.ReadFull(c.conn, ln[:]); err != nil {
		t.Fatalf("read frame length: %v", err)
	}
	buf := make([]byte, int(binary.BigEndian.Uint16(ln[:])))
	if _, err := io.ReadFull(c.conn, buf); err != nil {
		t.Fatalf("read frame body: %v", err)
	}
	return buf[0], buf[1:]
}

// expect waits for a frame of the wanted type, skipping the others. It has to
// skip: once game data is flowing, the session's tick interleaves heartbeat
// acks with everything else on the same socket.
func (c *leakConn) expect(t *testing.T, want byte, timeout time.Duration) []byte {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		typ, body := c.readFrame(t, time.Until(deadline))
		if typ == want {
			return body
		}
	}
	t.Fatalf("no frame of type 0x%02x within %s", want, timeout)
	return nil
}

// expectAck waits for a 0x0A that actually carries flushed sub-messages, which is
// the difference between a session whose tick ran and one whose peers merely
// connected. A heartbeat (an empty 0x0A) does not count.
func (c *leakConn) expectAck(t *testing.T, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		typ, body := c.readFrame(t, time.Until(deadline))
		if typ == protocol.MsgEndOfGameData && len(body) >= 2 && body[0] == 0xFE {
			return
		}
	}
	t.Fatal("no 0x0A carrying a flushed sub-message: the session tick never ran")
}

// expectEOF waits for the SERVER to close this connection. It is the assertion
// for the reconnect swap and for the sync-timeout reap: in both the teardown is
// the server's decision, and this test never closes those sockets.
func (c *leakConn) expectEOF(t *testing.T, timeout time.Duration) {
	t.Helper()
	_ = c.conn.SetReadDeadline(time.Now().Add(timeout))
	scratch := make([]byte, 256)
	for {
		if _, err := c.conn.Read(scratch); err != nil {
			var nerr net.Error
			if errors.As(err, &nerr) && nerr.Timeout() {
				t.Fatalf("the server did not close this connection within %s", timeout)
			}
			return
		}
		// Trailing frames before the close: keep reading.
	}
}

// close shuts the socket, which is a client-initiated disconnect: the paths that
// end this way are the ones whose RemoveClient runs from the read loop.
func (c *leakConn) close() { _ = c.conn.Close() }

// leakWSTap records the WebSocket frames a client receives, so an assertion can
// wait for one type and ignore the rest (an observer also gets the session's
// 0x0A stream).
type leakWSTap struct {
	frames chan []byte
}

func (h *leakWSTap) OnOpen(*gws.Conn)         {}
func (h *leakWSTap) OnClose(*gws.Conn, error) {}
func (h *leakWSTap) OnPing(*gws.Conn, []byte) {}
func (h *leakWSTap) OnPong(*gws.Conn, []byte) {}
func (h *leakWSTap) OnMessage(_ *gws.Conn, m *gws.Message) {
	defer func() { _ = m.Close() }()
	if len(h.frames) == cap(h.frames) {
		return // a tap that grew without bound would be the test's own leak
	}
	h.frames <- append([]byte(nil), m.Bytes()...)
}

// take waits for the next frame of the wanted type, skipping the others.
func (h *leakWSTap) take(want byte, timeout time.Duration) []byte {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case f := <-h.frames:
			if typ, body := frameType(f); typ == want {
				return body
			}
		case <-time.After(min(time.Until(deadline), leakSettlePoll)):
		}
	}
	return nil
}

func dialLeakWS(t *testing.T, port int) (*gws.Conn, *leakWSTap, <-chan struct{}) {
	t.Helper()
	tap := &leakWSTap{frames: make(chan []byte, leakWSTapFrames)}
	addr := fmt.Sprintf("ws://127.0.0.1:%d/", port)
	deadline := time.Now().Add(leakWaitDeadline)
	for {
		conn, _, err := gws.NewClient(tap, &gws.ClientOption{Addr: addr})
		if err == nil {
			// closed when the read loop ends. That is the only way to tell "the
			// server closed this connection" from "the client stopped
			// listening", and it is what the tick's kick has to be caught doing.
			// One goroutine either way: the loop is not wrapped in a second one.
			done := make(chan struct{})
			go func() {
				defer close(done)
				conn.ReadLoop()
			}()
			return conn, tap, done
		}
		if time.Now().After(deadline) {
			t.Fatalf("dial wsPort: %v", err)
		}
		time.Sleep(leakSettlePoll)
	}
}

// expectWSClosed requires the client's own read loop to have ended, which is what
// a WebSocket close frame or a socket close from the server looks like from here
// (gws echoes the close and returns, conn.go:110-127). It is the assertion for
// the paths where the SERVER ends the connection and nothing in the test does.
func expectWSClosed(t *testing.T, done <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(leakWaitDeadline):
		t.Errorf("%s: the server did not close this connection within %s", what, leakWaitDeadline)
	}
}

// closeLeakWS is a client-side teardown: the close frame first, so the server's
// gws ReadLoop returns the way it would for a real client, then the socket.
func closeLeakWS(c *gws.Conn) {
	_ = c.WriteClose(1000, nil)
	_ = c.NetConn().Close()
}
