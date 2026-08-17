package java

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/ui"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type JavaPackages struct {
	Package struct {
		Link string `json:"link"`
	} `json:"package"`
}

type JavaBinaries struct {
	Binaries []JavaPackages `json:"binaries"`
}

func InstallCustomJavaVersion(version, url string) error {

	fmt.Print("\n")
	defer fmt.Print("\n")

	if _, err := os.Stat(paths.Java(version)); err == nil {
		return fmt.Errorf("Java version %s already exists", version)
	}

	if version == "5" {
		return installJava5(url)
	}

	return downloadAndExtractJava(url, version)
}

func InstallJavaVersion(version string) error {

	fmt.Print("\n")
	defer fmt.Print("\n")

	if _, err := os.Stat(paths.Java(version)); err == nil {
		return fmt.Errorf("Java version %s already exists", version)
	}

	switch version {
	case "6", "7", "16":
		return installJavaLegacy(version)
	default:
		return installJavaAdoptium(version)
	}
}

func installJavaLegacy(version string) error {

	ui.PrintWarning("Downloading Legacy versions. This version has no SHA256 verification.")

	switch version {
	case "6":
		return downloadAndExtractJava("https://cdn.azul.com/zulu/bin/zulu6.22.0.3-ca-jdk6.0.119-linux_x64.tar.gz", version)
	case "7":
		return downloadAndExtractJava("https://cdn.azul.com/zulu/bin/zulu7.25.0.5-ca-jdk7.0.201-linux_x64.tar.gz", version)
	case "16":
		return downloadAndExtractJava("https://github.com/adoptium/temurin16-binaries/releases/download/jdk-16.0.2%2B7/OpenJDK16U-jdk_x64_linux_hotspot_16.0.2_7.tar.gz", version)
	}

	return fmt.Errorf("unsupported")
}

func installJava5(url string) error {

	ui.PrintInfo("Creating temporary file")

	tmp, err := os.CreateTemp("", "java5-*.bin")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	ui.PrintInfo("Retrieving " + url)

	resp, err := http.Get(url)
	if err != nil {
		tmp.Close()
		return err
	}
	defer resp.Body.Close()

	//TODO: check SHA256

	if resp.StatusCode != http.StatusOK {
		tmp.Close()
		return fmt.Errorf("download Java 5: %s", resp.Status)
	}

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return err
	}

	workDir, err := os.MkdirTemp("", "java5-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)

	ui.PrintInfo("Running Java 5 installer")

	cmd := exec.Command("bash", tmpPath)
	cmd.Dir = workDir
	cmd.Stdin = strings.NewReader("yes\n")

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("install Java 5: %w\n%s", err, output)
	}

	entries, err := os.ReadDir(workDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		src := filepath.Join(workDir, entry.Name())
		dest := "/opt/java/5"

		if err := os.RemoveAll(dest); err != nil {
			return err
		}

		ui.PrintSuccess("Installed Java 5")

		return os.Rename(src, dest)
	}

	return fmt.Errorf("Java 5 installer did not produce a JDK directory")
}

func downloadAndExtractJava(url string, version string) error {

	ui.PrintInfo("Retrieving from: " + url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//TODO: check SHA256

	ui.PrintInfo("Extracting archive")

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	dest := filepath.Join("/opt/java", version)

	for {
		h, err := tr.Next()
		if err == io.EOF {
			ui.PrintSuccess("Installed Java Version " + version)
			return nil
		}
		if err != nil {
			return err
		}

		name := strings.SplitN(h.Name, "/", 2)
		if len(name) != 2 {
			continue
		}

		path := filepath.Join(dest, name[1])

		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}

		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}

			f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(h.Mode))
			if err != nil {
				return err
			}

			_, err = io.Copy(f, tr)
			if closeErr := f.Close(); err == nil {
				err = closeErr
			}
			if err != nil {
				return err
			}
		}
	}
}

func installJavaAdoptium(version string) error {

	apiURL := fmt.Sprintf(
		"https://api.adoptium.net/v3/assets/feature_releases/%s/ga?architecture=x64&heap_size=normal&image_type=jre&jvm_impl=hotspot&os=linux&page_size=1&sort_order=DESC",
		version,
	)

	ui.PrintInfo("Contacting API: " + apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	var assets []JavaBinaries

	if err := json.NewDecoder(resp.Body).Decode(&assets); err != nil {
		return err
	}
	if len(assets) == 0 || len(assets[0].Binaries) == 0 {
		return fmt.Errorf("Java %s not found", version)
	}

	return downloadAndExtractJava(assets[0].Binaries[0].Package.Link, version)

}
