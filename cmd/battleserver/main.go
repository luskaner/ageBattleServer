package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/luskaner/ageBattleServer/internal/config"
	"github.com/luskaner/ageBattleServer/internal/logger"
	"github.com/luskaner/ageBattleServer/internal/server"
)

func main() {
	cfg := config.ParseFlags()

	globalLevel, err := logger.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Invalid -logLevel: %v", err)
	}
	logger.SetGlobalLevel(globalLevel)

	if config.LogModName != "" {
		modLevel, err := logger.ParseLevel(config.LogModLevel)
		if err != nil {
			log.Fatalf("Invalid -logModLevel: %v", err)
		}
		logger.SetModule(config.LogModName, modLevel)
		logger.Debug("Main", "SetModule", "module=%s level=%s", config.LogModName, config.LogModLevel)
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	if err := srv.Run(ctx); err != nil {
		log.Fatalf("Error running server: %v", err)
	}
}
