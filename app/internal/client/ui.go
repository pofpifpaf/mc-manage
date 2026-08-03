package client

import (
	"fmt"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/paths"
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

func printRunningServers(servers []ServerInfo) {

	fmt.Printf("\n")
	defer fmt.Print("\n")

	if len(servers) == 0 {
		fmt.Println("No servers running")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tVERSION\tJAVA\tPORT\tAUTO RESTART\tBOOT\tSIZE\tUPTIME")
	fmt.Fprintln(w, "----\t-------\t----\t----\t------------\t----\t----\t------")

	for _, server := range servers {
		dirSize, _ := paths.DirSize(paths.Server(server.Name))
		uptime := time.Since(server.StartedAt).Round(time.Second)

		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%t\t%s\t%s\n",
			server.Name,
			server.Version,
			server.JavaVersion,
			server.Port,
			server.AutomaticRestarts,
			server.StartOnBoot,
			dirSize,
			uptime,
		)
	}
}

func printServerList(servers []ServerInfo) {

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
			"%s\t%s\t%s\t%s\t%t\t%s\t%t\t%t\n",
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

func printInspectServer(server ServerInfo, cfg *config.Config) {

	fmt.Print(("\n"))
	defer fmt.Print("\n")

	fmt.Printf("Info for server: %s\n\n", cfg.Name)

	if server.Running {
		fmt.Println("status: " + cGreen + "running" + ansiReset)
	} else {
		fmt.Println("status: " + cRed + "stopped" + ansiReset)
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

	if server.Running {
		uptime := time.Since(server.StartedAt).Round(time.Second)
		fmt.Fprintf(w, "Uptime\t%s\n", uptime)
	} else {
		fmt.Fprintln(w, "Uptime\t-")
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
