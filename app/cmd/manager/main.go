package main

import (
	"fmt"
	"minecraft-manager/internal/client"
	"minecraft-manager/internal/create"
	"minecraft-manager/internal/daemon"
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
		if len(os.Args) != 3 {
			return fmt.Errorf("usage: manager start <server>")
		}
		return client.StartServer(os.Args[2])

	case "create":
		if len(os.Args) != 5 {
			return fmt.Errorf("usage: manager create <server> <type> <version>")
		}
		return create.Create(os.Args[2], os.Args[3], os.Args[4])

	case "daemon":
		d := daemon.New()
		return d.Run()

	case "ping":
		return client.PingDaemon()

	case "ps":
		return client.GetPS()

	case "list":
		return client.GetList()

	case "screen":
		if len(os.Args) != 3 {
			return fmt.Errorf("usage: manager screen <server>")
		}
		return client.Screen(os.Args[2])

	case "stop":
		if len(os.Args) != 3 {
			return fmt.Errorf("usage: manager stop <server>")
		}
		return client.StopServer(os.Args[2])

	case "set":
		if len(os.Args) != 5 {
			return fmt.Errorf("usage: manager set <server> <parameter> <argument>")
		}
		return client.SetParameter(os.Args[2], os.Args[3], os.Args[4])

	case "import":
		return fmt.Errorf("not yet implemented") // TODO

	default:
		return fmt.Errorf("unknown command %q", os.Args[1])
	}
}
