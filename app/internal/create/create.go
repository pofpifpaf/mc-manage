package create

import (
	"fmt"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/download"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/templates"
	"os"
)

func Create(name, serverType, version string) error {
	serverDir := paths.Server(name)

	fmt.Printf("Creating server %q\n", name)

	if _, err := os.Stat(serverDir); err == nil {
		return fmt.Errorf("server %q already exists", name)
	}

	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return err
	}

	fmt.Printf("Copying template to %q\n", serverDir)
	if err := templates.CopyTemplate(serverDir); err != nil {
		return err
	}

	cfg, err := config.Load(name)
	if err != nil {
		return err
	}

	cfg.Name = name
	cfg.Version = version
	cfg.Type = serverType

	fmt.Printf("Saving config file with default config\n")
	if err := config.Save(name, cfg); err != nil {
		return err
	}

	if err := download.DownloadJar(cfg); err != nil {
		return err
	}

	fmt.Printf("Created server %q\n", name)

	return nil
}

func ImportServer(name, serverType, version string) error {

	serverDir := paths.Server(name)

	if fileinfo, err := os.Stat(serverDir); err != nil || !fileinfo.IsDir() {
		return fmt.Errorf("unable to import: %s doesn't exist, or is not a folder", name)
	}

	if _, err := os.Stat(paths.Config(name)); err == nil {
		return fmt.Errorf("unable to import: %s already exists", paths.Config(name))
	}

	if err := templates.CreateConfigJsonFile(name, serverType); err != nil {
		return err
	}

	fmt.Printf("Created config file at %s\n", paths.Config(name))

	cfg, err := config.Load(name)
	if err != nil {
		return err
	}

	cfg.Type = serverType
	cfg.Version = version

	if err := config.Save(cfg.Name, cfg); err != nil {
		return err
	}

	if err := download.DownloadJar(cfg); err != nil {
		fmt.Printf("unable to download jar : %s", err)
	}

	fmt.Printf("Successfully imported server %s into manager\n\n", name)

	return nil
}
