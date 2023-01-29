package core

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Module struct {
	Send       string `json:"send"`
	Receive    string `json:"receive"`
	IsRegex    bool   `json:"is_regex"`
	PacketSize int    `json:"packet_size"`
}

func (m *Module) Marshal(data []byte) []byte {
	sendStr := m.Send
	lenStr := fmt.Sprintf("%d", len(string(data))-1)

	sendStr = strings.ReplaceAll(sendStr, "{LEN}", lenStr)
	sendStr = strings.ReplaceAll(sendStr, "{DATA}", string(data))

	return []byte(sendStr)
}

func ReadModule(path string) (*Module, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	newModule := new(Module)

	if err := json.Unmarshal(data, newModule); err != nil {
		return nil, err
	}

	return newModule, nil
}
