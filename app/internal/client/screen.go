package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"minecraft-manager/internal/ui"
	"net"
	"os"
	"strings"
	"sync"

	"golang.org/x/term"
)

func Screen(server string) error {
	conn, err := net.Dial("unix", paths.ScreenSocketPath)
	if err != nil {
		return fmt.Errorf("connect screen socket: %w", err)
	}

	ui.PrintInfo("Connected to screen socket")

	req := protocol.Request{
		Command: "SCREEN",
		Server:  server,
	}

	if err := json.NewEncoder(conn).Encode(&req); err != nil {
		conn.Close()
		return fmt.Errorf("send screen request: %w", err)
	}

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

	fmt.Print("Attached. Press Ctrl+[ to detach\r\n")
	fmt.Print("-------------------------------------\r\n\r\n\r\n")

	done := make(chan error, 2)

	var screenMu sync.Mutex
	var line []byte

	// Terminal -> server
	go func() {
		buf := make([]byte, 1)

		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				done <- err
				return
			}

			// Ctrl+[ detach
			if n == 1 && buf[0] == 0x1d {
				done <- nil
				return
			}
			screenMu.Lock()
			switch buf[0] {
			case '\r', '\n':
				conn.Write(append(line, '\n'))
				line = line[:0]
				redraw(line)

			case 127: // Backspace
				if len(line) > 0 {
					line = line[:len(line)-1]
					redraw(line)
				}

			default:
				line = append(line, buf[0])
				redraw(line)
			}
			screenMu.Unlock()
		}
	}()

	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)
			if n > 0 {
				screenMu.Lock()

				fmt.Print("\r\033[2K")
				os.Stdout.Write(buf[:n])
				redraw(line)

				screenMu.Unlock()
			}

			if err != nil {
				done <- err
				return
			}
		}
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

	fmt.Print("\r\033[2K")

	fmt.Print("\nClosing screening session\n\n")

	fmt.Print("\r\033[2K")

	return nil
}

func redraw(line []byte) {
	fmt.Print("\r\033[2K")

	fmt.Print("Server > ")
	os.Stdout.Write(line)
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
