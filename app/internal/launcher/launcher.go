package launcher

import (
	"fmt"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/java"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"minecraft-manager/internal/ui"
	"os"
	"os/exec"
)

func Build(server string) (*exec.Cmd, bool, string, error) {
	cfg, err := config.Load(server)
	if err != nil {
		return nil, false, "", err
	}

	javaPath, err := java.Find(cfg.Java)
	if err != nil {
		return nil, false, "", err
	}

	var cmd *exec.Cmd
	switch cfg.Type {
	case "neoforge":
		cmd, err = buildRunSH(cfg, "run.sh")
		if err != nil {
			return nil, false, "", err
		}
	default:
		cmd, err = buildJarFile(cfg, javaPath)
		if err != nil {
			return nil, false, "", err
		}
	}

	ui.PrintInfo(fmt.Sprintf("Starting %s, with Java Path: %s and Server Directory : %s ", cfg.Name, javaPath, paths.Server(cfg.Name)))

	return cmd, cfg.AutomaticRestarts, cfg.Port, nil
}

func buildJarFile(cfg *protocol.Config, javaPath string) (*exec.Cmd, error) {
	serverDir := paths.Server(cfg.Name)

	jarPath := paths.Jar(cfg.Name, cfg.Jar)

	if _, err := os.Stat(jarPath); err != nil {
		return nil, fmt.Errorf("jar not found: %s", jarPath)
	}

	args := []string{
		"-Xms" + cfg.MemoryAllocated,
		"-Xmx" + cfg.MemoryMax,
	}

	args = append(args, cfg.AdditionalJVMArgs...)

	args = append(args,
		"-jar",
		cfg.Jar,
		"nogui",
	)

	args = append(args, cfg.AdditionalServArgs...)

	cmd := exec.Command(javaPath, args...)

	cmd.Dir = serverDir

	return cmd, nil
}

func buildRunSH(cfg *protocol.Config, scriptName string) (*exec.Cmd, error) {

	serverDir := paths.Server(cfg.Name)

	runPath := paths.Jar(cfg.Name, scriptName)

	if _, err := os.Stat(runPath); err != nil {
		return nil, fmt.Errorf("run script not found: %s", runPath)
	}

	cmd := exec.Command(runPath, cfg.AdditionalServArgs...)

	cmd.Dir = serverDir

	return cmd, nil
}
