package config

import (
	"encoding/json"
	"fmt"
	"minecraft-manager/internal/paths"
	"os"
	"strconv"
)

type Config struct {
	Name               string   `json:"name"`
	Type               string   `json:"type"`
	Version            string   `json:"version"`
	Java               string   `json:"java"`
	Memory             string   `json:"memory"`
	Jar                string   `json:"jar"`
	Port               string   `json:"port"`
	LevelName          string   `json:"level"`
	AutomaticRestarts  bool     `json:"autorestart"`
	StartOnBoot        bool     `json:"boot"`
	AdditionalJVMArgs  []string `json:"additionaljvmargs"`
	AdditionalServArgs []string `json:"additionalservargs"`
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
	path := paths.Config(server)

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
	path := paths.Config(server)

	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func updateAdditionalArgs(server string, fn func(*Config) error) error {
	configFilePath := paths.Config(server)

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	if err := fn(&cfg); err != nil {
		return err
	}

	data, err = json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFilePath, data, 0644)
}

func removeAt(args []string, index int) ([]string, error) {
	if index < 0 || index >= len(args) {
		return nil, fmt.Errorf("index %d out of range", index+1)
	}

	fmt.Printf("Removing argument at index %d, which is %s\n", index+1, args[index])

	return append(args[:index], args[index+1:]...), nil
}

func addArg(args []string, arg string) []string {

	fmt.Printf("Adding argument %s\n", arg)

	return append(args, arg)
}

func AddAdditionalJVMArg(server, arg string) error {

	return updateAdditionalArgs(server, func(cfg *Config) error {

		cfg.AdditionalJVMArgs = addArg(cfg.AdditionalJVMArgs, arg)
		return nil
	})
}

func AddAdditionalServArg(server, arg string) error {
	return updateAdditionalArgs(server, func(cfg *Config) error {

		cfg.AdditionalServArgs = addArg(cfg.AdditionalServArgs, arg)
		return nil
	})
}

func RemoveAdditionalJVMArg(server, argIndex string) error {
	return updateAdditionalArgs(server, func(cfg *Config) error {
		index, err := strconv.Atoi(argIndex)
		if err != nil {
			return err
		}

		index--

		cfg.AdditionalJVMArgs, err = removeAt(cfg.AdditionalJVMArgs, index)
		return err
	})
}

func RemoveAdditionalServArg(server, argIndex string) error {
	return updateAdditionalArgs(server, func(cfg *Config) error {
		index, err := strconv.Atoi(argIndex)
		if err != nil {
			return err
		}

		index--

		cfg.AdditionalServArgs, err = removeAt(cfg.AdditionalServArgs, index)
		return err
	})
}
