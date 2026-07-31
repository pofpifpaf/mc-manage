package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"net"
	"os"
	"strconv"
	"text/tabwriter"
	"time"
)

type ServerInfo struct {
	Name              string
	Port              string
	AutomaticRestarts bool
	CreatedAt         time.Time
}

type ServerListResponse struct {
	OK      bool         `json:"ok"`
	Message string       `json:"message,omitempty"`
	Data    []ServerInfo `json:"data"`
}

func send(req protocol.Request) error {

	conn, err := net.Dial("unix", paths.SocketPath)
	if err != nil {
		return err
	}

	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return err
	}

	var resp protocol.Response

	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return err
	}

	if !resp.OK {
		return errors.New(resp.Message)
	}

	switch req.Command {

	case "LIST":

		raw := resp.Data.([]interface{})

		servers := make([]ServerInfo, len(raw))

		for i, v := range raw {
			m := v.(map[string]interface{})

			b, err := json.Marshal(m)
			if err != nil {
				return err
			}

			var server ServerInfo
			if err := json.Unmarshal(b, &server); err != nil {
				return err
			}

			servers[i] = server
		}

		printList(servers)

	case "START", "STOP":
		fmt.Printf("Daemon responded without error - %s\n", resp.Message)

	case "SET":
		fmt.Printf("SET - %s\n", resp.Message)

	default:
		fmt.Printf("Response: %q\n", resp.Message)
	}

	return nil
}

func printList(servers []ServerInfo) {

	fmt.Printf("\n")

	if len(servers) == 0 {
		fmt.Println("No servers running")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tPORT\tAUTO RESTART\tSIZE\tUPTIME")
	fmt.Fprintln(w, "----\t----\t------------\t----\t------")

	for _, server := range servers {
		dirSize, _ := paths.DirSize(paths.Server(server.Name))
		uptime := time.Since(server.CreatedAt).Round(time.Second)

		fmt.Fprintf(
			w,
			"%s\t%s\t%t\t%s\t%s\n",
			server.Name,
			server.Port,
			server.AutomaticRestarts,
			dirSize,
			uptime,
		)
	}
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

func GetList() error {
	return send(
		protocol.Request{
			Command: "LIST",
		},
	)
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

	default:
		return errors.New("Incorrect set paramater")
	}
}
