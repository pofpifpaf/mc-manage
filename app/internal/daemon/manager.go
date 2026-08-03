package daemon

import (
	"errors"
	"fmt"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/launcher"
	"minecraft-manager/internal/protocol"
	"minecraft-manager/internal/ringbuffer"
	"minecraft-manager/internal/ui"
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
	Port              string
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

func (m *Manager) ListRunning() []protocol.ServerInfo {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	result := make([]protocol.ServerInfo, 0, len(m.servers))

	for _, server := range m.servers {

		serv, _ := MakeServerInfo(server)

		result = append(result, serv)
	}

	return result
}

func MakeServerInfo(server *Server) (protocol.ServerInfo, error) {

	cfg, err := config.Load(server.Name)
	if err != nil {
		return protocol.ServerInfo{}, err
	}

	serv := protocol.ServerInfo{
		Name:              server.Name,
		Port:              server.Port,
		AutomaticRestarts: server.AutomaticRestarts,
		StartedAt:         server.StartedAt,
		Running:           true,
		Version:           cfg.Version,
		JavaVersion:       cfg.Java,
	}

	return serv, nil
}

func (m *Manager) Stop(name string) error {
	server, serverExists := m.Get(name)

	if !serverExists {
		return errors.New("server is not running")
	}

	ui.PrintInfo("Stopping server " + name)

	server.AutomaticRestarts = false
	server.PTY.Write([]byte("\nstop\n"))

	return nil
}

func (m *Manager) Kill(server string) error {
	serv, exist := m.Get(server)
	if !exist {
		return fmt.Errorf("server is not running")
	}

	pid := serv.Cmd.Process.Pid

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("Error finding process: %s", err)

	}

	err = process.Kill()
	if err != nil {
		return fmt.Errorf("Error killing process: %s", err)
	}

	return nil
}

func (m *Manager) Start(name string) error {
	cmd, autorestarts, port, err := launcher.Build(name)
	if err != nil {
		return err
	}

	portAlreadyUsed := false
	for _, server := range m.servers {
		portAlreadyUsed = portAlreadyUsed || (port == server.Port)
	}
	if portAlreadyUsed {
		return fmt.Errorf("Port %s already used", port)
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
		Port:              port,
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

	ui.PrintSuccess(fmt.Sprintf("%s started (PID %d)", name, cmd.Process.Pid))

	go func() {

		err := cmd.Wait()

		if err != nil {
			ui.PrintError(fmt.Sprintf("%s exited: %v\n", name, err))
			send := fmt.Sprintf("\n----------- Server exited with error code %v, screen session can now be detached...\n", err)
			server.Broadcast([]byte(send))
		} else {
			ui.PrintInfo(name + " exited normally")
			server.Broadcast([]byte("\n----------- Server exited normally, screen session can now be detached...\n"))
		}

		m.Remove(name)

		if server.AutomaticRestarts {
			time.Sleep(5 * time.Second)
			ui.PrintInfo(server.Name + " exited, automatic restart is turned on, restarting...")
			if err := m.Start(name); err != nil {
				ui.PrintError(err.Error())
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
	case "autorestart":

		cfg, err := config.Load(server)
		if err != nil {
			return err
		}

		switch data.(string) {
		case "false":
			cfg.AutomaticRestarts = false
			if m.servers[server] != nil {
				m.servers[server].AutomaticRestarts = false
			}
		case "true":
			cfg.AutomaticRestarts = true
			if m.servers[server] != nil {
				m.servers[server].AutomaticRestarts = true
			}
		default:
			return fmt.Errorf("incompatible data")
		}

		err = config.Save(server, cfg)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("Unknown parameter for SET %s", paramType)

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
			ui.PrintError("send failed: " + err.Error())
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
			ui.PrintInfo(fmt.Sprintf("received: %q", buf[:n]))
			s.PTY.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
}
