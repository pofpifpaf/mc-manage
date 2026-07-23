package launcher

import (
	"fmt"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/java"
	"os"
	"os/exec"
	"path/filepath"
)

func Start(server string) error {
	cfg, err := config.Load(server)
	if err != nil {
		return err
	}

	javaPath, err := java.Find(cfg.Java)
	if err != nil {
		return err
	}

	serverDir := filepath.Join("server", server)

	jarPath := filepath.Join(serverDir, cfg.Jar)

	if _, err := os.Stat(jarPath); err != nil {
		return fmt.Errorf("jar not found: %s", jarPath)
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

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Println("Starting", cfg.Name)
	fmt.Println("Java:", javaPath)
	fmt.Println("Directory:", serverDir)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("minecraft exited with error: %w", err)
	}

	return nil
}