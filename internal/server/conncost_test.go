package server

import (
	"runtime"
	"testing"
	"time"

	"github.com/lxzan/gws"
)

// WHAT THIS FILE IS
//
// The third per-client cost, and the one neither of the other two files covers.
// peersweep_test.go measures what a PEER costs inside a session: the frame store
// block it adds and the upload queue it can fill. Both of those are per session.
// What a CONNECTED CLIENT costs before it is anybody's peer is a different set of
// objects: the listener's socket manager entry, the transport's read and write
// buffers, the client record in three maps (Server.clients, clientSet and
// pingEchoBudget, server.go:128-168) and the goroutines that own the connection
// (the http.Server's per-connection goroutine plus gws's ReadLoop,
// ws_listener.go:162 and :176).
//
// SESSION-JOURNAL.md 35 measured the RESIDUE of that: 1000 observer join and leave
// cycles leaving 0 KiB behind. Residue is not cost. This file measures what a
// live connection holds while it is connected, which is the number that decides
// how many clients one process can serve before the machine runs out of memory,
// and it is a DIFFERENTIAL: the same server with no connections at all is the
// control, because the server itself, its listeners and its metrics collector are
// in both readings and are not what is being measured.
//
// There is no formula from the binary here, and there is not going to be one: the
// original's per-connection cost is a socket manager, a 24-byte connection object
// and a RLink buffer pool whose sizes are not in this repository's evidence (the
// decompilation has the allocations but not the allocator). What CAN be asserted
// is the slope, in the same spirit as the peer sweep: a client has to cost about
// what a connection's buffers and its two goroutines cost, and not ten times that.
//
// Both stacks and heap are reported, because a connection is mostly stack: the
// gws ReadLoop and net/http's connection goroutine between them hold a few
// kilobytes of stack that HeapAlloc cannot see at all. A measurement that only
// looked at the heap would call a connection nearly free.

// ---------------------------------------------------------------------------
// THE CAPS, and what each one bounds
// ---------------------------------------------------------------------------

const (
	// connCostPasses is how many times each point runs, keeping the minimum. The
	// first measurement of a live heap in a process reads high whatever it
	// measures (SESSION-JOURNAL.md 34), and here there is more that can be cold
	// than there is in a session: the goroutine stacks, the http server's
	// per-connection buffers and the allocator's own page cache.
	connCostPasses = 2

	// connCostMaxHeapPerConn bounds the HEAP slope of a live connection. Measured
	// 26 KiB a connection, of which 2 KiB is the gws read and write buffers
	// (ws_listener.go:127, ReadBufferSize 1024) and 4 KiB each for net/http's
	// bufio reader and writer around the socket. 64 KiB is 2.4 times that, and
	// the thing it has to catch is a per-connection buffer that was sized for the
	// worst case instead of for a frame.
	connCostMaxHeapPerConn = 64 << 10

	// connCostMaxTotalPerConn bounds the heap PLUS the stacks, which is what the
	// process actually holds. Measured 40 KiB a connection: the heap's 26 plus
	// about 14 of stack for the two goroutines that own the socket (a 2 KiB
	// minimum stack that grows to a few more in a frame loop, and gws's ReadLoop
	// plus net/http's conn goroutine both hold one). 128 KiB is 3.2 times that.
	// It is the looser of the two on purpose: StackInuse counts whole 8 KiB spans,
	// so it moves in steps rather than in bytes and is the noisiest number here.
	connCostMaxTotalPerConn = 128 << 10

	// connCostSettleWindow is how long a point waits for the server to have
	// registered the connections it was asked to open. Without it the heap would
	// be read while some of them are still handshaking, which reads LOW and would
	// quietly understate the cost. This is a ceiling on a wait that normally ends
	// in a millisecond, not a sleep.
	connCostSettleWindow = 3 * time.Second

	// connCostSettlePoll is the poll interval inside that window.
	connCostSettlePoll = 5 * time.Millisecond
)

