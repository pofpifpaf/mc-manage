package client

import (
	"errors"
	"fmt"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/download"
	"minecraft-manager/internal/java"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"minecraft-manager/internal/ui"
	"os"
	"strconv"
	"strings"
	"time"
)

func SetParameter(args []string) error {

	var server, arg1, arg2, arg3 string
	switch {
	case len(args) == 5:
		server = args[2]
		arg1 = args[3]
		arg2 = args[4]
		arg3 = ""
	case len(args) >= 6:
		server = args[2]
		arg1 = args[3]
		arg2 = args[4]
		arg3 = args[5]
		for _, arg := range args[6:] {
			arg3 = arg3 + "-" + arg
		}
	default:
		return fmt.Errorf("Invalid number of arguments: %d", len(args))
	}

	valid, server := paths.ValidateServerName(server)
	if !valid {
		return fmt.Errorf("Invalid server name %s", server)
	}

	setForce := false
	if strings.Contains(arg3, "--force") {
		arg3 = strings.Replace(arg3, "---force", "", 1)
		arg3 = strings.Replace(arg3, "--force", "", 1)
		setForce = true
	}

	fmt.Print("\n")
	defer fmt.Print("\n")

	isServerRunning, err := daemonIsServerRunning(server)
	if err != nil {
		return err
	}
	if isServerRunning != protocol.StateStopped && arg1 != "autorestart" {
		return fmt.Errorf("Unable to use set, server %s is already running", server)
	}

	switch arg1 {
	case "port":

		if err := setServerPort(server, arg2); err != nil {
			return err
		}

	case "autorestart":

		if err := setServerAutoRestart(server, arg2); err != nil {
			return err
		}

	case "boot":

		if err := setStartOnBoot(server, arg2); err != nil {
			return err
		}

	case "motd":

		if err := config.SetServerProperty(paths.ServerProperties(server), config.MotdKey, arg2); err != nil {
			return err
		}

	case "version":

		if err := setServerVersion(server, arg2, arg3, setForce); err != nil {
			return err
		}

	case "java":

		if err := setJavaVersion(server, arg2); err != nil {
			return err
		}

	case "world":

		if err := setWorldName(server, arg2); err != nil {
			return err
		}

	case "mem-allocated", "xms", "mem-a":

		if err := setAllocatedMemory(server, arg2); err != nil {
			return err
		}

	case "mem-max", "xmx", "mem-m":

		if err := setMaxMemory(server, arg2); err != nil {
			return err
		}

	case "runasroot":

		switch arg2 {
		case "true":
			ui.PrintWarning(fmt.Sprintf("Running %s as root: This operation is not recommended", server))
			if err := config.SetUserSpecificFalse(server); err != nil {
				return err
			}
		case "false":
			if err := ResumeFromRootConfig(server); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unrecognized parameter %q, try \"true\"/\"false\"", arg2)
		}

	default:
		return fmt.Errorf("Incorrect set parameter %s, see manager set help for more information", arg1)
	}

	ui.PrintSuccess("Successfully changed parameter " + arg1 + " for server " + server + " to \"" + arg2 + "\"")

	return nil
}

func setStartOnBoot(name, value string) error {

	cfg, err := config.Load(name)
	if err != nil {
		return err
	}

	switch value {
	case "true":
		cfg.StartOnBoot = true
	case "false":
		cfg.StartOnBoot = false
	default:
		return fmt.Errorf("Unable to set paramter: Invalid value parameter %s", value)
	}

	return config.Save(name, cfg)
}

func setServerPort(name, port string) error {

	portInt, err := strconv.Atoi(port)
	if err != nil || (portInt < 0 || portInt > 65535) {
		return errors.New("port number out of range, must be between 0 and 65535")
	}

	err = config.SetServerProperty(paths.ServerProperties(name), "server-port", port)
	if err != nil {
		return err
	}

	cfg, err := config.Load(name)
	if err != nil {
		return err
	}

	cfg.Port = port

	return config.Save(name, cfg)
}

func setServerAutoRestart(name, autoRestart string) error {
	if autoRestart == "true" || autoRestart == "false" {
		return send(protocol.Request{
			Command: "SET",
			Server:  name,
			Text:    "autorestart",
			Data:    autoRestart,
		})
	} else {
		return fmt.Errorf("unrecognized argument to autorestart")
	}
}

