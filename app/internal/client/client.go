package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"net"
	"os"
	"strings"

	"golang.org/x/term"
)

func Send(req protocol.Request) error {

	conn, err := net.Dial("unix", paths.SocketPath)
	if err != nil {
		return err
	}

	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return err
	}

	var resp protocol.Response

	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return err
	}

	fmt.Printf("Response: %q\n", resp.Message)

	if req.Command == "LIST" {

		raw := resp.Data.([]interface{})

		servers := make([]string, len(raw))
		for i, v := range raw {
			servers[i] = v.(string)
		}

		fmt.Println(servers)
	}

	if !resp.OK {
		return errors.New(resp.Message)
	}

	return nil
}

func Screen(server string) error {
	conn, err := net.Dial("unix", paths.ScreenSocketPath)
	if err != nil {
		return fmt.Errorf("connect screen socket: %w", err)
	}

	fmt.Println("Connected to screen socket")

	req := protocol.Request{
		Command: "SCREEN",
		Server:  server,
	}

	if err := json.NewEncoder(conn).Encode(&req); err != nil {
		conn.Close()
		return fmt.Errorf("send screen request: %w", err)
	}

	fmt.Println("Sent screen socket request")

	var resp protocol.Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		conn.Close()
		return fmt.Errorf("read screen response: %w", err)
	}

	if !resp.OK {
		conn.Close()
		return errors.New(resp.Message)
	}

	fd := int(os.Stdin.Fd())

	if !term.IsTerminal(fd) {
		conn.Close()
		return errors.New("screen requires an interactive terminal")
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		conn.Close()
		return fmt.Errorf("set raw mode: %w", err)
	}

	defer func() {
		_ = term.Restore(fd, oldState)
	}()

	fmt.Print("Attached. Press Ctrl+] to detach\r\n")

	done := make(chan error, 2)

	// Terminal -> server
	go func() {
		buf := make([]byte, 1)

		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				done <- err
				return
			}

			// Ctrl+] detach
			if n == 1 && buf[0] == 0x1d {
				done <- nil
				return
			}

			if _, err := conn.Write(buf[:n]); err != nil {
				done <- err
				return
			}
		}
	}()

	// Server -> terminal
	go func() {
		_, err := io.Copy(os.Stdout, conn)
		done <- err
	}()

	err = <-done
	_ = conn.Close()
	otherErr := <-done

	if err != nil && !isExpectedDisconnect(err) {
		return err
	}

	if otherErr != nil && !isExpectedDisconnect(otherErr) {
		return otherErr
	}

	return nil
}

func isExpectedDisconnect(err error) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "use of closed network connection") ||
		strings.Contains(msg, "closed network connection")
}
