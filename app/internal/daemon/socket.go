package daemon

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"minecraft-manager/internal/paths"
)

func Listen(manager *Manager) error {
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
			return err
		}

		go handleConnection(conn, manager)
	}
}

func handleConnection(conn net.Conn, manager *Manager) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	message, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	parts := strings.Fields(message)

	if len(parts) == 0 {
		return
	}

	switch parts[0] {

	case "PING":
		fmt.Fprintln(conn, "PONG")

	case "LIST":
		for _, server := range manager.List() {
			fmt.Fprintln(conn, server.Name)
		}
	
	case "START":

		if len(parts) != 2 {
			fmt.Fprintln(conn, "usage: START <server>")
			return
		}

		err := manager.Start(parts[1])

		if err != nil {
			fmt.Fprintln(conn, err)
			return
		}

		fmt.Fprintln(conn, "OK")

	default:
		fmt.Fprintln(conn, "UNKNOWN COMMAND")
	}
}