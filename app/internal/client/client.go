package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/download"
	"minecraft-manager/internal/java"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type ServerInfo struct {
	Name              string
	Port              string
	AutomaticRestarts bool
	StartedAt         time.Time
	Running           bool
	Version           string
	JavaVersion       string
	StartOnBoot       bool
}

type ServerPSResponse struct {
	OK      bool         `json:"ok"`
	Message string       `json:"message,omitempty"`
	Data    []ServerInfo `json:"data"`
}

func makeServerInfoInterface(data []interface{}) ([]ServerInfo, error) {

	servers := make([]ServerInfo, len(data))

	for i, v := range data {
		m := v.(map[string]interface{})

		b, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}

		var server ServerInfo
		if err := json.Unmarshal(b, &server); err != nil {
			return nil, err
		}

		servers[i] = server
	}

	return servers, nil
}

func sendProtocol(req protocol.Request) (protocol.Response, error) {
	conn, err := net.Dial("unix", paths.SocketPath)
	if err != nil {
		return protocol.Response{}, err
	}

	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return protocol.Response{}, err
	}

	var resp protocol.Response

	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return protocol.Response{}, err
	}

	return resp, nil
}

func send(req protocol.Request) error {

	resp, err := sendProtocol(req)
	if err != nil {
		return err
	}

	if !resp.OK {
		return errors.New(resp.Message)
	}

	switch req.Command {

	case "PS":

		servers, _ := makeServerInfoInterface(resp.Data.([]interface{}))

		printRunningServers(servers)

	case "START", "STOP":
		fmt.Printf("Daemon responded without error - %s\n", resp.Message)

	case "SET":
		fmt.Printf("SET - %s\n", resp.Message)

	case "PING":
		if resp.Message == "PONG" {
			fmt.Println("Ping successful, daemon is running")
		} else {
			fmt.Println("Ping failure, received ", resp.Message)
		}

	default:
		fmt.Printf("Response: %q\n", resp.Message)
	}

	return nil
}

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

	printServerList(allServers)

	return nil
}

func daemonIsServerRunning(name string) (bool, error) {
	resp, err := sendProtocol(protocol.Request{
		Command: "CHECK",
		Text:    name,
	})
	if err != nil {
		return false, err
	}

	if resp.OK && resp.Message == name {
		return resp.Data.(bool), nil
	} else {
		return false, fmt.Errorf("Incorrect response from daemon")
	}
}

func MakeList() ([]ServerInfo, error) {

	var result []ServerInfo

	err := filepath.WalkDir(paths.ServerRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if _, err := os.Stat(paths.Config(d.Name())); err == nil {

				isServerRunning, _ := daemonIsServerRunning(d.Name())

				cfg, err := config.Load(d.Name())
				if err == nil {
					server := ServerInfo{
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

func SetParameter(server, arg1, arg2 string) error {

	fmt.Print("\n")
	defer fmt.Print("\n")

	isServerRunning, err := daemonIsServerRunning(server)
	if err != nil {
		return err
	}
	if isServerRunning && arg1 != "autorestart" {
		return fmt.Errorf("Unable to use set server %s is already running", server)
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

		if err := setServerVersion(server, arg2); err != nil {
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

	default:
		return errors.New("Incorrect set paramater")
	}

	fmt.Printf("Successfully changed parameter %s for server %s to %q\n", arg1, server, arg2)

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
		return errors.New("unrecognized argument to autorestart")
	}
}

func setServerVersion(name, serverVersion string) error {
	cfg, err := config.Load(name)
	if err != nil {
		return err
	}

	oldServerVersion := cfg.Version

	if cfg.Version == serverVersion {
		return fmt.Errorf("Version for server %s is already %s", name, serverVersion)
	}

	if err := download.ArchiveJarFile(cfg); err != nil {
		fmt.Print("Unable to archive old jar file\n")
	}

	cfg.Version = serverVersion

	if err := download.DownloadJar(cfg); err != nil {
		fmt.Printf("Unable to find version %s, undoing changes...", serverVersion)
		cfg.Version = oldServerVersion
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

	return config.Save(cfg.Name, cfg)
}

func setWorldName(name, worldName string) error {

	serverPropertiesPath := paths.ServerProperties(name)

	if fileInfo, err := os.Stat(serverPropertiesPath); err != nil || fileInfo.IsDir() {
		return fmt.Errorf("unable to set world parameter: server.properties doesn't exist or is a directory")
	}

	return config.SetServerProperty(serverPropertiesPath, config.LevelNamePropertyKey, worldName)
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
