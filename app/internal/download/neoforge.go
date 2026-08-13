package download

import (
	"bufio"
	"encoding/xml"
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
	"sort"
	"strconv"
	"strings"

	"golang.org/x/term"
)

type NeoforgeMetadata struct {
	Versioning struct {
		Latest   string `xml:"latest"`
		Versions struct {
			Version []string `xml:"version"`
		} `xml:"versions"`
	} `xml:"versioning"`
}

func neoforgeDisplayAdjacentBuilds(build string) error {

	releasesURL := "https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml"

	resp, err := http.Get(releasesURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var metadata NeoforgeMetadata

	if err := xml.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return err
	}

	ui.PrintInfo("Latest available Neoforge version is " + metadata.Versioning.Latest)

	parts := strings.Split(build, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid NeoForge version: %s", build)
	}

	minecraftVersion := strings.Join(parts[:2], ".")

	var builds []string

	for _, version := range metadata.Versioning.Versions.Version {
		if !strings.HasPrefix(version, minecraftVersion+".") {
			continue
		}

		buildNumber := strings.TrimPrefix(version, minecraftVersion+".")

		if _, err := strconv.Atoi(buildNumber); err != nil {
			continue
		}

		builds = append(builds, version)
	}

	sort.Slice(builds, func(i, j int) bool {
		a, _ := strconv.Atoi(strings.TrimPrefix(builds[i], minecraftVersion+"."))
		b, _ := strconv.Atoi(strings.TrimPrefix(builds[j], minecraftVersion+"."))

		return a > b
	})

	if len(builds) > 10 {
		builds = builds[:10]
	}

	ui.PrintInfo(fmt.Sprintf("Latest 10 builds available for Minecraft %s:", minecraftVersion))

	for _, version := range builds {
		fmt.Printf("     %s\n", version)
	}

	return nil
}

func downloadNeoforgeInstaller(version, destination string) error {

	ui.PrintInfo(fmt.Sprintf("Downloading %q for version %q", destination, version))

	downloadURL := "https://maven.neoforged.net/releases/net/neoforged/neoforge/" + version + "/neoforge-" + version + "-installer.jar"

	ui.PrintInfo(fmt.Sprintf("Downloading from url %q", downloadURL))

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ui.PrintError("Could not get jar, is server version correct?")
		neoforgeDisplayAdjacentBuilds(version)
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func neoforgeSelectJavaVersion(server string) error {
	cfg, err := config.Load(server)
	if err != nil {
		return err
	}

	// TODO: Find java version here

	return config.Save(server, cfg)
}

func neoforgeRunInstaller(server string) error {

	cfg, err := config.Load(server)
	if err != nil {
		return err
	}

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

	ui.PrintInfo(fmt.Sprintf("Starting Installer %s, with Java Path: %s and Server Directory : %s ", cfg.Name, javaPath, serverDir))

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

	if err := config.NeoforgeSetJVMMemoryArgs(server); err != nil {
		ui.PrintWarning("could not set memory arguments : " + err.Error())
	}

	return config.Save(server, cfg)
}

func InstallNeoforge(cfg *protocol.Config) error {
	if err := downloadNeoforgeInstaller(cfg.Version, paths.Jar(cfg.Name, cfg.Jar)); err != nil {
		return err
	}

	if err := neoforgeSelectJavaVersion(cfg.Name); err != nil {
		return err
	}

	if err := neoforgeRunInstaller(cfg.Name); err != nil {
		return err
	}

	return config.NeoforgeConfigureJavaRunScript(cfg.Name)
}