// connCostCounts is the sweep: no connections at all (the control), then a few,
// then a lot. The count has to be small enough that the whole file is affordable
// on every change and large enough that the per-connection figure is above the
// noise of a live server's own heap: measured, the heap moves about a kilobyte
// between two identical readings, and a connection costs tens of kilobytes. It is
// a var because a composite literal is not a constant expression.
var connCostCounts = [...]int{0, 8, 32}

// connCostReading is one point of the sweep, taken with the connections open.
type connCostReading struct {
	conns      int
	heap       uint64
	stacks     uint64
	goroutines int
}

// connCostSnapshot is the live heap and the live stacks with the connections
// still open. Two forced collections, for the reason peersweep_test.go gives:
// sync.Pool drops its victim cache on the second one.
func connCostSnapshot() (heap, stacks uint64) {
	runtime.GC()
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.HeapAlloc, ms.StackInuse
}

// connCostPoint opens conns WebSocket connections against a running server and
// reports what the process holds. Everything it measures is read with the
// connections and the server alive, so the reading is of a live system and not of
// what is left after one.
func connCostPoint(t *testing.T, conns int) connCostReading {
	t.Helper()
	cfg := leakServerConfig(t)
	var clients []*gws.Conn
	var taps []*leakWSTap
	var done []<-chan struct{}
	defer func() {
		for _, c := range clients {
			closeLeakWS(c)
		}
		for _, d := range done {
			expectWSClosed(t, d, "the connection the test opened")
		}
	}()

	var r connCostReading
	runServerCycle(t, cfg, func(t *testing.T, srv *Server) {
		for range conns {
			c, tap, closed := dialLeakWS(t, srv.wsListener.GetPort())
			clients = append(clients, c)
			// The tap is the client's event handler, so it has to outlive the
			// connection: dropping it here would leave the client's read loop
			// calling into nothing.
			taps = append(taps, tap)
			done = append(done, closed)
		}
		// The server has to have REGISTERED them before the heap is read, or the
		// figure is of a connection that exists on one side only.
		deadline := time.Now().Add(connCostSettleWindow)
		for srv.clientCount.Load() != int64(conns) {
			if time.Now().After(deadline) {
				t.Fatalf("the server registered %d of the %d connections within %s",
					srv.clientCount.Load(), conns, connCostSettleWindow)
			}
			time.Sleep(connCostSettlePoll)
		}
		heap, stacks := connCostSnapshot()
		r.conns = conns
		r.heap = heap
		r.stacks = stacks
		r.goroutines = runtime.NumGoroutine()
		runtime.KeepAlive(clients)
		runtime.KeepAlive(taps)
		runtime.KeepAlive(srv)
	})
	return r
}

// connCostMin runs a point connCostPasses times and keeps the smallest heap: the
// minimum is the reading where nothing else in the process was allocating, and a
// real per-connection cost is in every pass.
func connCostMin(t *testing.T, conns int) connCostReading {
	t.Helper()
	best := connCostReading{heap: ^uint64(0)}
	for range connCostPasses {
		if got := connCostPoint(t, conns); got.heap < best.heap {
			best = got
		}
	}
	return best
}

