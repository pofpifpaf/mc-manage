package create

import (
	"fmt"
	"io/fs"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/download"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/templates"
	"minecraft-manager/internal/ui"
	"os"
)

func Create(args []string) error {

	var name, serverType, version, versionArg string

	name = os.Args[2]
	serverType = os.Args[3]
	version = os.Args[4]

	switch len(os.Args) {
	case 5:
		versionArg = ""
	case 6:
		versionArg = os.Args[5]
	default:
		return fmt.Errorf("Invalid number of arguments: %d", len(os.Args))
	}

	fmt.Print("\n")
	defer fmt.Print("\n")

	serverDir := paths.Server(name)

	ui.PrintInfo("Creating server \"" + name + "\"")

	if _, err := os.Stat(serverDir); err == nil {
		return fmt.Errorf("server %q already exists", name)
	}

	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return err
	}

	if err := templates.CopyTemplate(serverDir, serverType); err != nil {
		return err
	}

	if err := templates.CreateConfigJsonFile(name); err != nil {
		return err
	}

	cfg, err := config.Load(name)
	if err != nil {
		return err
	}

	cfg.Name = name
	cfg.Version = version
	cfg.Type = serverType
	cfg.VersionArg = versionArg

	ui.PrintInfo("Saving config file with default config")
	if err := config.Save(name, cfg); err != nil {
		return err
	}

	if err := download.DownloadJar(cfg); err != nil {
		ui.PrintWarning("Unable to download jar, use \"manager set <server> version\"  to retry jar download")
		ui.PrintWarning("Error downloading jar: " + err.Error())
	}

	ui.PrintSuccess("Created server \"" + name + "\"")

	return nil
}

func ImportServer(args []string) error {

	var name, serverType, version, versionArg string

	name = os.Args[2]
	serverType = os.Args[3]
	version = os.Args[4]

	switch len(os.Args) {
	case 5:
		versionArg = ""
	case 6:
		versionArg = os.Args[5]
	default:
		return fmt.Errorf("Invalid number of arguments: %d", len(os.Args))
	}

	fmt.Print("\n")
	defer fmt.Print("\n")

	serverDir := paths.Server(name)

	if fileinfo, err := os.Stat(serverDir); err != nil || !fileinfo.IsDir() {
		return fmt.Errorf("unable to import: %s doesn't exist, or is not a folder", name)
	}

	if _, err := os.Stat(paths.Config(name)); err == nil {
		return fmt.Errorf("unable to import: %s already exists", paths.Config(name))
	}

	if err := templates.CreateConfigJsonFile(name); err != nil {
		if err == fs.ErrNotExist {
			return fmt.Errorf("unable to import: %s template not supported", serverType)
		}
		return err
	}

	ui.PrintInfo("Created config file at " + paths.Config(name))

	cfg, err := config.Load(name)
	if err != nil {
		return err
	}

	cfg.Name = name
	cfg.Type = serverType
	cfg.Version = version
	cfg.VersionArg = versionArg

	serverPropertiesPath := paths.ServerProperties(name)
	if _, err := os.Stat(serverPropertiesPath); err == nil {

		ui.PrintInfo("Found server.properties file at " + serverPropertiesPath)

		port, err := config.GetServerProperty(serverPropertiesPath, "server-port")
		if err == nil {
			ui.PrintInfo("Found server-port key, port = " + port)
			cfg.Port = port
		}

		worldName, err := config.GetServerProperty(serverPropertiesPath, "level-name")
		if err == nil {
			ui.PrintInfo("Found level-name key, world = " + worldName)
			cfg.LevelName = worldName
		}
	}

	if err := config.Save(cfg.Name, cfg); err != nil {
		return err
	}

	if err := download.DownloadJar(cfg); err != nil {
		ui.PrintError("unable to download jar :" + err.Error())
	}

	if _, err := os.Stat(paths.Eula(name)); err != nil {
		ui.PrintWarning("No eula.txt file found")
	}

	ui.PrintSuccess("Imported server \"" + name + "\" into manager")

	return nil
}
