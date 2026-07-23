package daemon

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func Run() error {

	manager := NewManager()

	go func() {
		if err := Listen(manager); err != nil {
			fmt.Println("socket error:", err)
			os.Exit(1)
		}
	}()

	fmt.Println("Minecraft manager daemon started")
	fmt.Println("PID:", os.Getpid())

	// Wait for shutdown signals
	signals := make(chan os.Signal, 1)

	signal.Notify(
		signals,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-signals

	fmt.Println("Daemon shutting down")

	return nil
}