// TestLiveWebSocketConnectionCostIsBounded is the per-connection slope.
//
// It uses the WebSocket port because that is the transport a real observer and a
// real player both arrive on (webSocketPort is the port every committed capture
// uses), and because it is the expensive one: the bsPort client reads straight
// off its own socket while this one goes through net/http and gws. The bound is
// the WebSocket one, so a client on bsPort can only be cheaper.
func TestLiveWebSocketConnectionCostIsBounded(t *testing.T) {
	readings := make([]connCostReading, len(connCostCounts))
	for i, n := range connCostCounts {
		readings[i] = connCostMin(t, n)
	}

	t.Logf("live connection cost: webSocketPort, %d passes (min of each), %d points of %v",
		connCostPasses, len(readings), connCostCounts)
	for _, r := range readings {
		t.Logf("%4d connections: live heap %7d KiB, live stacks %6d KiB, %d goroutines",
			r.conns, r.heap>>10, r.stacks>>10, r.goroutines)
	}

	control := readings[0]
	if control.conns != 0 {
		t.Fatalf("the control of this measurement holds %d connections; the whole file is a differential "+
			"against a server with none, and a control with connections in it measures the connections twice",
			control.conns)
	}
	if len(readings) < 3 {
		t.Fatalf("the sweep has %d points, too few to fit a slope", len(readings))
	}

	// --- the slope, in heap and in heap plus stacks -------------------------
	// Least squares over the three points, which is the same shape of statement
	// as the peer sweep's: not "32 connections use N bytes" but "each connection
	// costs about what a connection's buffers and goroutines cost".
	var heapSlope, totalSlope float64
	points := 0
	for _, r := range readings {
		if r.conns == 0 {
			continue // the control is the origin of the slope, not a point on it
		}
		heapSlope += float64(r.heap-control.heap) / float64(r.conns)
		totalSlope += float64(r.heap-control.heap+r.stacks-control.stacks) / float64(r.conns)
		points++
	}
	heapSlope /= float64(points)
	totalSlope /= float64(points)
	t.Logf("per connection: %.1f KiB of live heap and %.1f KiB of heap plus stacks, against bounds of %d KiB "+
		"and %d KiB", heapSlope/1024, totalSlope/1024, connCostMaxHeapPerConn>>10, connCostMaxTotalPerConn>>10)

	// --- and every point really had its connections -------------------------
	for _, r := range readings {
		if r.heap == 0 {
			t.Errorf("%d connections: the live heap read 0, so the point measured nothing", r.conns)
		}
	}
	// The goroutine count is the shape of the cost: two per connection (gws's
	// ReadLoop and net/http's per-connection goroutine), which is what
	// SESSION-JOURNAL.md 35 measured as residue-free. A connection that owned
	// three would be a leak of a goroutine per client wearing a hat.
	for _, r := range readings {
		if r.conns == 0 {
			continue
		}
		perConn := float64(r.goroutines-control.goroutines) / float64(r.conns)
		if perConn > 3.5 {
			t.Errorf("%d connections moved the goroutine count from %d to %d, %.1f goroutines a connection "+
				"where two is the shape (gws's ReadLoop at ws_listener.go:162 and net/http's per-connection "+
				"goroutine)", r.conns, control.goroutines, r.goroutines, perConn)
		}
	}

	if heapSlope > connCostMaxHeapPerConn {
		t.Errorf("a live webSocketPort connection costs %.1f KiB of live heap, over the %d KiB the bound allows. "+
			"The buffers this transport is known to allocate are 2 KiB of gws read and write buffer "+
			"(ws_listener.go:127) and about 8 KiB of net/http bufio around the socket; a connection costing "+
			"several times that is a per-connection buffer sized for something else",
			heapSlope/1024, connCostMaxHeapPerConn>>10)
	}
	if totalSlope > connCostMaxTotalPerConn {
		t.Errorf("a live webSocketPort connection costs %.1f KiB of live heap and stacks together, over the %d "+
			"KiB the bound allows. Two goroutines own the socket and their stacks are not in HeapAlloc at all, "+
			"so a connection that leaked a stack would only show here",
			totalSlope/1024, connCostMaxTotalPerConn>>10)
	}
	// There is deliberately NO lower bound on the stack figure, and the reason is
	// worth writing down: StackInuse counts whole 8 KiB spans, so it moves in
	// steps, and a span freed anywhere in the process takes the reading DOWN with
	// it. Measured across three runs of this file on the same build, the control's
	// own StackInuse read 288, 960 and 1504 KiB with the same six goroutines,
	// which puts the per-connection stack figure anywhere between 7 and 20 KiB. A
	// bound that tight would be a coin toss. An upper bound, which is what a
	// leaked stack blows through, is not.
}
