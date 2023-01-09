package core

import (
	"go.bug.st/serial"
)

type Handler struct {
	Port serial.Port
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
