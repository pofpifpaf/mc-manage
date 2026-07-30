package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func SetProperty(filename, key, value string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	found := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			lines = append(lines, line)
			continue
		}

		if strings.HasPrefix(line, key+"=") {
			lines = append(lines, fmt.Sprintf("%s=%s", key, value))
			found = true
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if !found {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	output := strings.Join(lines, "\n")
	return os.WriteFile(filename, []byte(output), 0644)
}

func GetParameter(filename, key string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	found := false
	result := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		if strings.HasPrefix(line, key+"=") {
			result, found = strings.CutPrefix(line, key+"=")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if !found {
		return "", fmt.Errorf("key %s not found in server.properties", key)
	}

	return result, nil
}
