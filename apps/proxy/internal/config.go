package internal

import (
	"fmt"
	"strings"
	"waguri-centralized-control/packages/go-utils/config"
)

type ProxyConfig struct {
	config.Config `yaml:",inline"`
	Routes        []RoutesConfig `yaml:"routes"`
	Menu          string         `yaml:"menu"`
}

type RoutesConfig struct {
	Host        string `yaml:"host" validate:"required"`
	Target      string `yaml:"target" validate:"required"`
	Name        string `yaml:"name" validate:"required"`
	Description string `yaml:"description" validate:"required"`
	Icon        string `yaml:"icon" validate:"required"`
	Category    string `yaml:"category" validate:"required"`
}

// IsRedirect checks if this route should trigger a redirect instead of proxy
func (r *RoutesConfig) IsRedirect() bool {
	return strings.HasPrefix(r.Target, "rhttp://") || strings.HasPrefix(r.Target, "rhttps://")
}

// GetRedirectURL returns the actual redirect URL by removing the 'r' prefix
func (r *RoutesConfig) GetRedirectURL() string {
	if strings.HasPrefix(r.Target, "rhttp://") {
		return "http://" + r.Target[8:] // Remove "rhttp://"
	}
	if strings.HasPrefix(r.Target, "rhttps://") {
		return "https://" + r.Target[9:] // Remove "rhttps://"
	}
	return r.Target
}

func LoadProxyConfig(path string) (*ProxyConfig, error) {
	cfg := &ProxyConfig{}
	err := config.Load(path, cfg)
	if err != nil {
		return nil, err
	}

	// Validate that all routes have required fields
	if err := validateProxyConfig(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// validateProxyConfig ensures all routes have required fields
func validateProxyConfig(cfg *ProxyConfig) error {
	for i, route := range cfg.Routes {
		if route.Host == "" {
			return fmt.Errorf("route %d: host is required", i)
		}
		if route.Target == "" {
			return fmt.Errorf("route %d (%s): target is required", i, route.Host)
		}
		if route.Name == "" {
			return fmt.Errorf("route %d (%s): name is required", i, route.Host)
		}
		if route.Description == "" {
			return fmt.Errorf("route %d (%s): description is required", i, route.Host)
		}
		if route.Icon == "" {
			return fmt.Errorf("route %d (%s): icon is required", i, route.Host)
		}
		if route.Category == "" {
			return fmt.Errorf("route %d (%s): category is required", i, route.Host)
		}
	}
	return nil
}
