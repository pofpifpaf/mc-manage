package main

import (
	"fmt"
	"minecraft-manager/internal/launcher"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args

	if len(args) != 3 {
		return fmt.Errorf("usage: manager start <server>")
	}

	command := args[1]
	server := args[2]

	switch command {
	case "start":
		return launcher.Start(server)

	default:
		return fmt.Errorf("unknown command %q", command)
	}

	return nil
}