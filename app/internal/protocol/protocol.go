package protocol

import "time"

type Request struct {
	Command string `json:"command"`

	Server string `json:"server,omitempty"`

	Version string `json:"version,omitempty"`

	Type string `json:"type,omitempty"`

	Text string `json:"text,omitempty"`

	Data interface{} `json:"data,omitempty"`
}

type Response struct {
	OK bool `json:"ok"`

	Message string `json:"message,omitempty"`

	Data interface{} `json:"data,omitempty"`
}

type Config struct {
	Name               string   `json:"name"`
	Type               string   `json:"type"`
	Version            string   `json:"version"`
	VersionArg         string   `json:"versionArg"`
	Java               string   `json:"java"`
	MemoryAllocated    string   `json:"memoryAllocated"`
	MemoryMax          string   `json:"memoryMax"`
	Jar                string   `json:"jar"`
	Port               string   `json:"port"`
	LevelName          string   `json:"level"`
	AutomaticRestarts  bool     `json:"autorestart"`
	StartOnBoot        bool     `json:"boot"`
	Username           string   `json:"username"`
	Uid                int      `json:"uid"`
	Gid                int      `json:"gid"`
	AdditionalJVMArgs  []string `json:"additionaljvmargs"`
	AdditionalServArgs []string `json:"additionalservargs"`
}

type MainConfig struct {
	ServerFilePath     string `json:"server-files-path"`
	GracePeriodSeconds int    `json:"grace-period-seconds"`
}

type ServerState string

const (
	StateStopSent ServerState = "stopping"
	StateStopped  ServerState = "stopped"
	StateStarting ServerState = "starting"
	StateRunning  ServerState = "running"
)

type ServerInfo struct {
	Name              string
	Type              string
	Port              string
	AutomaticRestarts bool
	StartedAt         time.Time
	Running           ServerState
	Version           string
	JavaVersion       string
	StartOnBoot       bool
	PlayersOnline     int
	PlayersOnlineMax  int
	MemoryUsed        int64
	Username          string
	Uid               int
	Gid               int
}

type ServerPSResponse struct {
	OK      bool         `json:"ok"`
	Message string       `json:"message,omitempty"`
	Data    []ServerInfo `json:"data"`
}

const (
	setParameterPort        = "port"
	setParameterAutoRestart = "autorestart"
)

type PaperVersionManifest struct {
	Version struct {
		Java struct {
			Version struct {
				Minimum int `json:"minimum"`
			} `json:"version"`
			Flags struct {
				Recommended []string `json:"recommended"`
			} `json:"flags"`
		} `json:"java"`
	} `json:"version"`
}
