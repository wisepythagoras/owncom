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

func (h *Handler) ConnectToSerial(device string, baudRate int) error {
	mode := &serial.Mode{
		BaudRate: baudRate,
	}
	port, err := serial.Open(device, mode)

	if err == nil {
		h.Port = port
	}

	return err
}

func (h *Handler) SendRaw(data []byte) error {
	_, err := h.Port.Write(data)

	return err
}

func (h *Handler) Send(packets []Packet) error {
	for _, packet := range packets {
		var msgHex string
		var err error

		if msgHex, err = packet.MarshalToHex(); err != nil {
			return err
		}

		msgBytes := []byte(msgHex)
		msgBytes = append(msgBytes, 0)

		if _, err = h.Port.Write(msgBytes); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) readBytes() ([]byte, int) {
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

	return msg, read
}

func (h *Handler) ListenRaw() {
	defer h.WG.Done()

	for {
		buff := make([]byte, 1)
		_, err := h.Port.Read(buff)

		if err != nil {
			fmt.Println("Read Error:", err)
			continue
		}

		fmt.Print(string(buff))
	}
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
