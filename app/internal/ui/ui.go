package ui

import (
	"fmt"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"os"
	"text/tabwriter"
	"time"
)

const (
	ansiReset     = "\033[0m"
	ansiBold      = "\033[1m"
	ansiUnderline = "\033[4m"
	ansiStrike    = "\033[9m"
	ansiItalic    = "\033[3m"

	cRed    = "\033[31m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
	cBlue   = "\033[34m"
	cPurple = "\033[35m"
	cCyan   = "\033[36m"
	cWhite  = "\033[37m"
)

func PrintError(errorStr string) {
	fmt.Printf(cRed+"[ERROR] "+ansiReset+"%s\n", errorStr)
}

func PrintWarning(warnStr string) {
	fmt.Printf(cYellow+"[WARN] "+ansiReset+"%s\n", warnStr)
}

func PrintInfo(infoStr string) {
	fmt.Printf(cBlue+"[INFO] "+ansiReset+"%s\n", infoStr)
}

func PrintSuccess(succStr string) {
	fmt.Printf(cGreen+"[SUCCESS] "+ansiReset+"%s\n", succStr)
}

func PrintRunningServers(servers []protocol.ServerInfo) {

	fmt.Printf("\n")
	defer fmt.Print("\n")

	if len(servers) == 0 {
		fmt.Println("No servers running")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tVERSION\tJAVA\tPORT\tAUTO RESTART\tBOOT\tSIZE\tUPTIME\tPLAYERS\tSTATE")
	fmt.Fprintln(w, "----\t-------\t----\t----\t------------\t----\t----\t------\t-------\t-----")

	for _, server := range servers {

		dirSize, _ := paths.DirSize(paths.Server(server.Name))
		uptime := time.Since(server.StartedAt).Round(time.Second)

		playerString := "-/-"
		if server.PlayersOnlineMax != -1 {
			playerString = fmt.Sprintf("%d/%d", server.PlayersOnline, server.PlayersOnlineMax)
		}

		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%t\t%t\t%s\t%s\t%s\t%s\n",
			server.Name,
			server.Version,
			server.JavaVersion,
			server.Port,
			server.AutomaticRestarts,
			server.StartOnBoot,
			dirSize,
			uptime,
			playerString,
			server.Running,
		)
	}
}

func PrintServerList(servers []protocol.ServerInfo) {

	fmt.Printf("\n")
	defer fmt.Print("\n")

	if len(servers) == 0 {
		fmt.Println("No servers found, to be considered valid, a server needs a config.json file")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tVERSION\tJAVA\tPORT\tAUTO RESTART\tSIZE\tBOOT\tRUNNING")
	fmt.Fprintln(w, "----\t-------\t----\t----\t------------\t----\t----\t-------")

	for _, server := range servers {

		dirSize, _ := paths.DirSize(paths.Server(server.Name))

		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%t\t%s\t%t\t%s\n",
			server.Name,
			server.Version,
			server.JavaVersion,
			server.Port,
			server.AutomaticRestarts,
			dirSize,
			server.StartOnBoot,
			server.Running,
		)
	}
}

func PrintInspectServer(server protocol.ServerInfo, cfg *protocol.Config) {

	fmt.Print(("\n"))
	defer fmt.Print("\n")

	fmt.Printf("Info for server: %s\n\n", cfg.Name)

	switch server.Running {
	case protocol.StateRunning:
		fmt.Println("status: " + cGreen + "running" + ansiReset)
	case protocol.StateStarting:
		fmt.Println("status: " + cYellow + "starting" + ansiReset)
	case protocol.StateStopped:
		fmt.Println("status: " + cRed + "not running" + ansiReset)
	}

	fmt.Print("\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(w, "Type\t%s\n", cfg.Type)
	fmt.Fprintf(w, "Version\t%s\n", cfg.Version)
	fmt.Fprintf(w, "Java Version\t%s\n", cfg.Java)
	fmt.Fprintf(w, "Memory allocation\t%s\n", cfg.Memory)
	fmt.Fprintf(w, "Jar file name\t%s\n", cfg.Jar)
	fmt.Fprintf(w, "Server port\t%s\n", cfg.Port)
	fmt.Fprintf(w, "Level name used\t%s\n", cfg.LevelName)
	fmt.Fprintf(w, "Auto restarts\t%t\n", cfg.AutomaticRestarts)
	fmt.Fprintf(w, "Boot\t%t\n", cfg.StartOnBoot)

	dirSize, _ := paths.DirSize(paths.Server(server.Name))
	fmt.Fprintf(w, "Size\t%s\n", dirSize)

	if server.Running != protocol.StateStopped {
		uptime := time.Since(server.StartedAt).Round(time.Second)
		fmt.Fprintf(w, "Uptime\t%s\n", uptime)
		if server.PlayersOnlineMax == -1 {
			fmt.Fprintln(w, "Players\t-/-")
		} else {
			fmt.Fprintf(w, "Players\t%d/%d\n", server.PlayersOnline, server.PlayersOnlineMax)
		}
	} else {
		fmt.Fprintln(w, "Uptime\t-")
		fmt.Fprintln(w, "Players\t-/-")
	}

	fmt.Fprintln(w, "")

	fmt.Fprint(w, "Additional JVM Args")
	if len(cfg.AdditionalJVMArgs) == 0 {
		fmt.Fprintln(w, "\t-")
	}
	for index, arg := range cfg.AdditionalJVMArgs {
		fmt.Fprintf(w, "\t%d - %s\n", index+1, arg)
	}

	fmt.Fprint(w, "Additional Server Args")
	if len(cfg.AdditionalServArgs) == 0 {
		fmt.Fprintln(w, "\t-")
	}
	for index, arg := range cfg.AdditionalServArgs {
		fmt.Fprintf(w, "\t%d - %s\n", index+1, arg)
	}
}
