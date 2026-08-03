package daemon

import (
	"encoding/json"
	"fmt"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"minecraft-manager/internal/ui"
	"net"
	"os"
)

func (d *Daemon) Listen() error {
	// Remove old socket if it exists
	os.Remove(paths.SocketPath)

	listener, err := net.Listen("unix", paths.SocketPath)
	if err != nil {
		return err
	}

	defer listener.Close()

	ui.PrintSuccess("CLIENT - Listening on" + paths.SocketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			ui.PrintError("accept: " + err.Error())
			continue
		}

		go d.handleConnection(conn)
	}
}

func (d *Daemon) listenScreen() error {
	// Remove old socket if it exists
	os.Remove(paths.ScreenSocketPath)

	listener, err := net.Listen("unix", paths.ScreenSocketPath)
	if err != nil {
		return err
	}

	defer listener.Close()

	ui.PrintSuccess("SCREEN - Listening on" + paths.ScreenSocketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			ui.PrintError("accept: " + err.Error())
			continue
		}

		go d.handleScreenConn(conn)
	}
}

func (d *Daemon) handleScreenConn(conn net.Conn) {

	ui.PrintInfo("Received screen request")

	var req protocol.Request
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		_ = json.NewEncoder(conn).Encode(protocol.Response{
			OK:      false,
			Message: "invalid screen request",
		})
		return
	}

	if req.Command != "SCREEN" || req.Server == "" {
		_ = json.NewEncoder(conn).Encode(protocol.Response{
			OK:      false,
			Message: "usage: SCREEN <server>",
		})
		return
	}

	server, ok := d.manager.Get(req.Server)
	if !ok {
		_ = json.NewEncoder(conn).Encode(protocol.Response{
			OK:      false,
			Message: fmt.Sprintf("server %q not running", req.Server),
		})
		return
	}

	_ = json.NewEncoder(conn).Encode(protocol.Response{
		OK:      true,
		Message: "hello!",
	})

	screenClient := NewScreenClient(conn)
	go server.Attach(screenClient)
}

func (d *Daemon) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)

	var req protocol.Request

	if err := decoder.Decode(&req); err != nil {
		return
	}

	switch req.Command {

	case "PING":
		json.NewEncoder(conn).Encode(
			protocol.Response{
				OK:      true,
				Message: "PONG",
			},
		)

	case "PS":
		servers := d.manager.ListRunning()

		json.NewEncoder(conn).Encode(
			protocol.Response{
				OK:      true,
				Message: "ps",
				Data:    servers,
			},
		)

	case "CHECK":

		_, isServerRunning := d.manager.Get(req.Server)

		json.NewEncoder(conn).Encode(
			protocol.Response{
				OK:      true,
				Message: req.Server,
				Data:    isServerRunning,
			},
		)

	case "INSPECT":

		server, exist := d.manager.Get(req.Server)

		if !exist {
			json.NewEncoder(conn).Encode(
				protocol.Response{
					OK:      false,
					Message: "Server not found, or is not running",
				},
			)
			return
		}

		serverInfo, err := MakeServerInfo(server)
		if err != nil {
			json.NewEncoder(conn).Encode(
				protocol.Response{
					OK:      false,
					Message: "error: " + err.Error(),
				},
			)
			return
		}

		json.NewEncoder(conn).Encode(
			protocol.Response{
				OK:   true,
				Data: serverInfo,
			},
		)

	case "START":

		if req.Server == "" {
			json.NewEncoder(conn).Encode(
				protocol.Response{
					OK:      false,
					Message: "usage START <server>",
				},
			)
			return
		}

		err := d.manager.Start(req.Server)

		if err != nil {
			json.NewEncoder(conn).Encode(
				protocol.Response{
					OK:      false,
					Message: err.Error(),
				},
			)
			return
		}

		json.NewEncoder(conn).Encode(
			protocol.Response{
				OK:      true,
				Message: fmt.Sprintf("Starting server %s", req.Server),
			},
		)

	case "STOP":

		if req.Server == "" {
			json.NewEncoder(conn).Encode(
				protocol.Response{
					OK:      false,
					Message: "usage STOP <server>",
				},
			)
			return
		}

		err := d.manager.Stop(req.Server)

		if err != nil {
			json.NewEncoder(conn).Encode(
				protocol.Response{
					OK:      false,
					Message: err.Error(),
				},
			)
			return
		}

		json.NewEncoder(conn).Encode(
			protocol.Response{
				OK:      true,
				Message: fmt.Sprintf("Stop server command sent to server %s", req.Server),
			},
		)

	case "KILL":

		if req.Server == "" {
			json.NewEncoder(conn).Encode(
				protocol.Response{
					OK:      false,
					Message: "usage KILL <server>",
				},
			)
			return
		}

		err := d.manager.Kill(req.Server)

		if err != nil {
			json.NewEncoder(conn).Encode(
				protocol.Response{
					OK:      false,
					Message: err.Error(),
				},
			)
			return
		}

		json.NewEncoder(conn).Encode(
			protocol.Response{
				OK:      true,
				Message: fmt.Sprintf("Kill server command sent to server %s", req.Server),
			},
		)

	case "SET":

		err := d.manager.SetParameter(req.Server, req.Text, req.Data)

		if err != nil {
			json.NewEncoder(conn).Encode(
				protocol.Response{
					OK:      false,
					Message: err.Error(),
				},
			)
			return
		}

		json.NewEncoder(conn).Encode(
			protocol.Response{
				OK:      true,
				Message: fmt.Sprintf("Set Parameter %s for server %s succesful", req.Text, req.Server),
			},
		)

	default:
		json.NewEncoder(conn).Encode(
			protocol.Response{
				OK:      false,
				Message: "UNKNOWN COMMAND",
			},
		)
	}
}
