package server

import (
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/luskaner/ageBattleServer/internal/loadgen"
)

var (
	loadMatches   = flag.Int("load.matches", 4, "concurrent matches (sessions) in the load test")
	loadPlayers   = flag.Int("load.players", 8, "players per match, including the host")
	loadObservers = flag.Int("load.observers", 8, "observers per match")
	loadDuration  = flag.Duration("load.duration", 20*time.Second, "load test duration")
	loadPageEvery = flag.Duration("load.pageinterval", 125*time.Millisecond, "host game-data page interval")
)

// TestLoadRealistic drives a real server through capture-derived traffic at a
// configurable scale, then prints a report. It is not a pass/fail unit test; it
// is a regression harness for throughput, latency, CPU and allocation. Run:
//
//	go test -run TestLoadRealistic -v ./internal/server/ \
//	    -load.matches=4 -load.players=8 -load.observers=24 -load.duration=30s
func TestLoadRealistic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping load test in -short mode")
	}
	silenceLogger()

	cfg := &Config{
		EphemeralListenPorts:    EphemeralWS,
		SimulationPeriod:        125,
		ObserverMax:             512,
		ObserverSocketSkipMS:    0,
		LogLevel:                "none",
	}
	srv := startServer(t, cfg)
	port := srv.wsListener.GetPort()
	if port == 0 {
		t.Fatal("webSocketPort 0 after Start: the listener is not listening on a port")
	}

	profile := loadgen.DefaultProfile()
	profile.Matches = *loadMatches
	profile.Players = *loadPlayers
	profile.Observers = *loadObservers
	profile.Duration = *loadDuration
	profile.PageInterval = *loadPageEvery

	runner, err := loadgen.New(fmt.Sprintf("ws://127.0.0.1:%d/", port), profile)
	if err != nil {
		t.Fatalf("loadgen.New: %v", err)
	}
	start := time.Now()
	runner.Debug = func(format string, args ...any) {
		t.Logf("%6.2fs "+format, append([]any{time.Since(start).Seconds()}, args...)...)
	}

	rep := runner.Run(t.Context())

	counters := srv.MetricsCollector().GetAllCounters()
	rep.ServerBytesReceived = counters["BytesReceived"]
	rep.ServerBytesSent = counters["BytesSent"]

	t.Log("\n" + loadgen.FormatReport(rep))
	if rep.Stats.Errors > 0 {
		t.Logf("WARNING: %d client-side errors during the run", rep.Stats.Errors)
	}
}
