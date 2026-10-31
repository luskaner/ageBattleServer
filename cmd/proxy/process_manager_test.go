package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"testing"
	"time"
)

// The launcher passes "-bsPort <port>" to both children, and Start waits for that
// port to accept before moving on. If the flag cannot be recovered the wait is
// skipped rather than probing something meaningless.
func TestArgInt(t *testing.T) {
	cases := []struct {
		name string
		args []string
		flag string
		want int
	}{
		{"bsPort present", []string{"-relaybroadcastPort", "0", "-bsPort", "14001", "-region", "Server"}, "bsPort", 14001},
		{"double dash", []string{"--bsPort", "15001"}, "bsPort", 15001},
		{"wrong case", []string{"-bSport", "15001"}, "bsPort", 15001},
		{"absent", []string{"-region", "Server"}, "bsPort", 0},
		{"last flag with no value", []string{"-region", "Server", "-bsPort"}, "bsPort", 0},
		{"not a number", []string{"-bsPort", "abc"}, "bsPort", 0},
		{"empty", nil, "bsPort", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := argInt(tc.args, tc.flag); got != tc.want {
				t.Errorf("argInt(%v) = %d, want %d", tc.args, got, tc.want)
			}
		})
	}
}

func TestListenAddr(t *testing.T) {
	if got := listenAddr("127.0.0.1", 14001); got != "127.0.0.1:14001" {
		t.Errorf("listenAddr = %q, want 127.0.0.1:14001", got)
	}
	for _, tc := range []struct {
		host string
		port int
	}{
		{"", 14001},
		{"127.0.0.1", 0},
		{"127.0.0.1", -1},
		{"127.0.0.1", 70000},
	} {
		if got := listenAddr(tc.host, tc.port); got != "" {
			t.Errorf("listenAddr(%q, %d) = %q, want \"\" so the wait is skipped", tc.host, tc.port, got)
		}
	}
}

// waitForPort returns as soon as the address accepts, and gives up with a warning
// once the timeout passes so a child that never binds cannot hang startup.
func TestWaitForPortReturnsWhenListening(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			_ = c.Close()
		}
	}()

	pm := &ProcessManager{logOut: os.Stderr}
	addr := ln.Addr().String()

	done := make(chan struct{})
	go func() {
		pm.waitForPort("test child", addr)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("waitForPort did not return for a listening address %s", addr)
	}
}

func TestWaitForPortSkipsEmptyAddr(t *testing.T) {
	pm := &ProcessManager{logOut: os.Stderr}
	done := make(chan struct{})
	go func() {
		pm.waitForPort("test child", "")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("waitForPort must return immediately when there is nothing to wait for")
	}
}

// releasedTCPPortCandidates is how many ports the two tests that need a RELEASED
// port try before they declare the machine the problem. It is a CAP and not a
// hope: an unbounded retry would turn a machine that hands ephemeral ports
// straight back to whoever asks for one into a test that never finishes.
const releasedTCPPortCandidates = 8

// releasedTCPPort returns a port number this process has just bound and let go,
// which is the only shape a test that needs a port it is not going to keep can
// use. It does not claim the port is still free; takeUnusedPort is what checks.
func releasedTCPPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for a port to release: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	return port
}

// takeUnusedPort asks candidate for a port number until it gets one that nothing
// is listening on, and returns it with the number of candidates it had to refuse.
//
// The bind that produces the number cannot promise the port stays free, and this
// repository has already paid for that once: on Windows a listener bind is
// exclusive, so whatever takes a released port in the window owns it, and the
// caller dies on "could not bind to TCP address" (SESSION-JOURNAL.md 38). Asking
// the OS at bind time removes the window, and the two tests that genuinely cannot
// do that are here, so the window is closed the only other way it can be: the
// port is CHECKED, and a port whose "nothing is listening on it" cannot be
// established is refused rather than handed over.
//
// A refused connect is the property both callers need and the one a bind does not
// carry. The attempts are the cap, so a machine that is taking every ephemeral
// port as fast as it is handed out fails the test instead of hanging it.
func takeUnusedPort(candidate func() int, attempts int) (port, refused int, err error) {
	for range attempts {
		p := candidate()
		conn, dialErr := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(p)), time.Second)
		if dialErr == nil {
			_ = conn.Close()
			refused++
			continue
		}
		return p, refused, nil
	}
	return 0, refused, fmt.Errorf("nothing was listening on %d consecutive ports, so this machine is binding and "+
		"releasing them as fast as they are handed out", refused)
}

