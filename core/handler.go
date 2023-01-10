package core

import (
	"encoding/hex"
	"fmt"
	"sync"

	"go.bug.st/serial"
)

type Handler struct {
	Port    serial.Port
	WG      *sync.WaitGroup
	MsgChan chan *Packet
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

func (h *Handler) Send(msg []byte) error {
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

// Listen listens for new messages from the serial device. It's meant to be used in
// a goroutine.
func (h *Handler) Listen() {
	defer h.WG.Done()

	if h.MsgChan == nil {
		fmt.Println("No channel for message communication!")
		return
	}

	for {
		msg := make([]byte, 0)
		read := 0

		for {
			buff := make([]byte, 1)
			n, err := h.Port.Read(buff)

			if err != nil {
				fmt.Println("Read Error:", err)
				continue
			}

			if n == 0 || buff[0] == 0 {
				break
			}

			msg = append(msg, buff...)
			read += 1
		}

		raw, err := hex.DecodeString(string(msg))

		if err != nil {
			fmt.Println(err)
			continue
		}

		packet, err := UnmarshalPacket(raw)

		if err != nil {
			fmt.Println(err)
			continue
		} else if packet.Checksum != packet.GetChecksum() {
			fmt.Println("Checksums didn't match!")
			continue
		}

		h.MsgChan <- packet
	}
}
