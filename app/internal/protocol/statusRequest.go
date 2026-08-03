package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

type Status struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`

	Players struct {
		Max    int `json:"max"`
		Online int `json:"online"`
	} `json:"players"`

	Description any `json:"description"`
}

func writeVarInt(w io.Writer, value int32) error {
	uv := uint32(value)
	for {
		if uv&^0x7F == 0 {
			_, err := w.Write([]byte{byte(uv)})
			return err
		}

		if _, err := w.Write([]byte{byte(uv&0x7F | 0x80)}); err != nil {
			return err
		}

		uv >>= 7
	}
}

func readVarInt(r io.Reader) (int32, error) {
	var num int32
	var shift uint

	for {
		var b [1]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, err
		}

		num |= int32(b[0]&0x7F) << shift

		if b[0]&0x80 == 0 {
			break
		}

		shift += 7

		if shift > 35 {
			return 0, fmt.Errorf("VarInt too big")
		}
	}

	return num, nil
}

func writeString(w io.Writer, s string) error {
	if err := writeVarInt(w, int32(len(s))); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}

func GetServerStatus(port string) (Status, error) {

	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
	if err != nil {
		return Status{}, err
	}
	defer conn.Close()

	var payload bytes.Buffer

	writeVarInt(&payload, 0)   // Packet ID
	writeVarInt(&payload, 772) // Protocol version - to see if there is a need to change this.
	writeString(&payload, "localhost")
	binary.Write(&payload, binary.BigEndian, uint16(25565))
	writeVarInt(&payload, 1) // Status state

	writeVarInt(conn, int32(payload.Len()))
	conn.Write(payload.Bytes())

	var req bytes.Buffer

	writeVarInt(&req, 0)

	writeVarInt(conn, int32(req.Len()))
	conn.Write(req.Bytes())

	length, _ := readVarInt(conn)
	_ = length // total packet length

	if length < 15 {
		return Status{}, fmt.Errorf("json response too short")
	}

	packetID, _ := readVarInt(conn)
	if packetID != 0 {
		return Status{}, fmt.Errorf("unexpected packet")
	}

	strLen, _ := readVarInt(conn)

	jsonBytes := make([]byte, strLen)
	io.ReadFull(conn, jsonBytes)

	var status Status
	if err := json.Unmarshal(jsonBytes, &status); err != nil {
		return Status{}, err
	}

	return status, nil
}
