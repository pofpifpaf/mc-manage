package templates

import (
	"io/fs"
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