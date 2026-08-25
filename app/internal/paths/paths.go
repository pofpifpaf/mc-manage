package paths

import (
	"fmt"
	"path/filepath"
)

var serverRoot = "/servers"

const (
	JavaRoot         = "/opt/java"
	SocketPath       = "/run/minecraft-manager.sock"
	Logs             = "logs/latest.log"
	ScreenSocketPath = "/run/minecraft-manager-screen.sock"
	ConfigJson       = "config.json"
	MainConfigJson   = "main-cfg.json"
	TemplatesDir     = "templates"
	ServProperties   = "server.properties"
	EulaTxt          = "eula.txt"
)

func GetServerRoot() string {
	return serverRoot
}

func SetServerRoot(newRoot string) {
	serverRoot = newRoot
}

func Server(name string) string {
	return filepath.Join(serverRoot, name)
}

func Config(name string) string {
	return filepath.Join(Server(name), ConfigJson)
}

func MainConfig() string {
	return filepath.Join(serverRoot, MainConfigJson)
}

func ServerProperties(name string) string {
	return filepath.Join(Server(name), ServProperties)
}

func Eula(name string) string {
	return filepath.Join(Server(name), EulaTxt)
}

func Jar(name, jar string) string {
	return filepath.Join(Server(name), jar)
}

func Log(name string) string {
	return filepath.Join(Server(name), Logs)
}

func PidStatus(pid int) string {
	return filepath.Join("/proc", fmt.Sprintf("%d", pid), "smaps_rollup")
}

func Java(version string) string {
	return filepath.Join(
		JavaRoot,
		version,
		"bin",
		"java",
	)
}

func Templates(file string) string {
	return filepath.Join(TemplatesDir, file)
}

func ModsFolder(name string) string {
	return filepath.Join(Server(name), "mods")
}

func Mod(server, name string) string {
	return filepath.Join(ModsFolder(server), name)
}
