package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Listen    string `yaml:"listen"`
	Telemetry struct {
		Output string `yaml:"output"`
		Header string `yaml:"header"`
	} `yaml:"telemetry"`

	Domains map[string]string `yaml:"domains"`
}

func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &Config{}
	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
