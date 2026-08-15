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
	"slices"
	"strings"
)

type ForgeMetadata struct {
	Versioning struct {
		Latest   string `xml:"latest"`
		Versions struct {
			Version []string `xml:"version"`
		} `xml:"versions"`
	} `xml:"versioning"`
}

func forgeGetCompatibleVersions(version string) ([]string, error) {
	ui.PrintInfo("Getting latest release for version " + version)

	manifestURL := "https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml"

	resp, err := http.Get(manifestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var metadata ForgeMetadata

	if err := xml.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, err
	}

	var versions []string

	for _, metadataVersion := range metadata.Versioning.Versions.Version {
		output, in := strings.CutPrefix(metadataVersion, version+"-")
		if in {
			versions = append(versions, output)
		}
	}

	if len(versions) < 1 {
		return nil, fmt.Errorf("No forge version found for version %s", version)
	}

	slices.Sort(versions)
	slices.Reverse(versions)

	return versions, nil
}

func forgeGetLatestDownloadURL(version string) (string, string, error) {

	versions, err := forgeGetCompatibleVersions(version)
	if err != nil {
		return "", "", err
	}

	downloadURL := "https://maven.minecraftforge.net/net/minecraftforge/forge/" + version + "-" + versions[0] + "/forge-" + version + "-" + versions[0] + "-installer.jar"

	return downloadURL, versions[0], nil
}

func forgeDisplayAdjacentBuilds(version string) error {
	versions, err := forgeGetCompatibleVersions(version)
	if err != nil {
		return err
	}

	if len(versions) > 10 {
		versions = versions[:10]
	}

	ui.PrintInfo(fmt.Sprintf("Latest 10 versions available for Minecraft %s:", version))

	for _, metadataVersion := range versions {
		fmt.Printf("     %s\n", metadataVersion)
	}

	return nil
}

func ForgeExtractVersionFromForge(link string) (string, string, error) {

	buf, _ := strings.CutPrefix(link, "https://maven.minecraftforge.net/net/minecraftforge/forge/")

	bufs := strings.Split(buf, "-")

	if len(bufs) < 2 {
		return "", "", fmt.Errorf("Malformed link")
	}

	version := bufs[0]
	versionArg := bufs[1]

	versionArg, _ = strings.CutSuffix(versionArg, "/forge")

	ui.PrintInfo(fmt.Sprintf("From link, found minecraft version %s and forge version %s", version, versionArg))

	return version, versionArg, nil
}

func downloadForgeInstaller(cfg *protocol.Config) error {

	version := cfg.Version
	forgeVersion := cfg.VersionArg
	destination := paths.Jar(cfg.Name, cfg.Jar)

	if version != "link" {
		ui.PrintWarning("Consider using \"manager [create/import] forge link <link>\" to input download link directly and support Minecraft Forge creators")
	}

	var downloadURL string
	downloadURL = "https://maven.minecraftforge.net/net/minecraftforge/forge/" + version + "-" + forgeVersion + "/forge-" + version + "-" + forgeVersion + "-installer.jar"

	if forgeVersion == "" {
		var err error
		downloadURL, cfg.VersionArg, err = forgeGetLatestDownloadURL(version)
		if err != nil {
			return err
		}
	} else if version == "link" {
		downloadURL = forgeVersion
		var err error
		cfg.Version, cfg.VersionArg, err = ForgeExtractVersionFromForge(downloadURL)
		if err != nil {
			return err
		}
	}

	ui.PrintInfo(fmt.Sprintf("Downloading from url %q", downloadURL))

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ui.PrintError("Could not get jar, is server version correct?")
		forgeDisplayAdjacentBuilds(version)
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

func InstallForge(cfg *protocol.Config) error {
	if err := downloadForgeInstaller(cfg); err != nil {
		return err
	}

	if err := runInstaller(cfg); err != nil {
		return err
	}

	return config.ConfigureJavaRunScript(cfg.Name)
}
