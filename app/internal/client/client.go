package client

import (
	"bufio"
	"fmt"
	"net"

	"minecraft-manager/internal/paths"
)

func Send(command string) error {
	conn, err := net.Dial("unix", paths.SocketPath)
	if err != nil {
		return err
	}

	defer conn.Close()

	fmt.Fprintln(conn, command)

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return nil
}

func Start(name string) error {
	return Send("START " + name)
}