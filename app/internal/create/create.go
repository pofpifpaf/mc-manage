package create

import (
	"fmt"
	"io/fs"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/download"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/templates"
	"minecraft-manager/internal/ui"
	"minecraft-manager/internal/users"
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
	case 7:
		versionArg = os.Args[5] + "-" + os.Args[6]
	default:
		return fmt.Errorf("Invalid number of arguments: %d", len(os.Args))
	}

	valid, name := paths.ValidateServerName(name)
	if !valid {
		return fmt.Errorf("Invalid server name %s", name)
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

	if err := users.CreateUser(cfg); err != nil {
		ui.PrintWarning("Error while creating user : " + err.Error())
		config.SetConfigUserSpecificFalse(cfg)
	} else if err := users.SetServerPermissions(cfg); err != nil {
		ui.PrintWarning("Error while setting folder permissions: " + err.Error())
	}

	if err := config.Save(cfg.Name, cfg); err != nil {
		return err
	}

	if err := download.DownloadJar(cfg); err != nil {
		ui.PrintWarning("Unable to download jar, use \"manager set <server> version <version> --force\" to retry jar download")
		ui.PrintWarning("Error downloading jar: " + err.Error())
	}

	if err := config.Save(cfg.Name, cfg); err != nil {
		return err
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
	case 7:
		versionArg = os.Args[5] + "-" + os.Args[6]
	default:
		return fmt.Errorf("Invalid number of arguments: %d", len(os.Args))
	}

	valid, name := paths.ValidateServerName(name)
	if !valid {
		return fmt.Errorf("Invalid server name %s", name)
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

	if err := users.CreateUser(cfg); err != nil {
		ui.PrintWarning("Error while creating user : " + err.Error())
		config.SetConfigUserSpecificFalse(cfg)
	} else if err := users.SetServerPermissions(cfg); err != nil {
		ui.PrintWarning("Error while setting folder permissions: " + err.Error())
	}

	if cfg.Type == "forge" && version == "link" {
		cfg.Version, cfg.VersionArg, err = download.ForgeExtractVersionFromForge(versionArg)
		if err != nil {
			return err
		}
	}

	_ = config.LoadFromExisting(cfg)

	if err := config.Save(cfg.Name, cfg); err != nil {
		return err
	}

	if cfg.Type == "neoforge" || cfg.Type == "forge" {
		if err := config.ScriptGetJVMArgs(name); err != nil {
			ui.PrintError("could not retrieve jvm memory arguments")
		}
	}

	if err := download.DownloadJar(cfg); err != nil {
		ui.PrintError("unable to download jar :" + err.Error())
	}

	if err := config.Save(cfg.Name, cfg); err != nil {
		return err
	}

	if _, err := os.Stat(paths.Eula(name)); err != nil {
		ui.PrintWarning("No eula.txt file found")
	}

	ui.PrintSuccess("Imported server \"" + name + "\" into manager")

	return nil
}
