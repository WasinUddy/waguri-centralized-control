package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config contains common configuration fields shared across all apps
type Config struct {
	Listen    string `yaml:"listen"`
	Telemetry struct {
		Output string `yaml:"output"`
		Header string `yaml:"header"`
	} `yaml:"telemetry"`
}

// Load loads a YAML configuration file into any struct
func Load(path string, target interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return yaml.NewDecoder(file).Decode(target)
}

// LoadBase loads base configuration (telemetry and listen)
func LoadBase(path string) (*Config, error) {
	cfg := &Config{}
	err := Load(path, cfg)
	return cfg, err
}
