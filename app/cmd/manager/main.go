package main

import (
	"fmt"
	"minecraft-manager/internal/create"
	"minecraft-manager/internal/daemon"
	"minecraft-manager/internal/client"
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
		return client.Start(os.Args[2])

	case "create":
		if len(os.Args) != 5 {
			return fmt.Errorf("usage: manager create <server> <type> <version>")
		}
		return create.Create(os.Args[2], os.Args[3], os.Args[4])

	case "daemon":
		return daemon.Run()

	case "ping":
		return client.Send("PING")

	case "list":
		return client.Send("LIST")

	default:
		return fmt.Errorf("unknown command %q", os.Args[1])
	}

	return nil
}