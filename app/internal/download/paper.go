package download

import (
	"encoding/json"
	"fmt"
	"io"
	"minecraft-manager/internal/ui"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"
)

type paperManifestBuildVersion struct {
	ID        int    `json:"id"`
	Channel   string `json:"channel"`
	Downloads struct {
		ServerDefault struct {
			Checksum struct {
				Sha256 string `json:"sha256"`
			} `json:"checksums"`
			URL string `json:"url"`
		} `json:"server:default"`
	} `json:"downloads"`
}

const paperBuildNumberDisplayed = 10

func paperPrintAvailableVersions(builds []paperManifestBuildVersion) error {

	if builds == nil {
		return fmt.Errorf("No builds available for this version")
	}

	ui.PrintInfo(fmt.Sprintf("The last %d available builds for this version are :\n", paperBuildNumberDisplayed))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "BUILD NUMBER\tCHANNEL")
	fmt.Fprintln(w, "------------\t-------")

	for i, build := range builds {
		fmt.Fprintf(w, "%d\t%s\n", build.ID, build.Channel)
		if i > paperBuildNumberDisplayed {
			fmt.Fprintln(w, "...\t...")
			break
		}
	}

	fmt.Print("\n")

	return nil
}

func paperManifest(version, buildString string) (string, error) {

	buildNumber, err := strconv.Atoi(buildString)
	if err != nil && buildString != "" {
		ui.PrintWarning("Invalid build number, retrieving default")
		buildString = ""
	}

	URL := "https://fill.papermc.io/v3/projects/paper/versions/" + version + "/builds"
	resp, err := http.Get(URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ui.PrintError("Unable to reach: Is version number correct?")
		return "", fmt.Errorf("unexpected status %s", resp.Status)
	}

	var builds []paperManifestBuildVersion

	if err := json.NewDecoder(resp.Body).Decode(&builds); err != nil {
		return "", err
	}

	if buildString != "" {
		for _, build := range builds {
			if build.ID == buildNumber {
				ui.PrintInfo(fmt.Sprintf("Found %s version %s and build number %d", build.Channel, version, build.ID))
				return build.Downloads.ServerDefault.URL, nil
			}
		}
		paperPrintAvailableVersions(builds)
		return "", fmt.Errorf("build %d not found for Paper %s", buildNumber, version)
	}

	var selected *paperManifestBuildVersion

	for i := range builds {
		b := &builds[i]

		if b.Channel != "STABLE" {
			continue
		}

		if selected == nil || b.ID > selected.ID {
			selected = b
		}
	}

	if selected == nil {
		if err := paperPrintAvailableVersions(builds); err != nil {
			return "", err
		}
		return "", fmt.Errorf("no stable builds found for Paper %s", version)
	}

	ui.PrintInfo(
		fmt.Sprintf(
			"Found %s version %s and build number %d",
			selected.Channel,
			version,
			selected.ID,
		),
	)

	return selected.Downloads.ServerDefault.URL, nil
}

func DownloadPaper(version, build, destination string) error {

	ui.PrintInfo(fmt.Sprintf("Downloading %q for version %q", destination, version))

	downloadURL, err := paperManifest(version, build)
	if err != nil {
		return err
	}

	ui.PrintInfo(fmt.Sprintf("Downloading from url %q", downloadURL))

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
