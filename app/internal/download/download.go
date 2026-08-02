package download

import (
	"fmt"
	"io"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/paths"
	"net/http"
	"os"
	"path/filepath"
)

func RetrieveJarIfArchived(cfg *config.Config) (bool, error) {

	jarArchivePath := filepath.Join(paths.Server(cfg.Name), cfg.Jar+"."+cfg.Version+".old")

	if _, err := os.Stat(jarArchivePath); err == nil {
		if err := os.Rename(jarArchivePath, paths.Jar(cfg.Name, cfg.Jar)); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func DownloadJar(cfg *config.Config) error {
	var err error

	if isArchived, _ := RetrieveJarIfArchived(cfg); isArchived {
		fmt.Printf("Retrieved config file from archive for server %s and version %s\n", cfg.Name, cfg.Version)
		return nil
	}

	switch cfg.Type {
	case "vanilla":
		err = DownloadVanilla(cfg.Version, paths.Jar(cfg.Name, cfg.Jar))
		if err != nil {
			return err
		}
	default:
		fmt.Printf("%q, Unsupported type\n", cfg.Type)
	}

	return err
}

func DownloadCustomJar(cfg *config.Config, downloadURL string) error {

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Downloading from url %q\n", downloadURL)

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

func ArchiveJarFile(cfg *config.Config) error {

	oldPath := paths.Jar(cfg.Name, cfg.Jar)
	newPath := filepath.Join(paths.Server(cfg.Name), cfg.Jar+"."+cfg.Version+".old")

	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	return nil
}
