package config

import (
	"bufio"
	"fmt"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/ui"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ConfigureJavaRunScript(server string) error {

	cfg, err := Load(server)
	if err != nil {
		return err
	}

	runPath := paths.Jar(server, cfg.Jar)
	javaPath := paths.Java(cfg.Java)

	data, err := os.ReadFile(runPath)
	if err != nil {
		return err
	}

	content := string(data)

	content = strings.Replace(content, "java @user_jvm_args.txt", javaPath+" @user_jvm_args.txt", 1)

	err = os.WriteFile(runPath, []byte(content), 0644)
	if err != nil {
		return err
	}

	return nil
}

func ScriptSetJVMMemoryArgs(server string) error {

	cfg, err := Load(server)
	if err != nil {
		return err
	}

	userArgsPath := filepath.Join(paths.Server(server), "user_jvm_args.txt")

	file, err := os.Open(userArgsPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	foundXmx := false
	foundXms := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			lines = append(lines, line)
			continue
		}

		if strings.HasPrefix(line, "-Xms") {
			lines = append(lines, fmt.Sprintf("-Xms%s", cfg.MemoryAllocated))
			foundXms = true
		} else if strings.HasPrefix(line, "-Xmx") {
			lines = append(lines, fmt.Sprintf("-Xmx%s", cfg.MemoryMax))
			foundXmx = true
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if !foundXms {
		lines = append(lines, fmt.Sprintf("-Xms%s", cfg.MemoryAllocated))
	}
	if !foundXmx {
		lines = append(lines, fmt.Sprintf("-Xmx%s", cfg.MemoryMax))
	}

	output := strings.Join(lines, "\n")
	return os.WriteFile(userArgsPath, []byte(output), 0644)
}

func ScriptGetJVMArgs(server string) error {

	cfg, err := Load(server)
	if err != nil {
		return err
	}

	userArgsPath := filepath.Join(paths.Server(server), "user_jvm_args.txt")

	file, err := os.Open(userArgsPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	foundXmx := false
	foundXms := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			lines = append(lines, line)
			continue
		}

		if strings.HasPrefix(line, "-Xms") {
			cfg.MemoryAllocated, _ = strings.CutPrefix(line, "-Xms")
			foundXms = true
		} else if strings.HasPrefix(line, "-Xmx") {
			cfg.MemoryMax, _ = strings.CutPrefix(line, "-Xmx")
			foundXmx = true
		} else {
			cfg.AdditionalJVMArgs = append(cfg.AdditionalJVMArgs, line)
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if !foundXms {
		cfg.MemoryAllocated = ""
	}
	if !foundXmx {
		cfg.MemoryMax = ""
	}

	return Save(server, cfg)
}

func ScriptAddJVMArg(server, arg string) error {

	userArgsPath := filepath.Join(paths.Server(server), "user_jvm_args.txt")

	if strings.ContainsAny(arg, "#\"/~|") {
		return fmt.Errorf("invalid character in argument %q", arg)
	}

	data, err := os.ReadFile(userArgsPath)
	if err != nil {
		return fmt.Errorf("Has a correct jar file been downloaded and installed? %s", err.Error())
	}

	output := string(data) + "\n" + arg + "\n"

	return os.WriteFile(userArgsPath, []byte(output), 0644)
}

func ScriptRemoveJVMArg(server string, args []string, index int) error {

	if index < 0 || index >= len(args) {
		return fmt.Errorf("index %d out of range", index+1)
	}

	arg := args[index]
	userArgsPath := filepath.Join(paths.Server(server), "user_jvm_args.txt")

	file, err := os.Open(userArgsPath)
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

		if strings.EqualFold(line, arg) {
			ui.PrintInfo("Removing argument " + arg + " from " + userArgsPath)
			found = true
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("could not find argument %s", arg)
	}

	output := strings.Join(lines, "\n")
	return os.WriteFile(userArgsPath, []byte(output), 0644)
}

func ForgeIsVersionScriptBased(version string) (bool, error) {
	version, _ = strings.CutPrefix(version, "1.")

	versionFloat, err := strconv.ParseFloat(version, 32)
	if err != nil {
		return false, err
	}

	return (int(versionFloat) >= 17), nil
}
