package launcher

import (
	"fmt"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/java"
	"minecraft-manager/internal/paths"
	"os"
	"os/exec"
)

func Build(server string) (*exec.Cmd, bool, string, error) {
	cfg, err := config.Load(server)
	if err != nil {
		return nil, false, "", err
	}

	javaPath, err := java.Find(cfg.Java)
	if err != nil {
		return nil, false, "", err
	}

	serverDir := paths.Server(server)

	jarPath := paths.Jar(server, cfg.Jar)

	if _, err := os.Stat(jarPath); err != nil {
		return nil, false, "", fmt.Errorf("jar not found: %s", jarPath)
	}

	cmd := exec.Command(
		javaPath,
		"-Xms"+cfg.Memory,
		"-Xmx"+cfg.Memory,
		"-jar",
		cfg.Jar,
		"nogui",
	)

	cmd.Dir = serverDir

	fmt.Printf("Starting %s, with Java Path: %s and Server Directory : %s \n", cfg.Name, javaPath, serverDir)

	return cmd, cfg.AutomaticRestarts, cfg.Port, nil
}
