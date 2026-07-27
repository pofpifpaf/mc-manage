package daemon

import (
	"bufio"
	"fmt"
	"minecraft-manager/internal/launcher"
	"minecraft-manager/internal/ringbuffer"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/creack/pty"
)

type Server struct {
	Name    string
	Cmd     *exec.Cmd
	PTY     *os.File
	Log     *ringbuffer.RingBuffer
	mu      sync.Mutex
	Clients map[*ScreenClient]struct{}
}

type Manager struct {
	servers map[string]*Server
	mutex   sync.Mutex
}

type ScreenClient struct {
	Conn net.Conn
}

func NewManager() *Manager {
	return &Manager{
		servers: make(map[string]*Server),
	}
}

func (m *Manager) Add(server *Server) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.servers[server.Name]; exists {
		return fmt.Errorf("server %s already running", server.Name)
	}

	m.servers[server.Name] = server

	return nil
}

func (m *Manager) List() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	result := make([]string, 0, len(m.servers))

	for _, server := range m.servers {
		result = append(result, server.Name)
	}

	return result
}

func (m *Manager) Start(name string) error {
	cmd, err := launcher.Build(name)
	if err != nil {
		return err
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	server := &Server{
		Name:    name,
		Cmd:     cmd,
		PTY:     ptmx,
		Log:     ringbuffer.New(1000),
		Clients: make(map[*ScreenClient]struct{}),
	}

	go func() {

		reader := bufio.NewReader(server.PTY)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			server.Broadcast(strings.TrimRight(line, "\r\n"))
		}

	}()

	if err := m.Add(server); err != nil {
		_ = cmd.Process.Kill()
		return err
	}

	fmt.Printf("Started %s (PID %d)\n", name, cmd.Process.Pid)

	go func() {
		err := cmd.Wait()

		m.Remove(name)

		if err != nil {
			fmt.Printf("%s exited: %v\n", name, err)
		} else {
			fmt.Printf("%s exited normally\n", name)
		}
	}()

	return nil
}

func (m *Manager) Remove(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.servers, name)
}

func (m *Manager) Get(name string) (*Server, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	server, ok := m.servers[name]
	return server, ok
}

func NewScreenClient(conn net.Conn) *ScreenClient {
	return &ScreenClient{Conn: conn}
}

func (s *Server) readFromClient(c *ScreenClient) {
	defer func() {
		s.mu.Lock()
		delete(s.Clients, c)
		s.mu.Unlock()
		_ = c.Conn.Close()
	}()

	buf := make([]byte, 4096)

	for {
		n, err := c.Conn.Read(buf)
		if err != nil {
			return
		}

		if _, err := s.PTY.Write(buf[:n]); err != nil {
			return
		}
	}
}

func (s *Server) Attach(c *ScreenClient) error {
	fmt.Println("Attach called")

	s.mu.Lock()
	s.Clients[c] = struct{}{}
	s.mu.Unlock()

	for _, line := range s.Log.Snapshot() {
		fmt.Println("sending:", line)
		if _, err := fmt.Fprintf(c.Conn, "%s\r\n", line); err != nil {
			fmt.Println("send failed:", err)
			return err
		}
	}

	fmt.Println("starting client reader")
	go s.readFromClient(c)

	return nil
}

func (s *Server) Broadcast(line string) {
	s.Log.Add(line)

	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.Clients {
		if _, err := fmt.Fprintf(client.Conn, "%s\r\n", line); err != nil {
			_ = client.Conn.Close()
			delete(s.Clients, client)
		}
	}
}

func (s *Server) readInput(c *ScreenClient) {
	scanner := bufio.NewScanner(c.Conn)

	for scanner.Scan() {
		fmt.Fprintln(s.PTY, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}

	delete(s.Clients, c)

	c.Conn.Close()
}
