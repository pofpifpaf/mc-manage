package paths

import "path/filepath"

const (
	ServerRoot = "/servers"
	JavaRoot   = "/opt/java"
	SocketPath = "/tmp/minecraft-manager.sock"
)

func Server(name string) string {
	return filepath.Join(ServerRoot, name)
}

func Config(name string) string {
	return filepath.Join(Server(name), "config.json")
}

func Jar(name, jar string) string {
	return filepath.Join(Server(name), jar)
}

func Java(version string) string {
	return filepath.Join(
		JavaRoot,
		version,
		"bin",
		"java",
	)
}