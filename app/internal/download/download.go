package download

import (
	"bufio"
	"fmt"
	"io"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/java"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"minecraft-manager/internal/ui"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

func RemoveRecommendedJVMArguments(cfg *protocol.Config) error {

	switch cfg.Type {
	case "paper":
		return paperRemoveRecommendedJVMArguments(cfg)
	}

	return nil
}

func RetrieveJarIfArchived(cfg *protocol.Config) (bool, error) {

	if cfg.Type == "neoforge" || cfg.Type == "forge" {
		return false, nil
	}

	var jarArchivePath string
	if cfg.VersionArg == "" {
		jarArchivePath = filepath.Join(paths.Server(cfg.Name), cfg.Jar+"."+cfg.Version+".old")
	} else {
		jarArchivePath = filepath.Join(paths.Server(cfg.Name), cfg.Jar+"."+cfg.Version+"-"+cfg.VersionArg+".old")
	}

	if _, err := os.Stat(jarArchivePath); err == nil {
		if err := os.Rename(jarArchivePath, paths.Jar(cfg.Name, cfg.Jar)); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func DownloadJar(cfg *protocol.Config) error {
	var err error

	if isArchived, _ := RetrieveJarIfArchived(cfg); isArchived {
		ui.PrintInfo(fmt.Sprintf("Retrieved config file from archive for server %s and version %s", cfg.Name, cfg.Version))
		switch cfg.Type {
		case "paper":
			if err := paperSetRecommendedJVMArguments(cfg); err != nil {
				ui.PrintWarning("could not set recommended jvm args: " + err.Error())
			}
		}
		return nil
	}

	switch cfg.Type {
	case "vanilla":
		if err := DownloadVanilla(cfg.Version, paths.Jar(cfg.Name, cfg.Jar)); err != nil {
			return err
		}
	case "paper":
		if err := DownloadPaper(cfg.Version, cfg.VersionArg, paths.Jar(cfg.Name, cfg.Jar)); err != nil {
			return err
		}
		if err := paperSetRecommendedJVMArguments(cfg); err != nil {
			ui.PrintWarning("could not set recommended jvm args: " + err.Error())
		}
	case "purpur":
		if err := DownloadPurpur(cfg.Version, cfg.VersionArg, paths.Jar(cfg.Name, cfg.Jar)); err != nil {
			return err
		}
	case "neoforge":
		if err := InstallNeoforge(cfg); err != nil {
			return err
		}
	case "fabric":
		if err := DownloadFabric(cfg.Version, cfg.VersionArg, paths.Jar(cfg.Name, cfg.Jar)); err != nil {
			return err
		}
	case "forge":
		if err := InstallForge(cfg); err != nil {
			return err
		}
	default:
		ui.PrintError("\"" + cfg.Type + "\" Unsupported type")
	}

	return err
}

func DownloadCustomJar(cfg *protocol.Config, downloadURL string) error {

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ui.PrintInfo("Downloading from url \"" + downloadURL + "\"")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	destination := paths.Jar(cfg.Name, cfg.Jar)

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	switch cfg.Type {
	case "neoforge", "forge":
		if err := runInstaller(cfg); err != nil {
			return err
		}

		return config.ConfigureJavaRunScript(cfg.Name)
	default:
		return nil
	}
}

func ArchiveJarFile(cfg *protocol.Config) error {

	oldPath := paths.Jar(cfg.Name, cfg.Jar)
	var newPath string
	if cfg.VersionArg == "" {
		newPath = filepath.Join(paths.Server(cfg.Name), cfg.Jar+"."+cfg.Version+".old")
	} else {
		newPath = filepath.Join(paths.Server(cfg.Name), cfg.Jar+"."+cfg.Version+"-"+cfg.VersionArg+".old")
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	return nil
}

func forgeSetupServerJarFile(cfg *protocol.Config) error {

	ui.PrintInfo("Version predates 1.17, handling with jar instead of run script")

	if err := os.Remove(paths.Jar(cfg.Name, cfg.Jar)); err != nil {
		return err
	}

	serverJarFileName := "minecraft_server." + cfg.Version + ".jar"

	if err := os.Rename(paths.Jar(cfg.Name, serverJarFileName), paths.Jar(cfg.Name, cfg.Jar)); err != nil {
		return err
	}

	return nil
}

func runInstaller(cfg *protocol.Config) error {

	server := cfg.Name

	javaPath, err := java.Find(cfg.Java)
	if err != nil {
		return err
	}

	serverDir := paths.Server(server)

	jarPath := paths.Jar(server, cfg.Jar)

	if _, err := os.Stat(jarPath); err != nil {
		return fmt.Errorf("jar not found: %s", jarPath)
	}

	args := []string{
		"-jar",
		cfg.Jar,
		"--installServer",
	}

	cmd := exec.Command(javaPath, args...)

	cmd.Dir = serverDir

	ui.PrintInfo(fmt.Sprintf("Starting Installer %q, with Java Path: %s and Server Directory : %s ", cfg.Jar, javaPath, serverDir))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		reader := bufio.NewReader(stdout)
		var current strings.Builder

		width := 80
		w, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil {
			width = w
		}

		for {
			b, err := reader.ReadByte()
			if err != nil {
				return
			}

			switch b {
			case '\r', '\n':
				if current.Len() > 0 {
					text := strings.TrimSpace(current.String())

					if len(text) > width {
						text = text[:width-3] + "..."
					}
					fmt.Printf("\r\033[2K%s", text)
					current.Reset()
				}

			default:
				current.WriteByte(b)
			}
		}
	}()

	err = cmd.Wait()

	fmt.Print("\r\033[2K")

	if err != nil {
		ui.PrintError(fmt.Sprintf("Installer exited with error: %v", err))
	}
	ui.PrintSuccess("Installer exited normally")

	cfg.AdditionalServArgs = append(cfg.AdditionalServArgs, "nogui")

	script, err := config.ForgeIsVersionScriptBased(cfg.Version)
	if err != nil {
		ui.PrintWarning(err.Error())
	}
	if !script {
		if err := forgeSetupServerJarFile(cfg); err != nil {
			ui.PrintWarning(err.Error())
		}
	} else if err := config.ScriptSetJVMMemoryArgs(server); err != nil {
		ui.PrintWarning("could not set memory arguments : " + err.Error())
	}

	return config.Save(server, cfg)
}
