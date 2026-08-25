package daemon

import (
	"fmt"
	"minecraft-manager/internal/client"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/templates"
	"minecraft-manager/internal/ui"
	"minecraft-manager/internal/users"
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
		if server.Username == "disabled" || server.Uid == -1 || server.Gid == -1 {
			continue
		}
		if err := users.EnsureUserExistenceServerInfo(server); err != nil {
			ui.PrintWarning("Error while checking user existence for server " + server.Name + ": " + err.Error())
			if err := config.SetUserSpecificFalse(server.Name); err != nil {
				ui.PrintWarning("Error set user false: " + err.Error())
			}
		}
	}

	if err := d.handleMainConfig(); err != nil {
		return err
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

func (d *Daemon) handleMainConfig() error {

	ui.PrintInfo("Loading main config file")
	cfg, err := config.LoadMainConfig()
	if err != nil {

		ui.PrintWarning(fmt.Sprintf("Unable to load main config at %s, creating sample one", paths.GetServerRoot()))

		if err := templates.CreateMainConfigJsonFile(); err != nil {
			return err
		}

		cfg, err = config.LoadMainConfig()
		if err != nil {
			return err
		}

		cfg.ServerFilePath = paths.GetServerRoot()

		if err := config.SaveMainConfig(cfg); err != nil {
			return err
		}
	}

	d.manager.config = cfg

	return nil
}

func (d *Daemon) gracefulShutdown() error {

	if d.manager.ServerCount() == 0 {
		return nil
	}

	var gracePeriod time.Duration
	gracePeriod = time.Duration(d.manager.config.GracePeriodSeconds) * time.Second

	ui.PrintInfo(fmt.Sprintf("Found %d server(s) running, grace period is %s, attempting to stop...", d.manager.ServerCount(), gracePeriod))
	d.manager.StopAllServers()

	timeout := time.After(gracePeriod)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		if d.manager.ServerCount() == 0 {
			break
		}

		select {
		case <-timeout:
			ui.PrintError(fmt.Sprintf("Grace period timeout (%s), killing remaining servers", gracePeriod))
			d.manager.KillAllServers()
		case <-ticker.C:
		}
	}

	return nil
}
