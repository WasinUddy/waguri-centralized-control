package internal

import "waguri-centralized-control/packages/go-utils/config"

type ProxyConfig struct {
	config.Config `yaml:",inline"`
	Routes        []RoutesConfig `yaml:"routes"`
	Menu          string         `yaml:"menu"`
}

type RoutesConfig struct {
	Host        string `yaml:"host"`
	Target      string `yaml:"target"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Icon        string `yaml:"icon"`
	Category    string `yaml:"category"`
}

func LoadProxyConfig(path string) (*ProxyConfig, error) {
	cfg := &ProxyConfig{}
	err := config.Load(path, cfg)
	return cfg, err
}
