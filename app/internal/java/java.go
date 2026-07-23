package java

import (
	"fmt"
	"os"
	"path/filepath"
)

func Find(version string) (string, error) {
	path := filepath.Join(
		"/opt/java",
		version,
		"bin",
		"java",
	)

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("java %s not installed: %w", version, err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("%s is not a file", path)
	}

	return path, nil
}