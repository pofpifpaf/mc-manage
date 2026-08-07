package daemon

import (
	"bufio"
	"fmt"
	"minecraft-manager/internal/config"
	"minecraft-manager/internal/launcher"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"minecraft-manager/internal/ringbuffer"
	"minecraft-manager/internal/ui"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
	config  *protocol.MainConfig
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

	memoryUsed, err := server.getMemoryUsed()
	if err != nil {
		ui.PrintError(fmt.Sprintf("Could not get memory used for server %s, err: %s", server.Name, err.Error()))
	}

	serv := protocol.ServerInfo{
		Name:              server.Name,
		Port:              server.Port,
		AutomaticRestarts: server.AutomaticRestarts,
		StartedAt:         server.StartedAt,
		Running:           protocol.StateRunning,
		Version:           cfg.Version,
		JavaVersion:       cfg.Java,
		MemoryUsed:        memoryUsed,
	}

	return serv, nil
}

func (m *Manager) ServerCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return len(m.servers)
}

func (m *Manager) StopAllServers() {
	m.mutex.Lock()
	servers := m.servers
	m.mutex.Unlock()

	for _, server := range servers {
		err := m.Stop(server.Name)
		if err != nil {
			ui.PrintError("could not send stop command to server " + server.Name + " | error : " + err.Error())
		}
	}
}

func (m *Manager) KillAllServers() {
	m.mutex.Lock()
	servers := m.servers
	m.mutex.Unlock()

	for _, server := range servers {
		err := m.Kill(server.Name)
		if err != nil {
			ui.PrintError("could not send kill command to server " + server.Name + " | error : " + err.Error())
		}
	}
}

func (m *Manager) Stop(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	server, serverExists := m.Get(name)

	if !serverExists {
		return fmt.Errorf("Server %s is not running", name)
	}

	ui.PrintInfo("Stopping server " + name)

	server.AutomaticRestarts = false
	_, err := server.PTY.Write([]byte("\nstop\n"))

	return err
}

func (m *Manager) Kill(server string) error {

	m.mutex.Lock()
	defer m.mutex.Unlock()

	ui.PrintInfo(fmt.Sprintf("Sending kill command to server %s", server))

	serv, exist := m.Get(server)
	if !exist {
		return fmt.Errorf("server is not running")
	}

	pid := serv.Cmd.Process.Pid

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("Error finding process: %s", err)

	}

	serv.AutomaticRestarts = false
	err = process.Kill()
	if err != nil {
		return fmt.Errorf("Error killing process: %s", err)
	}

	return nil
}

func (m *Manager) Start(name string) error {

	if m.servers[name] != nil {
		return fmt.Errorf("server %s is already running", name)
	}

	cmd, autorestarts, port, err := launcher.Build(name)
	if err != nil {
		return err
	}

	portAlreadyUsed := false
	for _, server := range m.servers {
		portAlreadyUsed = portAlreadyUsed || (port == server.Port)
	}
	if portAlreadyUsed {
		return fmt.Errorf("Port %s already used", port) // See if downgrade to warning
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
			ui.PrintError(fmt.Sprintf("%s exited: %v", name, err))
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

	server, ok := m.servers[name]
	return server, ok
}

func (m *Manager) SetParameter(server string, paramType string, data any) error {

	m.mutex.Lock()
	defer m.mutex.Unlock()

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

	case "graceperiod":

		if err := m.setGracePeriodCfg(data); err != nil {
			return err
		}

		var periodDuration time.Duration
		periodDuration = time.Duration(m.config.GracePeriodSeconds) * time.Second

		ui.PrintInfo(fmt.Sprintf("Grace Period is now set to %s", periodDuration))

	default:
		return fmt.Errorf("Unknown parameter for SET %s", paramType)

	}
	return nil
}

func (m *Manager) setGracePeriodCfg(data any) error {
	gracePeriod := int(data.(float64))

	cfg, err := config.LoadMainConfig()
	if err != nil {
		return err
	}

	cfg.GracePeriodSeconds = gracePeriod

	m.config = cfg

	return config.SaveMainConfig(cfg)

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

func (s *Server) getMemoryUsed() (int64, error) {
	pidStatusPath := paths.PidStatus(s.Cmd.Process.Pid)

	f, err := os.Open(pidStatusPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Pss:") {

			fields := strings.Fields(line)
			if len(fields) < 2 {
				return 0, fmt.Errorf("malformed Pss line: %q", line)
			}

			pssKB, err := strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				return 0, err
			}

			return pssKB * 1000, nil // KB -> B
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return 0, fmt.Errorf("Pss not found")
}
