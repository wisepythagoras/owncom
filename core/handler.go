package core

import (
	"go.bug.st/serial"
)

type Handler struct {
	Port serial.Port
}

func (h *Handler) ConnectToSerial(device string) error {
	mode := &serial.Mode{
		BaudRate: 9600,
	}
	port, err := serial.Open(device, mode)

	if err == nil {
		h.Port = port
	}

	return err
}

func (h *Handler) Send(msg string) error {
	packet := &Packet{Content: msg}
	hexMsg, err := packet.MarshalToHex()
	msgBytes := []byte(hexMsg)

	if err != nil {
		return err
	}

	msgBytes = append(msgBytes, 0)
	h.Port.Write(msgBytes)

	return nil
}
