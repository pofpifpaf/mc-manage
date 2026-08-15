package download

import (
	"encoding/xml"
	"fmt"
	"io"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"minecraft-manager/internal/ui"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
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

func InstallNeoforge(cfg *protocol.Config) error {
	if err := downloadNeoforgeInstaller(cfg.Version, paths.Jar(cfg.Name, cfg.Jar)); err != nil {
		return err
	}

	if err := neoforgeSelectJavaVersion(cfg.Name); err != nil {
		return err
	}

	if err := runInstaller(cfg); err != nil {
		return err
	}

	return config.ConfigureJavaRunScript(cfg.Name)
}
