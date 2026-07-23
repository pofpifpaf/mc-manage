package download

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"io"
)

type versionManifest struct {
	Versions []manifestVersion `json:"versions"`
}

type manifestVersion struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type versionInfo struct {
	Downloads struct {
		Server struct {
			URL  string `json:"url"`
			SHA1 string `json:"sha1"`
		} `json:"server"`
	} `json:"downloads"`
}

const manifestURL = "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"

func manifest() (*versionManifest, error) {
	resp, err := http.Get(manifestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	var m versionManifest

	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

func VersionMetadataURL(version string) (string, error) {
	m, err := manifest()
	if err != nil {
		return "", err
	}

	for _, v := range m.Versions {
		if v.ID == version {
			return v.URL, nil
		}
	}

	return "", fmt.Errorf("minecraft version %s not found", version)
}

func DownloadVanilla(version, destination string) (error) {

	fmt.Printf("Downloading %q for version %q\n", destination, version)

	metadataURL, err := VersionMetadataURL(version)
	if err != nil {
		return err
	}

	resp, err := http.Get(metadataURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	var info versionInfo

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return err
	}

	if info.Downloads.Server.URL == "" {
		return fmt.Errorf("version %s has no server download", version)
	}

	resp, err = http.Get(info.Downloads.Server.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Downloading from url %q\n", info.Downloads.Server.URL)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	fmt.Printf("Copying to %q\n", destination)

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}