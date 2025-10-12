package internal

import "waguri-centralized-control/packages/go-utils/config"

type ProxyConfig struct {
	config.Config `yaml:",inline"`
	Routes        []RoutesConfig `yaml:"routes"`
}

type RoutesConfig struct {
	Host   string `yaml:"host"`
	Target string `yaml:"target"`
}

func LoadProxyConfig(path string) (*ProxyConfig, error) {
	cfg := &ProxyConfig{}
	err := config.Load(path, cfg)
	return cfg, err
}
