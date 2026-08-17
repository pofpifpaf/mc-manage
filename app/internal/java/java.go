package java

import (
	"fmt"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/ui"
	"os"
	"strconv"
	"strings"
)

const DefaultJavaVersion = "21"

func Find(version string) (string, error) {
	path := paths.Java(version)

	info, err := os.Stat(path)
	if err != nil {
		ui.PrintError("java " + version + " java %s not installed, use \"manager install-java <version>\"")
		return "", err
	}

	if info.IsDir() {
		return "", fmt.Errorf("%s is not a file", path)
	}

	return path, nil
}

func GetCorrectJavaVersion(version, serverType string) (string, error) {
	parsedVersion, _ := strings.CutPrefix(version, "1.")

	splits := strings.Split(parsedVersion, "-")
	parsedVersion = splits[0]

	splitsDots := strings.SplitAfter(parsedVersion, ".")
	if len(splitsDots) > 2 {
		parsedVersion = splitsDots[0] + splitsDots[1]
	} else if len(splitsDots) == 0 {
		return DefaultJavaVersion, fmt.Errorf("couldn't parse version")
	}

	versionFloat, err := strconv.ParseFloat(parsedVersion, 64)
	if err != nil {
		return DefaultJavaVersion, fmt.Errorf("couldn't parse version")
	}

	switch serverType {
	case "paper", "purpur":
		return paperGetMinecraftRecommendedJava(versionFloat)
	default:
		return vanillaGetMinecraftRecommendedJava(versionFloat)
	}
}

// source: https://minecraft.wiki/w/Tutorial%3AUpdate_Java#:~:text=Why%20update
func vanillaGetMinecraftRecommendedJava(version float64) (string, error) {
	switch {
	case version >= 26:
		return "25", nil
	case version >= 20.5:
		return "21", nil
	case version >= 18:
		return "17", nil
	case version >= 17:
		return "16", nil
	case version >= 12:
		return "8", nil
	case version >= 6.1:
		return "6", nil
	case version >= 1:
		return "5", nil
	default:
		return DefaultJavaVersion, fmt.Errorf("unrecognized minecraft version")
	}
}

// source: https://docs.papermc.io/paper/getting-started/
func paperGetMinecraftRecommendedJava(version float64) (string, error) {
	switch {
	case version >= 26.1:
		return "25", nil
	case version >= 20:
		return "21", nil
	case version >= 17:
		return "17", nil
	case version >= 16.5:
		return "16", nil
	case version >= 12:
		return "11", nil
	case version >= 7.10:
		return "8", nil
	default:
		return DefaultJavaVersion, fmt.Errorf("unrecognized minecraft version")
	}
}
