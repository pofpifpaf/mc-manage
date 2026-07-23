package create

import (
	"fmt"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/templates"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/download"
	"os"
)

func Create(name, serverType, version string) error {
	serverDir := paths.Server(name)

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
	cfg.Version = version
	cfg.Type = serverType

	fmt.Printf("Saving config file %q\n", name)
	if err := config.Save(name, cfg); err != nil {
		return err
	}

	switch serverType {
	case "vanilla":
		err = download.DownloadVanilla(cfg.Version, paths.Jar(name, cfg.Jar))
		if err != nil {
			return err
		}
	default:
		fmt.Printf("%q, Unsupported type\n", serverType)
	}

	fmt.Printf("Created server %q\n", name)

	return nil
}