package templates

import (
	"bufio"
	"embed"
	"fmt"
	"minecraft-manager/internal/paths"
)

//go:embed templates/*
var Files embed.FS

func PrintConfigFile() error {
	file, err := Files.Open(paths.Templates(paths.ConfigJson))
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
