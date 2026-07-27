package main

import (
	"fmt"
	"minecraft-manager/internal/client"
	"minecraft-manager/internal/create"
	"minecraft-manager/internal/daemon"
	"minecraft-manager/internal/protocol"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func run() error {

	switch os.Args[1] {

	case "start":
		return client.Send(
			protocol.Request{
				Command: "START",
				Server:  os.Args[2],
			},
		)

	case "create":
		if len(os.Args) != 5 {
			return fmt.Errorf("usage: manager create <server> <type> <version>")
		}
		return create.Create(os.Args[2], os.Args[3], os.Args[4])

	case "daemon":
		d := daemon.New()
		return d.Run()

	case "ping":
		return client.Send(
			protocol.Request{
				Command: "PING",
			},
		)

	case "list":
		return client.Send(
			protocol.Request{
				Command: "LIST",
			},
		)

	case "screen":
		if len(os.Args) != 3 {
			return fmt.Errorf("usage: manager screen <server>")
		}
		return client.Screen(os.Args[2])

	default:
		return fmt.Errorf("unknown command %q", os.Args[1])
	}
}
