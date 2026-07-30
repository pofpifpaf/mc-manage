package daemon

import (
	"fmt"
	"minecraft-manager/internal/client"
	"os"
	"os/signal"
	"syscall"
)

type Daemon struct {
	manager *Manager
}

func New() *Daemon {
	return &Daemon{
		manager: NewManager(),
	}
}

func (d *Daemon) Run() error {

	go func() {
		if err := d.Listen(); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := d.listenScreen(); err != nil {
			fmt.Println("screen listener error:", err)
			os.Exit(1)
		}
	}()

	if client.PingDaemon() == nil {
		return fmt.Errorf("daemon already running")
	}

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
