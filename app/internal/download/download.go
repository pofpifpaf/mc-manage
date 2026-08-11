package download

import (
	"fmt"
	"io"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"minecraft-manager/internal/ui"
	"net/http"
	"os"
	"path/filepath"
)

func RetrieveJarIfArchived(cfg *protocol.Config) (bool, error) {

	var jarArchivePath string
	if cfg.VersionArg == "" {
		jarArchivePath = filepath.Join(paths.Server(cfg.Name), cfg.Jar+"."+cfg.Version+".old")
	} else {
		jarArchivePath = filepath.Join(paths.Server(cfg.Name), cfg.Jar+"."+cfg.Version+"-"+cfg.VersionArg+".old")
	}

	if _, err := os.Stat(jarArchivePath); err == nil {
		if err := os.Rename(jarArchivePath, paths.Jar(cfg.Name, cfg.Jar)); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func DownloadJar(cfg *protocol.Config) error {
	var err error

	if isArchived, _ := RetrieveJarIfArchived(cfg); isArchived {
		ui.PrintInfo(fmt.Sprintf("Retrieved config file from archive for server %s and version %s", cfg.Name, cfg.Version))
		return nil
	}

	switch cfg.Type {
	case "vanilla":
		if err := DownloadVanilla(cfg.Version, paths.Jar(cfg.Name, cfg.Jar)); err != nil {
			return err
		}
	case "paper":
		if err := DownloadPaper(cfg.Version, cfg.VersionArg, paths.Jar(cfg.Name, cfg.Jar)); err != nil {
			return err
		}
	case "neoforge":
		if err := InstallNeoforge(cfg); err != nil {
			return err
		}
	default:
		ui.PrintError("\"" + cfg.Type + "\" Unsupported type")
	}

	return err
}

func DownloadCustomJar(cfg *protocol.Config, downloadURL string) error {

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ui.PrintInfo("Downloading from url \"" + downloadURL + "\"")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	destination := paths.Jar(cfg.Name, cfg.Jar)

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	return nil
}

func ArchiveJarFile(cfg *protocol.Config) error {

	oldPath := paths.Jar(cfg.Name, cfg.Jar)
	var newPath string
	if cfg.VersionArg == "" {
		newPath = filepath.Join(paths.Server(cfg.Name), cfg.Jar+"."+cfg.Version+".old")
	} else {
		newPath = filepath.Join(paths.Server(cfg.Name), cfg.Jar+"."+cfg.Version+"-"+cfg.VersionArg+".old")
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	return nil
}
