//go:build linux

package main

import (
	"fmt"
	"minecraft-manager/internal/client"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/create"
	"minecraft-manager/internal/daemon"
	"minecraft-manager/internal/java"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/templates"
	"minecraft-manager/internal/ui"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		ui.PrintMainHelpMessage()
		os.Exit(0)
	}

	if err := run(); err != nil {
		ui.PrintError(err.Error() + "\n")
		os.Exit(1)
	}
}

func run() error {

	switch os.Args[1] {

	case "start":
		if len(os.Args) != 3 {
			return fmt.Errorf("usage: manager start [server]")
		}
		return client.StartServer(os.Args[2])

	case "create":
		if len(os.Args) < 5 {
			return fmt.Errorf("usage: manager create [server] [type] [version]")
		}
		return create.Create(os.Args)

	case "daemon":
		if len(os.Args) == 3 {
			paths.SetServerRoot(os.Args[2])
		} else if len(os.Args) != 2 {
			return fmt.Errorf("usage: manager daemon [server-root (optional)]")
		}
		d := daemon.New()
		return d.Run()

	case "ping":
		return client.PingDaemon()

	case "ps":
		return client.GetPS()

	case "list":
		return client.GetList()

	case "screen":
		if len(os.Args) != 3 {
			return fmt.Errorf("usage: manager screen <server>")
		}
		return client.Screen(os.Args[2])

	case "stop":
		if len(os.Args) != 3 {
			return fmt.Errorf("usage: manager stop <server>")
		}
		return client.StopServer(os.Args[2])

	case "kill":
		if len(os.Args) != 3 {
			return fmt.Errorf("usage: manager kill <server>")
		}
		return client.KillServer(os.Args[2])

	case "set":
		if len(os.Args) < 5 {
			if len(os.Args) == 3 && os.Args[2] == "help" {
				return ui.PrintSetHelpMessage()
			}
			return fmt.Errorf("usage: manager set <server> <parameter> <argument>")
		}
		return client.SetParameter(os.Args)

	case "set-property":
		if len(os.Args) != 5 {
			return fmt.Errorf("usage: manager set-property <server> <key> <value>")
		}
		return client.SetProperty(os.Args[2], os.Args[3], os.Args[4])

	case "download":
		if len(os.Args) != 4 {
			return fmt.Errorf("usage: manager download <server> <URL>")
		}
		return client.DownloadJarToServer(os.Args[2], os.Args[3])

	case "import":
		if len(os.Args) < 5 {
			return fmt.Errorf("usage: manager import <server> <type> <version>")
		}
		return create.ImportServer(os.Args)

	case "add-launch-arg", "ala":

		if len(os.Args) != 5 {
			return fmt.Errorf("usage: manager %s <jvm/serv> <server> <arg>", os.Args[1])
		}
		switch os.Args[2] {
		case "jvm":
			return config.AddAdditionalJVMArg(os.Args[3], os.Args[4])
		case "serv":
			return config.AddAdditionalServArg(os.Args[3], os.Args[4])
		default:
			return fmt.Errorf("usage: manager %s <jvm/serv> <server> <arg>", os.Args[1])
		}

	case "rem-launch-arg", "rla":
		if len(os.Args) != 5 {
			return fmt.Errorf("usage: manager %s <jvm/serv> <server> <index>", os.Args[1])
		}
		switch os.Args[2] {
		case "jvm":
			return config.RemoveAdditionalJVMArg(os.Args[3], os.Args[4])
		case "serv":
			return config.RemoveAdditionalServArg(os.Args[3], os.Args[4])
		default:
			return fmt.Errorf("usage: manager %s <jvm/serv> <server> <arg>", os.Args[1])
		}

	case "help", "--help", "-?", "?":

		if len(os.Args) == 3 && os.Args[2] == "type" {
			return ui.PrintTypeHelpMessage()
		} else if len(os.Args) == 3 && os.Args[2] == "java" {
			return ui.PrintJavaHelpMessage()
		}

		return ui.PrintMainHelpMessage()

	case "inspect":

		if len(os.Args) != 3 {
			return fmt.Errorf("usage: manager inspect <server>")
		}

		return client.InspectServer(os.Args[2])

	case "grace-period":

		if len(os.Args) != 3 {
			return fmt.Errorf("usage: manager grace-period <grace-period-in-seconds>")
		}

		return client.SetGracePeriod(os.Args[2])

	case "get-sample-cfg":

		return templates.PrintConfigFile()

	case "install-java", "ijv":

		if len(os.Args) < 3 {
			return fmt.Errorf("usage: manager install-java [version]")
		}

		return java.InstallJavaVersion(os.Args[2])

	case "install-custom-java":

		if len(os.Args) < 4 {
			return fmt.Errorf("usage: manager install-custom-java [version] [url]")
		}

		return java.InstallCustomJavaVersion(os.Args[2], os.Args[3])

	case "list-java":

		return ui.PrintInstalledJavaVersions()

	case "reload-prop":
		if len(os.Args) < 3 {
			return fmt.Errorf("usage: manager reload-prop [server]")
		}

		return client.ReloadProperties(os.Args[2])

	case "reload-user":
		if len(os.Args) < 3 {
			return fmt.Errorf("usage: manager reload-user [server]")
		}

		return client.ReloadUser(os.Args[2])

	case "motd":

		return fmt.Errorf("not yet implemented") // TODO: motd generator

	case "backup":

		return fmt.Errorf("not yet implemented") // TODO

	default:
		return fmt.Errorf("Unknown Command %q", os.Args[1])
	}
}
