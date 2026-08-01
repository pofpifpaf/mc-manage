package paths

import "path/filepath"

const (
	ServerRoot       = "/servers"
	JavaRoot         = "/opt/java"
	SocketPath       = "/run/minecraft-manager.sock"
	Logs             = "logs/latest.log"
	ScreenSocketPath = "/run/minecraft-manager-screen.sock"
	ConfigJson       = "config.json"
	TemplatesDir     = "templates"
	ServProperties   = "server.properties"
	EulaTxt          = "eula.txt"
)

func Server(name string) string {
	return filepath.Join(ServerRoot, name)
}

func Config(name string) string {
	return filepath.Join(Server(name), ConfigJson)
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
