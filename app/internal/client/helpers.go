package client

import (
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"os"
)

func MakeList() ([]protocol.ServerInfo, error) {

	entries, err := os.ReadDir(paths.GetServerRoot())
	if err != nil {
		return nil, err
	}

	var result []protocol.ServerInfo

	for _, d := range entries {
		if !d.IsDir() {
			continue
		}

		name := d.Name()

		if _, err := os.Stat(paths.Config(name)); err != nil {
			continue
		}

		isServerRunning, _ := daemonIsServerRunning(name)

		cfg, err := config.Load(name)
		if err != nil {
			continue
		}

		result = append(result, protocol.ServerInfo{
			Name:              name,
			Type:              cfg.Type,
			Port:              cfg.Port,
			AutomaticRestarts: cfg.AutomaticRestarts,
			Running:           isServerRunning,
			Version:           cfg.Version,
			JavaVersion:       cfg.Java,
			StartOnBoot:       cfg.StartOnBoot,
			Username:          cfg.Username,
			Uid:               cfg.Uid,
			Gid:               cfg.Gid,
		})
	}

	return result, nil
}

func getActivePlayerInformation(server protocol.ServerInfo) (protocol.ServerInfo, error) {

	status, err := protocol.GetServerStatus(server.Port)
	if err != nil {
		server.Running = protocol.StateStarting
		return server, err
	}

	server.PlayersOnline = status.Players.Online
	server.PlayersOnlineMax = status.Players.Max

	return server, nil
}
