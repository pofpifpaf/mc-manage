package download

import (
	"encoding/json"
	"fmt"
	"io"
	"minecraft-manager/internal/ui"
	"net/http"
	"os"
	"strings"
)

// Go to https://meta.fabricmc.net/v2/versions/loader/ + Version to get the loader version
// Go to https://meta.fabricmc.net/v2/versions/installer to get the installer versions
// Go to https://meta.fabricmc.net/v2/versions/loader/ + Version + Loader + Installer + /server/jar to get the jar and just launch it

type fabricManifestLoaderVersion struct {
	Loader struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	} `json:"loader"`
}

type fabricManifestInstallerVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

const fabricNumberDisplayed = 5

func fabricPrintAvailableLoaderVersions(version string) error {

	downloadURL := "https://meta.fabricmc.net/v2/versions/loader/" + version

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var manifest []fabricManifestLoaderVersion

	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return err
	}

	if manifest == nil {
		return fmt.Errorf("No builds available for this version")
	}

	ui.PrintInfo(fmt.Sprintf("The last %d available loaders for this version are :\n", fabricNumberDisplayed))

	for i, build := range manifest {
		fmt.Printf("    %s", build.Loader.Version)
		if build.Loader.Stable {
			fmt.Print(", stable")
		}
		fmt.Print("\n")
		if i > fabricNumberDisplayed {
			fmt.Println("    ...")
			break
		}
	}

	fmt.Print("\n")

	return nil
}

func fabricPrintAvailableInstallerVersions() error {

	downloadURL := "https://meta.fabricmc.net/v2/versions/installer/"

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var manifest []fabricManifestInstallerVersion

	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return err
	}

	if manifest == nil {
		return fmt.Errorf("No builds available for this version")
	}

	ui.PrintInfo(fmt.Sprintf("The last %d available installers are :\n", fabricNumberDisplayed))

	for i, build := range manifest {
		fmt.Printf("    %s", build.Version)
		if build.Stable {
			fmt.Print(", stable")
		}
		fmt.Print("\n")
		if i > fabricNumberDisplayed {
			fmt.Println("    ...")
			break
		}
	}

	fmt.Print("\n")

	return nil
}

func fabricGetLatestURL(version, loaderVersion, installerVersion string) (string, error) {

	var selectedLoader string
	selectedLoader = loaderVersion

	if loaderVersion == "" {

		ui.PrintInfo("Getting latest loader for version " + version)

		loadersURL := "https://meta.fabricmc.net/v2/versions/loader/" + version

		resp, err := http.Get(loadersURL)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		var manifest []fabricManifestLoaderVersion

		if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
			return "", err
		}

		for _, build := range manifest {
			if build.Loader.Stable {
				selectedLoader = build.Loader.Version
				break
			}
		}

	}

	var selectedInstaller string
	selectedInstaller = installerVersion

	if installerVersion == "" {

		ui.PrintInfo("Getting latest installer version for version " + version)

		installersURL := "https://meta.fabricmc.net/v2/versions/installer"

		respI, err := http.Get(installersURL)
		if err != nil {
			return "", err
		}
		defer respI.Body.Close()

		var installerManifest []fabricManifestInstallerVersion

		if err := json.NewDecoder(respI.Body).Decode(&installerManifest); err != nil {
			return "", err
		}

		for _, build := range installerManifest {
			if build.Stable {
				selectedInstaller = build.Version
				break
			}
		}
	}

	downloadURL := "https://meta.fabricmc.net/v2/versions/loader/" + version + "/" + selectedLoader + "/" + selectedInstaller + "/server/jar"

	return downloadURL, nil
}

func DownloadFabric(version, versionArg, destination string) error {

	var loader string
	var installer string

	buf := strings.SplitAfter(versionArg, "-")
	switch len(buf) {
	case 1:
		loader = buf[0]
		installer = ""
	case 2:
		loader, _ = strings.CutSuffix(buf[0], "-")
		installer = buf[1]
	default:
		loader = ""
		installer = ""
	}

	ui.PrintInfo(fmt.Sprintf("Downloading %q for version %q", destination, version))

	downloadURL, err := fabricGetLatestURL(version, loader, installer)

	ui.PrintInfo(fmt.Sprintf("Downloading from url %q", downloadURL))

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ui.PrintError("Incorrect Loader or Installer version")
		_ = fabricPrintAvailableLoaderVersions(version)
		_ = fabricPrintAvailableInstallerVersions()
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
