package paths

import (
	"fmt"
	"io/fs"
	"minecraft-manager/internal/protocol"
	"path/filepath"
	"strings"
)

func DirSize(path string) (string, error) {
	var size int64

	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			size += info.Size()
		}

		return nil
	})

	return HumanBytes(size), err
}

func HumanBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB",
		float64(size)/float64(div),
		"KMGTPE"[exp],
	)
}

func ValidateServerName(server string) (bool, string) {
	server, _ = strings.CutSuffix(server, "/")

	return !strings.ContainsAny(server, protocol.InvalidNameCharacters), server
}
