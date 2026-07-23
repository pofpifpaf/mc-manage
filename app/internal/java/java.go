package java

import (
	"fmt"
	"os"
	"minecraft-manager/internal/paths"
)

func Find(version string) (string, error) {
	path := paths.Java(version)

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("java %s not installed: %w", version, err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("%s is not a file", path)
	}

	return path, nil
}