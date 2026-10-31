package main

import (
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"
)

// WHAT THIS FILE IS
//
// The client-facing port is a number the rest of the proxy has to believe, and it
// used to be a guess: main asked the OS for a free port by binding ":0", read the
// number and closed it, and the listener bound that number much later, from
// proxy.Start(). Between the two the port belonged to nobody, and on Windows a
// listener bind is exclusive, so whatever takes it in that window keeps it and
// the capture dies on "could not bind to TCP address" before it records a frame.
// The relay broadcast made it worse: it advertised the guessed number to the game
// while nothing was listening on it yet.
//
// Bind closes the window by binding first and reporting what it got, so the number
// is a real one this process already owns by the time it is advertised. Two things
// have to hold and are asserted here: the port that comes back is the one that is
// listening, and this process still owns it (a second bind on it fails).
//
// Falsification: make Bind pick a port the old way, or make main call Start before
// Bind, and the ownership assertion is what stops holding.

func TestBindAssignsAndKeepsTheClientPort(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		ProxyHost:        "127.0.0.1",
		ProxyPort:        0, // the OS assigns it, which is what no -bsPort means
		MainLogPath:      filepath.Join(dir, MainLogName),
		SecondaryLogPath: filepath.Join(dir, SecondaryLogName),
	}
	main := NewCSVLogger(cfg.MainLogPath)
	defer func() { _ = main.Close() }()
	secondary := NewCSVLogger(cfg.SecondaryLogPath)
	defer func() { _ = secondary.Close() }()

	s, err := NewTCPProxyServer(cfg, main, secondary)
	if err != nil {
		t.Fatalf("NewTCPProxyServer: %v", err)
	}
	if err := s.Bind(); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	defer s.Stop()

	port := s.ClientPort()
	t.Logf("client-facing port assigned at bind time: %d", port)
	if port == 0 {
		t.Fatal("ClientPort is 0 after Bind: nothing reports the port the game has to connect to")
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	// The number is owned. A second bind on it has to fail, and that is the whole
	// difference against a port picked by binding and letting go: there, this bind
	// would succeed and the two would fight over it.
	other, err := net.Listen("tcp", addr)
	if err == nil {
		_ = other.Close()
		t.Fatalf("bound %s a second time: the proxy advertised a port it does not own, which is the race this "+
			"replaces", addr)
	}

	// And it is listening, not merely recorded.
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("dial the port Bind reported (%s): %v", addr, err)
	}
	_ = conn.Close()
}

// TestClientPortIsTheConfiguredOneWhenThereIsOne keeps the other half: an operator
// who names a port gets that port, and the TLS decision in handleConnection, which
// compares the connection's listen port against this one, follows it.
//
// This is the one place in this package where port 0 cannot be used, and the
// reason is the claim itself: what is being asserted is that a CONFIGURED number
// is the number that ends up listening, and asking the OS would compare the OS's
// answer with itself. So the number is picked by binding ":0" and letting go, and
// that leaves a window in which something else can take it (on Windows a listener
// bind is exclusive, so the taker keeps it and Bind fails). The window is closed
// the only other way it can be closed: check the port, bind, and if the bind lost
// the race take another one, with the count written down. It used to be a
// t.Skipf, which is the same fact reported as "this test did not run".
func TestClientPortIsTheConfiguredOneWhenThereIsOne(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		ProxyHost:        "127.0.0.1",
		MainLogPath:      filepath.Join(dir, MainLogName),
		SecondaryLogPath: filepath.Join(dir, SecondaryLogName),
	}
	main := NewCSVLogger(cfg.MainLogPath)
	defer func() { _ = main.Close() }()
	secondary := NewCSVLogger(cfg.SecondaryLogPath)
	defer func() { _ = secondary.Close() }()

	for attempt := range releasedTCPPortCandidates {
		configured, refused, err := takeUnusedPort(func() int { return releasedTCPPort(t) }, releasedTCPPortCandidates)
		if err != nil {
			t.Fatalf("no port stayed free after refusing %d candidate(s): %v", refused, err)
		}
		cfg.ProxyPort = configured

		s, err := NewTCPProxyServer(cfg, main, secondary)
		if err != nil {
			t.Fatalf("NewTCPProxyServer: %v", err)
		}
		if err := s.Bind(); err != nil {
			t.Logf("attempt %d: Bind could not take the port %d, which had just answered no connection: %v. A number "+
				"that was free a moment ago is not guaranteed to stay free, which is the reason nothing in this package "+
				"asks for one that way; taking another", attempt+1, configured, err)
			continue
		}

		if got := s.ClientPort(); got != configured {
			t.Errorf("ClientPort = %d for a ProxyPort configured as %d. The advertised port has to be the one that is "+
				"listening", got, configured)
		}
		s.Stop()
		return
	}
	t.Fatalf("Bind lost the port on all of %d attempts. Something on this machine is binding and releasing ephemeral "+
		"ports as fast as they are handed out", releasedTCPPortCandidates)
}