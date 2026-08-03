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
	Java               string   `json:"java"`
	Memory             string   `json:"memory"`
	Jar                string   `json:"jar"`
	Port               string   `json:"port"`
	LevelName          string   `json:"level"`
	AutomaticRestarts  bool     `json:"autorestart"`
	StartOnBoot        bool     `json:"boot"`
	AdditionalJVMArgs  []string `json:"additionaljvmargs"`
	AdditionalServArgs []string `json:"additionalservargs"`
}

type ServerInfo struct {
	Name              string
	Port              string
	AutomaticRestarts bool
	StartedAt         time.Time
	Running           bool
	Version           string
	JavaVersion       string
	StartOnBoot       bool
	PlayersOnline     int
	PlayersOnlineMax  int
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
