package client

import (
	"fmt"
	"minecraft-manager/internal/paths"
	"os"
	"text/tabwriter"
	"time"
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
