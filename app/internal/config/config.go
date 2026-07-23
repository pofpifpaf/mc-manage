package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Name   string `json:"name"`
	Java   string `json:"java"`
	Memory string `json:"memory"`
	Jar    string `json:"jar"`
}

func (c *Config) Validate() error {

	if c.Java == "" {
		return fmt.Errorf("missing java version")
	}

	if c.Memory == "" {
		return fmt.Errorf("missing memory")
	}

	if c.Jar == "" {
		return fmt.Errorf("missing jar")
	}

	return nil
}

func Load(server string) (*Config, error) {
	path := filepath.Join("/server", server, "config.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read config: %w", err)
	}

	var cfg Config

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Save(server string, cfg *Config) error {
	path := filepath.Join("/server", server, "config.json")

	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}