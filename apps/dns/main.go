package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"waguri-centralized-control/dns/internal"
	"waguri-centralized-control/packages/go-utils/config"
	"waguri-centralized-control/packages/go-utils/telemetry"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger with telemetry output setting and header
	logger := telemetry.NewLogger(cfg.Telemetry.Output, cfg.Telemetry.Header)

	// Use the telemetry logger
	logger.Info("DNS server starting on", cfg.Listen)
	logger.Info("Configured domains:", cfg.Domains)

	// Create DNS server
	server := internal.NewServer(cfg, logger)

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Error("Server failed:", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sig := <-sigChan
	logger.Info("Received signal:", sig)

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown server:", err)
		os.Exit(1)
	}

	logger.Info("DNS server stopped gracefully")
}
