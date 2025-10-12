package main

import (
	"log"

	"waguri-centralized-control/packages/go-utils/telemetry"
	"waguri-centralized-control/proxy/internal"
)

func main() {
	// Load proxy-specific configuration
	cfg, err := internal.LoadProxyConfig("../../configs/proxy.yaml")
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
