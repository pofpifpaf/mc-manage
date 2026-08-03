package daemon

import (
	"fmt"
	"minecraft-manager/internal/client"
	"minecraft-manager/internal/ui"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	fmt.Println("----------------------------------------------------------------")
	ui.PrintInfo("Starting Daemon")

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

	ui.PrintSuccess(fmt.Sprintf("Minecraft manager daemon started, PID: %d", os.Getpid()))

	// Wait for shutdown signals
	signals := make(chan os.Signal, 1)

	signal.Notify(
		signals,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-signals

	ui.PrintInfo("Daemon shutting down")
	if err := d.gracefulShutdown(); err != nil {
		return fmt.Errorf("Error in graceful shutdown : %s", err)
	}
	ui.PrintSuccess("Daemon shut down with no errors")
	fmt.Println("----------------------------------------------------------------")

	return nil
}

func (d *Daemon) gracefulShutdown() error {

	if d.manager.ServerCount() == 0 {
		return nil
	}

	ui.PrintInfo(fmt.Sprintf("Found %d server(s) running, attempting to stop...", d.manager.ServerCount()))
	d.manager.StopAllServers()

	timeout := time.After(d.manager.gracePeriodSeconds)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		if d.manager.ServerCount() == 0 {
			break
		}

		select {
		case <-timeout:
			ui.PrintError(fmt.Sprintf("Timeout of %s elapsed, killing remaining servers", d.manager.gracePeriodSeconds))
			d.manager.KillAllServers()
		case <-ticker.C:
		}
	}

	return nil
}
