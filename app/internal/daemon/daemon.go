package daemon

import (
	"fmt"
	"minecraft-manager/internal/client"
	"minecraft-manager/internal/ui"
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
			ui.PrintError("screen listener error: " + err.Error())
			os.Exit(1)
		}
	}()

	if client.PingDaemon() == nil {
		return fmt.Errorf("daemon already running")
	}

	ui.PrintInfo("Starting servers that need to be started on boot")
	servers, err := client.MakeList()
	if err != nil {
		ui.PrintError("Unable to check for start on boot servers, err = " + err.Error())
	}
	for _, server := range servers {
		if server.StartOnBoot {
			d.manager.Start(server.Name)
		}
	}

	ui.PrintInfo("Minecraft manager daemon started")
	ui.PrintInfo(fmt.Sprintf("PID: %d", os.Getpid()))

	// Wait for shutdown signals
	signals := make(chan os.Signal, 1)

	signal.Notify(
		signals,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-signals

	fmt.Println("Daemon shutting down")
	fmt.Println("----------------------------")

	return nil
}