// unusedTCPPort is takeUnusedPort for the tests in this file: a port on
// 127.0.0.1 that nothing answers on, or a failure that says why.
func unusedTCPPort(t *testing.T) int {
	t.Helper()
	port, refused, err := takeUnusedPort(func() int { return releasedTCPPort(t) }, releasedTCPPortCandidates)
	if err != nil {
		t.Fatalf("no port stayed free after refusing %d candidate(s): %v", refused, err)
	}
	if refused > 0 {
		t.Logf("%d candidate port(s) were taken between the bind and the check", refused)
	}
	return port
}

func TestWaitForPortGivesUp(t *testing.T) {
	// The port has to be one nothing ever binds, which is the whole claim of this
	// test, and that is the one property port 0 cannot deliver: asking the OS for
	// a port and keeping it would make waitForPort return at once. So the number
	// is picked by binding ":0" and letting go, and unusedTCPPort checks it.
	port := unusedTCPPort(t)
	t.Logf("a port nothing is listening on: %d", port)

	// Shrink the deadline so the test does not sit for the production timeout.
	pm := &ProcessManager{logOut: os.Stderr, readyTimeout: 200 * time.Millisecond}
	start := time.Now()
	pm.waitForPort("test child", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if elapsed := time.Since(start); elapsed < pm.readyTimeout {
		t.Errorf("waitForPort returned after %s on port %d, want at least the timeout %s. Something answered on a port "+
			"this test established nobody was listening on, and the test's premise is what failed, not the timeout",
			elapsed, port, pm.readyTimeout)
	}
}

// TestTakeUnusedPortRefusesAPortInUse is the falsification of the check, and it
// drives it with a port that is DEFINITELY in use instead of hoping for a race:
// a listener this test holds, handed over as the candidate. The helper has to
// refuse it and go on to the next candidate, and when every candidate is in use it
// has to say so rather than return a port it never verified.
func TestTakeUnusedPortRefusesAPortInUse(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			_ = c.Close()
		}
	}()
	busy := ln.Addr().(*net.TCPAddr).Port

	candidates := 0
	free := releasedTCPPort(t)
	port, refused, err := takeUnusedPort(func() int {
		candidates++
		if candidates == 1 {
			return busy // in use: has to be refused
		}
		return free
	}, releasedTCPPortCandidates)
	if err != nil {
		t.Fatalf("takeUnusedPort gave up with one perfectly good candidate left: %v", err)
	}
	if refused != 1 {
		t.Errorf("takeUnusedPort refused %d candidate(s), want 1: the port %d was being listened on and the next one "+
			"was not", refused, busy)
	}
	if port != free {
		t.Errorf("takeUnusedPort returned %d, want the candidate that was free (%d). It returned the port it was "+
			"handed without checking it", port, free)
	}

	// And the cap: every candidate in use has to end in an error, not in a port.
	_, refused, err = takeUnusedPort(func() int { return busy }, releasedTCPPortCandidates)
	if err == nil {
		t.Fatal("takeUnusedPort returned a port after refusing every candidate, so a caller would go on to bind a port " +
			"it never established was free")
	}
	if refused != releasedTCPPortCandidates {
		t.Errorf("takeUnusedPort refused %d candidate(s) before giving up, want the %d it was allowed",
			refused, releasedTCPPortCandidates)
	}
	t.Logf("%d candidates in use in a row: %v", refused, err)
}

func TestChildReadyConstants(t *testing.T) {
	if childReadyTimeout <= 0 {
		t.Error("childReadyTimeout must be positive")
	}
	if childReadyPoll <= 0 {
		t.Error("childReadyPoll must be positive")
	}
	if childReadyPoll >= childReadyTimeout {
		t.Error("the poll interval must be well under the timeout")
	}
	_ = log.Flags() // keep the log import honest across build tags
}
