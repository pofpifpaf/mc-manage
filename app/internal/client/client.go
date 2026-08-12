package client

import (
	"errors"
	"fmt"
	"io/fs"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/download"
	"minecraft-manager/internal/java"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"minecraft-manager/internal/ui"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func StartServer(server string) error {
	return send(
		protocol.Request{
			Command: "START",
			Server:  os.Args[2],
		},
	)
}

func StopServer(server string) error {
	return send(
		protocol.Request{
			Command: "STOP",
			Server:  os.Args[2],
		},
	)
}

func PingDaemon() error {
	return send(
		protocol.Request{
			Command: "PING",
		},
	)
}

func GetPS() error {
	return send(
		protocol.Request{
			Command: "PS",
		},
	)
}

func GetList() error {

	allServers, err := MakeList()
	if err != nil {
		return err
	}

	ui.PrintServerList(allServers)

	return nil
}

func daemonIsServerRunning(name string) (protocol.ServerState, error) {
	resp, err := sendProtocol(protocol.Request{
		Command: "CHECK",
		Server:  name,
	})
	if err != nil {
		return protocol.StateStopped, err
	}

	if resp.OK && resp.Message == name {
		return protocol.ServerState(resp.Data.(string)), nil
	} else {
		return protocol.StateStopped, fmt.Errorf("Incorrect response from daemon")
	}
}

func MakeList() ([]protocol.ServerInfo, error) {

	var result []protocol.ServerInfo

	err := filepath.WalkDir(paths.GetServerRoot(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if _, err := os.Stat(paths.Config(d.Name())); err == nil {

				isServerRunning, _ := daemonIsServerRunning(d.Name())

				cfg, err := config.Load(d.Name())
				if err == nil {
					server := protocol.ServerInfo{
						Name:              d.Name(),
						Port:              cfg.Port,
						AutomaticRestarts: cfg.AutomaticRestarts,
						Running:           isServerRunning,
						Version:           cfg.Version,
						JavaVersion:       cfg.Java,
						StartOnBoot:       cfg.StartOnBoot,
					}
					result = append(result, server)
				}

			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func SetParameter(args []string) error {

	var server, arg1, arg2, arg3 string
	switch len(os.Args) {
	case 5:
		server = os.Args[2]
		arg1 = os.Args[3]
		arg2 = os.Args[4]
		arg3 = ""
	case 6:
		server = os.Args[2]
		arg1 = os.Args[3]
		arg2 = os.Args[4]
		arg3 = os.Args[5]
	default:
		return fmt.Errorf("Invalid number of arguments: %d", len(os.Args))
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

		if err := setServerVersion(server, arg2, arg3); err != nil {
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

func setServerVersion(name, serverVersion, serverVersionArg string) error {

	cfg, err := config.Load(name)
	if err != nil {
		return err
	}

	oldServerVersion := cfg.Version
	oldServerVersionArg := cfg.VersionArg

	if cfg.Version == serverVersion && cfg.VersionArg == serverVersionArg {
		return fmt.Errorf("Version for server %s is already %s", name, serverVersion)
	}

	if err := download.ArchiveJarFile(cfg); err != nil {
		ui.PrintWarning("Unable to archive old jar file")
	}

	cfg.Version = serverVersion
	cfg.VersionArg = serverVersionArg

	if err := download.DownloadJar(cfg); err != nil {
		ui.PrintError("Unable to find version \"" + serverVersion + "\", undoing version changes...")
		cfg.Version = oldServerVersion
		cfg.VersionArg = oldServerVersionArg
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
	case "neoforge":
		return config.NeoforgeConfigureJavaRunScript(name)
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
	case "neoforge":
		return config.NeoforgeSetJVMArg(server)
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
	case "neoforge":
		return config.NeoforgeSetJVMArg(server)
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

func DownloadJarToServer(server, downloadURL string) error {

	fmt.Print("\n")
	defer fmt.Print("\n")

	cfg, err := config.Load(server)
	if err != nil {
		return err
	}

	if err := download.DownloadCustomJar(cfg, downloadURL); err != nil {
		return err
	}

	cfg.Version = "custom"

	return config.Save(server, cfg)
}

func InspectServer(name string) error {

	resp, err := sendProtocol(protocol.Request{
		Command: "INSPECT",
		Server:  name,
	})
	if err != nil {
		return err
	}

	serverInfo := protocol.ServerInfo{}

	if !resp.OK {
		serverInfo.Running = protocol.StateStopped
	} else {
		serverInfo, err = makeServerInfoInterface(resp.Data)
	}

	cfg, err := config.Load(name)
	if err != nil {
		return err
	}

	ui.PrintInspectServer(serverInfo, cfg)

	return nil
}

func KillServer(name string) error {

	return send(protocol.Request{
		Command: "KILL",
		Server:  name,
	})
}

func getActivePlayerInformation(server protocol.ServerInfo) (protocol.ServerInfo, error) {

	status, err := protocol.GetServerStatus(server.Port)
	if err != nil {
		server.Running = protocol.StateStarting
		return server, err
	}

	server.PlayersOnline = status.Players.Online
	server.PlayersOnlineMax = status.Players.Max

	return server, nil
}
