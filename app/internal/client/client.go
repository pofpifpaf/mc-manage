package client

import (
	"errors"
	"fmt"
	"io"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/download"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"minecraft-manager/internal/ui"
	"minecraft-manager/internal/users"
	"net/http"
	"os"
	"os/user"
	"strconv"
)

func StartServer(server string) error {
	valid, server := paths.ValidateServerName(server)
	if !valid {
		return fmt.Errorf("Invalid server name %s", server)
	}
	return send(
		protocol.Request{
			Command: "START",
			Server:  server,
		},
	)
}

func StopServer(server string) error {
	valid, server := paths.ValidateServerName(server)
	if !valid {
		return fmt.Errorf("Invalid server name %s", server)
	}
	return send(
		protocol.Request{
			Command: "STOP",
			Server:  server,
		},
	)
}

func PingDaemon() error {
	return send(
		protocol.Request{
			Command: "PING",
		},
	)
}

func GetPS() error {
	return send(
		protocol.Request{
			Command: "PS",
		},
	)
}

func GetList() error {

	allServers, err := MakeList()
	if err != nil {
		return err
	}

	ui.PrintServerList(allServers)

	return nil
}

func daemonIsServerRunning(name string) (protocol.ServerState, error) {
	resp, err := sendProtocol(protocol.Request{
		Command: "CHECK",
		Server:  name,
	})
	if err != nil {
		return protocol.StateStopped, err
	}

	if resp.OK && resp.Message == name {
		return protocol.ServerState(resp.Data.(string)), nil
	} else {
		return protocol.StateStopped, fmt.Errorf("Incorrect response from daemon")
	}
}

func DownloadJarToServer(server, downloadURL string) error {

	valid, server := paths.ValidateServerName(server)
	if !valid {
		return fmt.Errorf("Invalid server name %s", server)
	}

	fmt.Print("\n")
	defer fmt.Print("\n")

	cfg, err := config.Load(server)
	if err != nil {
		return err
	}

	if err := download.DownloadCustomJar(cfg, downloadURL); err != nil {
		return err
	}

	cfg.Version = "custom"

	return config.Save(server, cfg)
}

func InspectServer(name string) error {

	valid, name := paths.ValidateServerName(name)
	if !valid {
		return fmt.Errorf("Invalid server name %s", name)
	}

	resp, err := sendProtocol(protocol.Request{
		Command: "INSPECT",
		Server:  name,
	})
	if err != nil {
		return err
	}

	serverInfo := protocol.ServerInfo{}

	if !resp.OK {
		serverInfo.Running = protocol.StateStopped
	} else {
		serverInfo, err = makeServerInfoInterface(resp.Data)
	}

	cfg, err := config.Load(name)
	if err != nil {
		return err
	}

	ui.PrintInspectServer(serverInfo, cfg)

	return nil
}

func KillServer(name string) error {

	valid, name := paths.ValidateServerName(name)
	if !valid {
		return fmt.Errorf("Invalid server name %s", name)
	}

	return send(protocol.Request{
		Command: "KILL",
		Server:  name,
	})
}

func ReloadProperties(server string) error {

	fmt.Print("\n")
	defer fmt.Print("\n")

	valid, server := paths.ValidateServerName(server)
	if !valid {
		return fmt.Errorf("Invalid server name %s", server)
	}

	cfg, err := config.Load(server)
	if err != nil {
		return err
	}

	if err := config.LoadFromExisting(cfg); err != nil {
		return err
	}

	if err := config.Save(server, cfg); err != nil {
		return err
	}

	ui.PrintSuccess("Reload complete")

	return nil
}

func ReloadUser(server string) error {

	fmt.Print("\n")
	defer fmt.Print("\n")

	valid, server := paths.ValidateServerName(server)
	if !valid {
		return fmt.Errorf("Invalid server name %s", server)
	}

	cfg, err := config.Load(server)
	if err != nil {
		return err
	}

	if cfg.Username == "disabled" || cfg.Uid == -1 || cfg.Gid == -1 {
		return fmt.Errorf("Per Server User disabled on this server")
	}

	if err := users.RemoveUser(cfg); err != nil {
		ui.PrintWarning("Couldn't remove user: " + err.Error())
	}

	if err := users.CreateUser(cfg); err != nil {
		ui.PrintWarning("Error while creating user : " + err.Error())
		config.SetConfigUserSpecificFalse(cfg)
	} else if err := users.SetServerPermissions(cfg); err != nil {
		ui.PrintWarning("Error while setting folder permissions: " + err.Error())
	}

	return config.Save(cfg.Name, cfg)
}

func ResumeFromRootConfig(server string) error {
	cfg, err := config.Load(server)
	if err != nil {
		return err
	}

	cfg.Username = users.UsernameFromServer(server)

	u, err := user.Lookup(cfg.Username)
	_, ok := errors.AsType[user.UnknownUserError](err)
	if ok {
		if err := users.CreateUser(cfg); err != nil {
			ui.PrintWarning("Error while creating user : " + err.Error())
			config.SetConfigUserSpecificFalse(cfg)
		} else if err := users.SetServerPermissions(cfg); err != nil {
			ui.PrintWarning("Error while setting folder permissions: " + err.Error())
		}
	} else if err == nil {
		uid, err := strconv.ParseUint(u.Uid, 10, 32)
		if err != nil {
			return err
		}

		gid, err := strconv.ParseUint(u.Gid, 10, 32)
		if err != nil {
			return err
		}

		cfg.Uid = int(uid)
		cfg.Gid = int(gid)
	} else {
		return err
	}

	return config.Save(cfg.Name, cfg)
}

func DownloadModToServer(server, url, modName string) error {

	fmt.Print("\n")
	defer fmt.Print("\n")

	valid, server := paths.ValidateServerName(server)
	if !valid {
		return fmt.Errorf("Invalid server name %s", server)
	}

	fi, err := os.Stat(paths.ModsFolder(server))
	if err != nil {
		ui.PrintInfo("Creating mods folder")
		if err := os.MkdirAll(paths.ModsFolder(server), 0644); err != nil {
			return err
		}
	} else if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", paths.ModsFolder(server))
	}

	ui.PrintInfo("Downloading mod")
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	destination := paths.Mod(server, modName)

	ui.PrintInfo("Copying to " + destination)
	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
