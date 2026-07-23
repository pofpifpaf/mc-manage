package daemon

import (
	"fmt"
	"sync"
	"os/exec"
	"minecraft-manager/internal/launcher"
)

type Server struct {
	Name string
	Cmd  *exec.Cmd
}

type Manager struct {
	servers map[string]*Server
	mutex   sync.Mutex
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

func (m *Manager) List() []*Server {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	result := make([]*Server, 0, len(m.servers))

	for _, server := range m.servers {
		result = append(result, server)
	}

	return result
}

func (m *Manager) Start(name string) error {
	cmd, err := launcher.Build(name)
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	server := &Server{
		Name: name,
		Cmd:  cmd,
	}

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