package create

import (
	"fmt"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/templates"
	"os"
	"path/filepath"
)

func Create(name string) error {
	serverDir := filepath.Join("/server", name)

	fmt.Printf("Creating server %q\n", name)

	if _, err := os.Stat(serverDir); err == nil {
		return fmt.Errorf("server %q already exists", name)
	}

	fmt.Printf("Creating server directory %q\n", serverDir)
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return err
	}

	fmt.Printf("Copying template %q\n", serverDir)
	if err := templates.CopyTemplate(serverDir); err != nil {
		return err
	}

	fmt.Printf("Loading default config %q\n", name)
	cfg, err := config.Load(name)
	if err != nil {
		return err
	}

	cfg.Name = name

	fmt.Printf("Saving config file %q\n", name)
	if err := config.Save(name, cfg); err != nil {
		return err
	}

	fmt.Printf("Created server %q\n", name)

	return nil
}