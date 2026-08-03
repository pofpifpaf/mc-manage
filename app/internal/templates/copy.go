package templates

import (
	"fmt"
	"io/fs"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/ui"
	"os"
	"path/filepath"
)

func CopyTemplate(destination, serverType string) error {
	if err := fs.WalkDir(Files, paths.Templates(serverType), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(paths.Templates(serverType), path)
		if err != nil {
			return err
		}

		if rel == "." {
			return nil
		}

		dst := filepath.Join(destination, rel)

		if d.IsDir() {
			return os.MkdirAll(dst, 0755)
		}

		data, err := Files.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dst, data, 0644)
	}); err != nil {
		return fmt.Errorf("Template %s not supported", serverType)
	}

	return nil
}

func CreateConfigJsonFile(destination string) error {

	dst := paths.Config(destination)

	ui.PrintInfo("Creating config file for destination " + dst)

	data, err := Files.ReadFile(paths.Templates(paths.ConfigJson))
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}
