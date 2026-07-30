package protocol

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
