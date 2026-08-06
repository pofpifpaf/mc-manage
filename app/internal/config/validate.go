package config

import (
	"fmt"
	"minecraft-manager/internal/protocol"
	"regexp"
)

func ValidateConfig(c protocol.Config) error {

	if c.Java == "" {
		return fmt.Errorf("missing java version")
	}

	if c.MemoryAllocated == "" {
		return fmt.Errorf("missing memory")
	}

	if c.MemoryMax == "" {
		return fmt.Errorf("missing memory")
	}

	if c.Jar == "" {
		return fmt.Errorf("missing jar")
	}

	return nil
}

func ValidateMemoryConfig(mem string) error {
	var memoryRegex = regexp.MustCompile(`^\d+([KkMmGg])?$`)
	if !memoryRegex.MatchString(mem) {
		return fmt.Errorf("invalid memory format %q", mem)
	}
	return nil
}
