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
	"strconv"
	"strings"
	"sync"

	"golang.org/x/term"
)

type ServerInfo struct {
	Name string
}

type ServerListResponse struct {
	OK      bool         `json:"ok"`
	Message string       `json:"message,omitempty"`
	Data    []ServerInfo `json:"data"`
}

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

	if !resp.OK {
		return errors.New(resp.Message)
	}

	switch req.Command {

	case "LIST":

		raw := resp.Data.([]interface{})

		servers := make([]ServerInfo, len(raw))

		for i, v := range raw {
			m := v.(map[string]interface{})

			b, err := json.Marshal(m)
			if err != nil {
				return err
			}

			var server ServerInfo
			if err := json.Unmarshal(b, &server); err != nil {
				return err
			}

			servers[i] = server
		}

		for _, server := range servers {
			fmt.Printf("%s\n", server.Name)
		}

	case "START", "STOP":
		fmt.Printf("Daemon responded without error - %s\n", resp.Message)

	default:
		fmt.Printf("Response: %q\n", resp.Message)
	}

	return nil
}

func SetParameter(server string, arg1 string, arg2 string) error {

	switch arg1 {
	case "port":

		port, err := strconv.Atoi(arg2)
		if err != nil || (port < 0 || port > 65535) {
			return errors.New("port number out of range, must be between 0 and 65535")
		}

		return Send(protocol.Request{
			Command: "SET",
			Server:  server,
			Text:    arg1,
			Data:    port,
		})

	case "autorestart":

		if arg2 == "true" || arg2 == "false" {
			return Send(protocol.Request{
				Command: "SET",
				Server:  server,
				Text:    arg1,
				Data:    arg2,
			})
		} else {
			return errors.New("unrecognized argument to autorestart")
		}

	default:
		return errors.New("Incorrect set paramater")
	}
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
