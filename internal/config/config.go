package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server        ServerConfig              `yaml:"server"`
	Admin         AdminConfig               `yaml:"admin"`
	ClientAPIKeys []string                  `yaml:"client_api_keys"`
	Providers     map[string]ProviderConfig `yaml:"providers"`
	Routes        map[string]string         `yaml:"routes"`
	Logging       LoggingConfig             `yaml:"logging"`
}

type AdminConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Password string `yaml:"password"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type ProviderConfig struct {
	Type             string   `yaml:"type"`     // openai, anthropic, google
	BaseURL          string   `yaml:"base_url"`
	APIKeys          []string `yaml:"api_keys"`
	RotationStrategy string   `yaml:"rotation_strategy"` // round_robin, random, least_used
}

type LoggingConfig struct {
	Level        string `yaml:"level"`
	Format       string `yaml:"format"`
	LogRequests  bool   `yaml:"log_requests"`
	LogResponses bool   `yaml:"log_responses"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}

	// Set default rotation strategy for providers
	for name, provider := range cfg.Providers {
		if provider.RotationStrategy == "" {
			provider.RotationStrategy = "round_robin"
		}
		cfg.Providers[name] = provider
	}

	return &cfg, nil
}
