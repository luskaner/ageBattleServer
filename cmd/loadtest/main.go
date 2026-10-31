// loadtest generates realistic game traffic against a running battleserver
// WebSocket endpoint, scaled to multiple matches with 8 players and observers.
// The traffic profile derives from the packet captures in
// battle-server-manager (main_log_*.csv / secondary_log_*.csv).
//
// Usage:
//
//	battleserver.exe -webSocketPort 47199 ...
//	loadtest -server ws://127.0.0.1:47199/ -matches 4 -players 8 -observers 24 -duration 30s
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luskaner/ageBattleServer/internal/loadgen"
)

func main() {
	var (
		server     = flag.String("server", "ws://127.0.0.1:47199/", "battleserver WebSocket endpoint")
		matches    = flag.Int("matches", 4, "concurrent matches (sessions)")
		players    = flag.Int("players", 8, "players per match, including the host")
		observers  = flag.Int("observers", 8, "observers per match")
		duration   = flag.Duration("duration", 30*time.Second, "how long to generate traffic")
		pageEvery  = flag.Duration("pageinterval", 125*time.Millisecond, "host game-data page interval")
		relayEvery = flag.Duration("relayinterval", 2*time.Second, "player relay interval")
		pingEvery  = flag.Duration("pinginterval", 10*time.Second, "keepalive ping interval")
		pageMax    = flag.Int("pagemax", 2600, "max generated game-data page size in bytes")
	)
	flag.Parse()

	profile := loadgen.DefaultProfile()
	profile.Matches = *matches
	profile.Players = *players
	profile.Observers = *observers
	profile.Duration = *duration
	profile.PageInterval = *pageEvery
	profile.RelayInterval = *relayEvery
	profile.PingInterval = *pingEvery
	profile.PageMaxBytes = *pageMax

	runner, err := loadgen.New(*server, profile)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "loadtest:", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("loadtest: %d matches x %d players + %d observers for %s against %s\n",
		profile.Matches, profile.Players, profile.Observers, profile.Duration, *server)
	rep := runner.Run(ctx)
	fmt.Print(loadgen.FormatReport(rep))

	if rep.Stats.Errors > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "loadtest: %d client-side errors\n", rep.Stats.Errors)
		os.Exit(1)
	}
}
