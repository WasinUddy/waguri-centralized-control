package internal

import (
	"waguri-centralized-control/packages/go-utils/config"
)

// DNSConfig embeds the base config and adds DNS-specific fields
type DNSConfig struct {
	config.Config `yaml:",inline"`
	Domains       map[string]string `yaml:"domains"`
}

// LoadDNSConfig loads DNS-specific configuration
func LoadDNSConfig(path string) (*DNSConfig, error) {
	cfg := &DNSConfig{}
	err := config.Load(path, cfg)
	return cfg, err
}
