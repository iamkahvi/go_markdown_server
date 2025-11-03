package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "config/config.yaml"

// App mirrors the configuration structure defined in config.yaml.
type App struct {
	Addr         string        `yaml:"addr"`
	Origin       string        `yaml:"origin"`
	DocumentPath string        `yaml:"documentPath"`
	ShutdownWait time.Duration `yaml:"shutdownWait"`
}

// Load reads config/config.yaml and unmarshals its values into App.
func Load() (*App, error) {
	data, err := os.ReadFile(defaultConfigPath)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", defaultConfigPath, err)
	}

	var cfg App
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file %q: %w", defaultConfigPath, err)
	}

	return &cfg, nil
}
