package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config contains common configuration fields shared across all apps
type Config struct {
	Listen    string    `yaml:"listen"`
	Telemetry Telemetry `yaml:"telemetry"`
}

// Telemetry represents telemetry configuration
type Telemetry struct {
	Output string `yaml:"output"`
	Header string `yaml:"header"`
}

// Load loads a YAML configuration file into any struct
func Load(pathOrURL string, target interface{}) error {
	var data []byte
	var err error

	// Check if it's a URL (starts with http:// or https://)
	if strings.HasPrefix(pathOrURL, "http://") || strings.HasPrefix(pathOrURL, "https://") {
		data, err = loadFromURL(pathOrURL)
	} else {
		data, err = loadFromFile(pathOrURL)
	}

	if err != nil {
		return fmt.Errorf("failed to load config from %s: %w", pathOrURL, err)
	}

	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return nil
}

// loadFromFile loads configuration from a local file
func loadFromFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// loadFromURL loads configuration from a remote URL
func loadFromURL(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch config: HTTP %d %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// LoadBase loads base configuration (telemetry and listen)
func LoadBase(path string) (*Config, error) {
	cfg := &Config{}
	err := Load(path, cfg)
	return cfg, err
}
