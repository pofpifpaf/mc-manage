package launcher

import (
	"fmt"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/java"
	"minecraft-manager/internal/paths"
	"os"
	"os/exec"
)

func Build(server string) (*exec.Cmd, error) {
	cfg, err := config.Load(server)
	if err != nil {
		return nil, err
	}

	javaPath, err := java.Find(cfg.Java)
	if err != nil {
		return nil, err
	}

	serverDir := paths.Server(server)

	jarPath := paths.Jar(server, cfg.Jar)

	if _, err := os.Stat(jarPath); err != nil {
		return nil, fmt.Errorf("jar not found: %s", jarPath)
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

	fmt.Println("Starting", cfg.Name)
	fmt.Println("Java:", javaPath)
	fmt.Println("Directory:", serverDir)

	return cmd, nil
}