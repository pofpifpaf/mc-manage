package config

import (
	"encoding/json"
	"fmt"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"minecraft-manager/internal/ui"
	"os"
	"strconv"
)

func Load(server string) (*protocol.Config, error) {
	path := paths.Config(server)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read config: %w", err)
	}

	var cfg protocol.Config

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Save(server string, cfg *protocol.Config) error {
	path := paths.Config(server)

	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func updateAdditionalArgs(server string, fn func(*protocol.Config) error) error {
	configFilePath := paths.Config(server)

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	var cfg protocol.Config
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

	ui.PrintInfo(fmt.Sprintf("Removing argument at index %d, which is %s\n", index+1, args[index]))

	return append(args[:index], args[index+1:]...), nil
}

func addArg(args []string, arg string) []string {

	ui.PrintInfo("Adding argument " + arg)

	return append(args, arg)
}

func AddAdditionalJVMArg(server, arg string) error {

	if err := updateAdditionalArgs(server, func(cfg *protocol.Config) error {

		if cfg.Type == "neoforge" {
			if err := NeoforgeAddJVMArg(server, arg); err != nil {
				return err
			}
		}

		cfg.AdditionalJVMArgs = addArg(cfg.AdditionalJVMArgs, arg)
		return nil
	}); err != nil {
		return err
	}

	ui.PrintSuccess("Added JVM Arg \"" + arg + "\" to server " + server)
	return nil
}

func AddAdditionalServArg(server, arg string) error {

	if err := updateAdditionalArgs(server, func(cfg *protocol.Config) error {

		cfg.AdditionalServArgs = addArg(cfg.AdditionalServArgs, arg)
		return nil
	}); err != nil {
		return err
	}

	ui.PrintSuccess("Added Serv Arg \"" + arg + "\" to server " + server)
	return nil
}

func RemoveAdditionalJVMArg(server, argIndex string) error {

	index, err := strconv.Atoi(argIndex)
	if err != nil {
		return err
	}

	index--

	if err := updateAdditionalArgs(server, func(cfg *protocol.Config) error {

		if cfg.Type == "neoforge" {
			if err := NeoforgeRemoveJVMArg(server, cfg.AdditionalJVMArgs, index); err != nil {
				return err
			}
		}

		cfg.AdditionalJVMArgs, err = removeAt(cfg.AdditionalJVMArgs, index)
		return err

	}); err != nil {
		return err
	}

	ui.PrintSuccess("Removed JVM Arg")
	return nil
}

func RemoveAdditionalServArg(server, argIndex string) error {

	if err := updateAdditionalArgs(server, func(cfg *protocol.Config) error {
		index, err := strconv.Atoi(argIndex)
		if err != nil {
			return err
		}

		index--

		cfg.AdditionalServArgs, err = removeAt(cfg.AdditionalServArgs, index)
		return err
	}); err != nil {
		return err
	}
	ui.PrintSuccess("Removed Serv Arg")
	return nil
}

func LoadMainConfig() (*protocol.MainConfig, error) {
	path := paths.MainConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read config: %w", err)
	}

	var cfg protocol.MainConfig

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func SaveMainConfig(cfg *protocol.MainConfig) error {

	path := paths.MainConfig()

	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
