package internal

import (
	"fmt"
	"waguri-centralized-control/packages/go-utils/config"
)

// DNSConfig embeds the base config and adds DNS-specific fields
type DNSConfig struct {
	config.Config `yaml:",inline"`
	Domains       map[string]string `yaml:"domains"`
}

// LoadDNSConfig loads DNS-specific configuration
func LoadDNSConfig(pathOrURL string) (*DNSConfig, error) {
	cfg := &DNSConfig{}
	err := config.Load(pathOrURL, cfg)
	if err != nil {
		return nil, err
	}

	// Validate that domains are configured
	if err := validateDNSConfig(cfg); err != nil {
		return nil, fmt.Errorf("DNS configuration validation failed: %w", err)
	}

	return cfg, nil
}

// validateDNSConfig ensures the DNS configuration is valid
func validateDNSConfig(cfg *DNSConfig) error {
	if len(cfg.Domains) == 0 {
		return fmt.Errorf("no domains configured")
	}

	for domain, ip := range cfg.Domains {
		if domain == "" {
			return fmt.Errorf("empty domain name found")
		}
		if ip == "" {
			return fmt.Errorf("domain '%s' has empty IP address", domain)
		}
	}

	return nil
}
