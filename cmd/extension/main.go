package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mumzworld-tech/lambdawatch/internal/config"
	"github.com/mumzworld-tech/lambdawatch/internal/extension"
	"github.com/mumzworld-tech/lambdawatch/internal/logger"
)

func main() {
	logger.Init()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Validate required config
	if cfg.LokiEndpoint == "" {
		logger.Fatal("LOKI_URL environment variable is required")
	}

	// Setup context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		logger.Infof("Received signal: %v", sig)
		cancel()
	}()

	// Create and run the extension
	mgr := extension.NewManager(cfg)
	if err := mgr.Run(ctx); err != nil {
		logger.Fatalf("Extension error: %v", err)
	}
}
