package daemon

import (
	"errors"
	"fmt"
	"minecraft-manager/internal/client"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/launcher"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/ringbuffer"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
)

type Server struct {
	Name              string
	Cmd               *exec.Cmd
	PTY               *os.File
	Log               *ringbuffer.RingBuffer
	mu                sync.Mutex
	Clients           map[*ScreenClient]struct{}
	AutomaticRestarts bool
	StartedAt         time.Time
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

func (m *Manager) List() []client.ServerInfo {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	result := make([]client.ServerInfo, 0, len(m.servers))

	for _, server := range m.servers {

		port, _ := config.GetServerProperty(paths.ServerProperties(server.Name), "server-port")

		serv := client.ServerInfo{
			Name:              server.Name,
			Port:              port,
			AutomaticRestarts: server.AutomaticRestarts,
		}
		result = append(result, serv)
	}

	return result
}

func (m *Manager) Stop(name string) error {
	server, serverExists := m.Get(name)

	if !serverExists {
		return errors.New("server is not running")
	}

	fmt.Printf("Stopping server %s\n", name)

	server.AutomaticRestarts = false
	server.PTY.Write([]byte("\nstop\n"))

	return nil
}

func (m *Manager) Start(name string) error {
	cmd, autorestarts, err := launcher.Build(name)
	if err != nil {
		return err
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	server := &Server{
		Name:              name,
		Cmd:               cmd,
		PTY:               ptmx,
		Log:               ringbuffer.New(1000),
		Clients:           make(map[*ScreenClient]struct{}),
		AutomaticRestarts: autorestarts,
		StartedAt:         time.Now(),
	}

	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := server.PTY.Read(buf)
			if n > 0 {
				server.Log.Add(string(buf[:n]))
				server.Broadcast(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	if err := m.Add(server); err != nil {
		_ = cmd.Process.Kill()
		return err
	}

	fmt.Printf("%s started (PID %d)\n", name, cmd.Process.Pid)

	go func() {

		err := cmd.Wait()

		if err != nil {
			fmt.Printf("%s exited: %v\n", name, err)
			send := fmt.Sprintf("\n----------- Server exited with error code %v, screen session can now be detached...\n", err)
			server.Broadcast([]byte(send))
		} else {
			fmt.Printf("%s exited normally\n", name)
			server.Broadcast([]byte("\n----------- Server exited normally, screen session can now be detached...\n"))
		}

		m.Remove(name)

		if server.AutomaticRestarts {
			time.Sleep(5 * time.Second)
			fmt.Printf("%s exited, automatic restart is turned on, restarting...\n", server.Name)
			if err := m.Start(name); err != nil {
				fmt.Println(err)
			}
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

func (m *Manager) SetParameter(server string, paramType string, data any) error {

	switch paramType {
	case "port":

		port := data.(string)

		err := config.SetServerProperty(paths.ServerProperties(server), "server-port", port)
		if err != nil {
			return err
		}

		cfg, err := config.Load(server)
		if err != nil {
			return err
		}

		cfg.Port = port

		err = config.Save(server, cfg)
		if err != nil {
			return err
		}

	case "autorestart":

		cfg, err := config.Load(server)
		if err != nil {
			return err
		}

		switch data.(string) {
		case "false":
			cfg.AutomaticRestarts = false
		case "true":
			cfg.AutomaticRestarts = true
		default:
			return fmt.Errorf("incompatible data")
		}

		err = config.Save(server, cfg)
		if err != nil {
			return err
		}
	}
	return nil
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

	s.mu.Lock()
	s.Clients[c] = struct{}{}
	s.mu.Unlock()

	for _, line := range s.Log.Snapshot() {
		if _, err := fmt.Fprintf(c.Conn, "%s", line); err != nil {
			fmt.Println("send failed:", err)
			return err
		}
	}

	go s.readFromClient(c)

	return nil
}

func (s *Server) Broadcast(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.Clients {
		if _, err := client.Conn.Write(data); err != nil {
			client.Conn.Close()
			delete(s.Clients, client)
		}
	}
}

func (s *Server) readInput(c *ScreenClient) {
	defer func() {
		delete(s.Clients, c)
		c.Conn.Close()
	}()

	buf := make([]byte, 1)
	for {
		n, err := c.Conn.Read(buf)
		if n > 0 {
			fmt.Printf("received: %q\n", buf[:n])
			s.PTY.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
}
