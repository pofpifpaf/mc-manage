package main

import (
	"fmt"
	"minecraft-manager/internal/launcher"
	"minecraft-manager/internal/create"
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
		return launcher.Start(os.Args[2])

	case "create":
		if len(os.Args) != 3 {
			return fmt.Errorf("usage: manager create <server>")
		}
		return create.Create(os.Args[2])

	default:
		return fmt.Errorf("unknown command %q", os.Args[1])
	}

	return nil
}