func setServerVersion(name, serverVersion, serverVersionArg string, force bool) error {

	cfg, err := config.Load(name)
	if err != nil {
		return err
	}

	oldServerVersion := cfg.Version
	oldServerVersionArg := cfg.VersionArg
	oldJVMArg := cfg.AdditionalJVMArgs

	if cfg.Version == serverVersion && cfg.VersionArg == serverVersionArg && !force {
		return fmt.Errorf("Version for server %s is already %q (use \"--force\" to bypass)", name, serverVersion)
	}

	if err := download.ArchiveJarFile(cfg); err != nil && cfg.Type != "neoforge" && cfg.Type != "forge" && !force {
		ui.PrintWarning("Unable to archive old jar file")
	}

	if err := download.RemoveRecommendedJVMArguments(cfg); err != nil {
		ui.PrintWarning("Couldn't remove old recommended jvm arguments: " + err.Error())
	}

	cfg.Version = serverVersion
	cfg.VersionArg = serverVersionArg

	if err := download.DownloadJar(cfg); err != nil {
		ui.PrintError("Unable to find version \"" + serverVersion + "\", undoing version changes...")
		cfg.Version = oldServerVersion
		cfg.VersionArg = oldServerVersionArg
		cfg.AdditionalJVMArgs = oldJVMArg
		download.RetrieveJarIfArchived(cfg)
		return err
	}

	return config.Save(cfg.Name, cfg)
}

func setJavaVersion(name, javaVersion string) error {
	if _, err := java.Find(javaVersion); err != nil {
		return err
	}

	cfg, err := config.Load(name)
	if err != nil {
		return err
	}

	cfg.Java = javaVersion

	if err := config.Save(cfg.Name, cfg); err != nil {
		return err
	}

	switch cfg.Type {
	case "neoforge", "forge":
		return config.ConfigureJavaRunScript(name)
	default:
		return nil
	}
}

func setWorldName(name, worldName string) error {

	serverPropertiesPath := paths.ServerProperties(name)

	if fileInfo, err := os.Stat(serverPropertiesPath); err != nil || fileInfo.IsDir() {
		return fmt.Errorf("unable to set world parameter: server.properties doesn't exist or is a directory")
	}

	return config.SetServerProperty(serverPropertiesPath, config.LevelNamePropertyKey, worldName)
}

func setAllocatedMemory(server, mem string) error {

	if err := config.ValidateMemoryConfig(mem); err != nil {
		return err
	}

	cfg, err := config.Load(server)
	if err != nil {
		return err
	}

	cfg.MemoryAllocated = mem

	if err := config.Save(server, cfg); err != nil {
		return err
	}

	switch cfg.Type {
	case "neoforge", "forge":
		return config.ScriptSetJVMMemoryArgs(server)
	default:
		return nil
	}
}

func setMaxMemory(server, mem string) error {

	if err := config.ValidateMemoryConfig(mem); err != nil {
		return err
	}

	cfg, err := config.Load(server)
	if err != nil {
		return err
	}

	cfg.MemoryMax = mem

	if err := config.Save(server, cfg); err != nil {
		return err
	}

	switch cfg.Type {
	case "neoforge", "forge":
		return config.ScriptSetJVMMemoryArgs(server)
	default:
		return nil
	}
}

func SetGracePeriod(period string) error {

	fmt.Print("\n")
	defer fmt.Print("\n")

	periodInt, err := strconv.Atoi(period)
	if err != nil || (periodInt < 0 || periodInt > 360) {
		return fmt.Errorf("grace period out of range, must be between 0 and 360")
	}

	var periodDuration time.Duration
	periodDuration = time.Duration(periodInt) * time.Second

	resp, err := sendProtocol(protocol.Request{
		Command: "SET",
		Server:  "name",
		Text:    "graceperiod",
		Data:    periodInt,
	})
	if err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf("NOK from daemon: %s", resp.Message)
	}

	ui.PrintSuccess(fmt.Sprintf("Grace period now set to %s", periodDuration))

	return nil
}

func SetProperty(server, key, value string) error {

	fmt.Print("\n")
	defer fmt.Print("\n")

	ui.PrintInfo(fmt.Sprintf("Setting key %s to %s in %s", key, value, paths.ServerProperties(server)))

	valid, server := paths.ValidateServerName(server)
	if !valid {
		return fmt.Errorf("Invalid server name %s", server)
	}

	if err := config.SetServerProperty(paths.ServerProperties(server), key, value); err != nil {
		return err
	}

	ui.PrintSuccess("Successfully set property")
	return nil
}
