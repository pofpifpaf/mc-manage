package download

import (
	"encoding/json"
	"fmt"
	"io"
	"minecraft-manager/internal/ui"
	"net/http"
	"os"
	"slices"
)

type purpurManifestBuildVersion struct {
	Builds struct {
		Latest string   `json:"latest"`
		All    []string `json:"all"`
	} `json:"builds"`
}

const purpurBuildNumberDisplayed = 10

func purpurPrintAvailableVersions(version string) error {

	downloadURL := "https://api.purpurmc.org/v2/purpur/" + version

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var manifest purpurManifestBuildVersion

	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return err
	}

	if manifest.Builds.All == nil {
		return fmt.Errorf("No builds available for this version")
	}

	ui.PrintError("Couldn't find the specified build number")
	ui.PrintInfo(fmt.Sprintf("The last %d available builds for this version are :\n", purpurBuildNumberDisplayed))

	slices.Sort(manifest.Builds.All)
	slices.Reverse(manifest.Builds.All)

	for i, build := range manifest.Builds.All {
		fmt.Printf("    %s\n", build)
		if i > purpurBuildNumberDisplayed {
			fmt.Println("    ...")
			break
		}
	}

	fmt.Print("\n")

	return nil
}

func DownloadPurpur(version, build, destination string) error {

	ui.PrintInfo(fmt.Sprintf("Downloading %q for version %q", destination, version))

	downloadURL := "https://api.purpurmc.org/v2/purpur/" + version + "/"

	if build == "" {
		downloadURL = downloadURL + "latest/download"
	} else {
		downloadURL = downloadURL + build + "/download"
	}

	ui.PrintInfo(fmt.Sprintf("Downloading from url %q", downloadURL))

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_ = purpurPrintAvailableVersions(version)
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
