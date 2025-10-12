package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"waguri-centralized-control/dns/internal"

	"waguri-centralized-control/packages/go-utils/telemetry"
)

func main() {
	// Get config URL from environment variable - mandatory
	configURL := os.Getenv("CONFIG_URL")
	if configURL == "" {
		log.Fatal("CONFIG_URL environment variable is required")
	}

	log.Println("Loading DNS config from:", configURL)

	// Load DNS-specific configuration
	cfg, err := internal.LoadDNSConfig(configURL)
	if err != nil {
		log.Fatalf("Failed to load DNS config: %v", err)
	}

	// Initialize logger with telemetry output setting and header
	logger := telemetry.NewLogger(cfg.Telemetry.Output, cfg.Telemetry.Header)

	// Log startup information
	logger.Info("DNS server starting on", cfg.Listen)
	logger.Info("Configured domains:", len(cfg.Domains))

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
