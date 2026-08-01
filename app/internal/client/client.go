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

	allServers, err := makeList()
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

func makeList() ([]ServerInfo, error) {

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

func SetParameter(server string, arg1 string, arg2 string) error {

	switch arg1 {
	case "port":

		port, err := strconv.Atoi(arg2)
		if err != nil || (port < 0 || port > 65535) {
			return errors.New("port number out of range, must be between 0 and 65535")
		}

		return send(protocol.Request{
			Command: "SET",
			Server:  server,
			Text:    arg1,
			Data:    arg2,
		})

	case "autorestart":

		if arg2 == "true" || arg2 == "false" {
			return send(protocol.Request{
				Command: "SET",
				Server:  server,
				Text:    arg1,
				Data:    arg2,
			})
		} else {
			return errors.New("unrecognized argument to autorestart")
		}

	case "version":

		isServerRunning, err := daemonIsServerRunning(server)
		if err != nil {
			return err
		}
		if isServerRunning {
			return fmt.Errorf("Unable to change version, server %s is already running", server)
		}

		cfg, err := config.Load(server)
		if err != nil {
			return err
		}

		if cfg.Version == arg2 {
			return fmt.Errorf("Version for server %s is already %s", server, arg2)
		}

		if err := download.ArchiveJarFile(cfg); err != nil {
			fmt.Print("Unable to archive old jar file\n")
		}

		cfg.Version = arg2

		if err := download.DownloadJar(cfg); err != nil {
			return err
		}

		if err := config.Save(cfg.Name, cfg); err != nil {
			return err
		}

		fmt.Printf("Successfully changed server %s to version %s", server, arg2)

		return nil

	case "java":

		isServerRunning, err := daemonIsServerRunning(server)
		if err != nil {
			return err
		}
		if isServerRunning {
			return fmt.Errorf("Unable to change java version, server %s is already running", server)
		}

		if _, err := java.Find(arg2); err != nil {
			return err
		}

		cfg, err := config.Load(server)
		if err != nil {
			return err
		}

		cfg.Java = arg2

		if err := config.Save(cfg.Name, cfg); err != nil {
			return err
		}

		fmt.Printf("Successfully changed server %s to java version %s\n", server, arg2)

		return nil

	default:
		return errors.New("Incorrect set paramater")
	}
}
