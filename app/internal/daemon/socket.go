package daemon

import (
	"encoding/json"
	"fmt"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
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

	fmt.Println("Listening on", paths.SocketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept:", err)
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

	fmt.Println("Listening on", paths.ScreenSocketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept:", err)
			continue
		}

		fmt.Println("Listener accepted")

		go d.handleScreenConn(conn)
	}
}

func (d *Daemon) handleScreenConn(conn net.Conn) {

	fmt.Println("Received screen request")

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

	// reader := bufio.NewReader(conn)

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

	case "LIST":
		servers := d.manager.List()

		json.NewEncoder(conn).Encode(
			protocol.Response{
				OK:      true,
				Message: "list",
				Data:    servers,
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
				Message: "Starting server",
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
