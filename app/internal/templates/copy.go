package templates

import (
	"fmt"
	"io/fs"
	"minecraft-manager/internal/paths"
	"os"
	"path/filepath"
)

func CopyTemplate(destination string) error {
	return fs.WalkDir(Files, "vanilla", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel("vanilla", path)
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
	})
}

func CreateConfigJsonFile(destination, serverType string) error {

	dst := paths.Config(destination)

	fmt.Printf("Creating config file for destination %s\n", dst)

	data, err := Files.ReadFile(filepath.Join(serverType, "config.json"))
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}
