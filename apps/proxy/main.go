package main

import (
	"log"
	"os"

	"waguri-centralized-control/packages/go-utils/telemetry"
	"waguri-centralized-control/proxy/internal"
)

func main() {
	// Get config URL from environment variable - mandatory
	configURL := os.Getenv("CONFIG_URL")
	if configURL == "" {
		log.Fatal("CONFIG_URL environment variable is required")
	}

	log.Println("Loading proxy config from:", configURL)

	// Load proxy-specific configuration
	cfg, err := internal.LoadProxyConfig(configURL)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger with telemetry output setting and header
	logger := telemetry.NewLogger(cfg.Telemetry.Output, cfg.Telemetry.Header)

	// Use the telemetry logger
	logger.Info("Proxy server starting on", cfg.Listen)
	logger.Info("Configured routes:", len(cfg.Routes))

	// Create proxy server
	server := internal.NewServer(cfg, logger)

	// Start the server
	if err := server.Start(); err != nil {
		logger.Error("Server failed:", err)
	}
}
