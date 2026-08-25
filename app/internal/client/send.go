package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"minecraft-manager/internal/ui"
	"net"
)

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

	fmt.Print("\n")
	defer fmt.Print("\n")

	resp, err := sendProtocol(req)
	if err != nil {
		return err
	}

	if !resp.OK {
		return errors.New(resp.Message)
	}

	switch req.Command {

	case "PS":

		servers, err := makeServerInfoListInterface(resp.Data.([]interface{}))
		if err != nil {
			return err
		}

		ui.PrintRunningServers(servers)

	case "START", "STOP", "KILL":
		ui.PrintSuccess(resp.Message)

	case "SET":
		ui.PrintSuccess("SET - " + resp.Message)

	case "PING":
		if resp.Message == "PONG" {
			ui.PrintSuccess("Ping successful, daemon is running")
		} else {
			ui.PrintError("Ping failure, received " + resp.Message)
		}

	default:
		ui.PrintInfo("Response: " + resp.Message)
	}

	return nil
}

func makeServerInfoListInterface(data []interface{}) ([]protocol.ServerInfo, error) {

	servers := make([]protocol.ServerInfo, len(data))

	for i, v := range data {
		m := v.(map[string]interface{})

		b, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}

		var server protocol.ServerInfo
		if err := json.Unmarshal(b, &server); err != nil {
			return nil, err
		}

		if server.Running == protocol.StateRunning {
			server, err = getActivePlayerInformation(server)
			if err != nil {
				server.PlayersOnlineMax = -1
			}
		}

		servers[i] = server
	}

	return servers, nil
}

func makeServerInfoInterface(data interface{}) (protocol.ServerInfo, error) {

	b, err := json.Marshal(data)
	if err != nil {
		return protocol.ServerInfo{}, err
	}

	var info protocol.ServerInfo
	err = json.Unmarshal(b, &info)
	if err != nil {
		return protocol.ServerInfo{}, err
	}

	if info.Running != protocol.StateStopped && info.Running != protocol.StateStopSent {
		info, err = getActivePlayerInformation(info)
		if err != nil {
			info.PlayersOnlineMax = -1
		}
	}

	return info, nil
}
