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
	Module  *Module
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

		var wg *sync.WaitGroup

		if h.Module != nil {
			temp := h.Module.Marshal(msgBytes)
			msgBytes = temp

			wg = new(sync.WaitGroup)
			wg.Add(1)

			go h.getOne()
		}

		numOfSegments := len(msgBytes) / 50
		remainderBytes := len(msgBytes) % 50

		for i := 0; i < numOfSegments; i++ {
			if _, err = h.Port.Write(msgBytes[50*i : 50*(i+1)]); err != nil {
				return err
			}
		}

		if remainderBytes > 0 {
			if _, err = h.Port.Write(msgBytes[len(msgBytes)-remainderBytes:]); err != nil {
				return err
			}
		}

		if h.Module != nil {
			wg.Wait()
		}
	}

	return nil
}

func (h *Handler) getOne() ([]byte, error) {
	msg := make([]byte, 0)

	for {
		buff := make([]byte, 1)
		_, err := h.Port.Read(buff)

		if err != nil {
			return nil, err
		}

		msg = append(msg, buff...)
		l := len(msg)
		fmt.Println(buff)

		if l > 2 && msg[l-2] == 13 && msg[l-1] == 10 {
			break
		}
	}

	return msg, nil
}

func (h *Handler) ListenRaw(onlyOne bool) {
	defer h.WG.Done()

	for {
		msg, err := h.getOne()

		if err != nil {
			fmt.Println(err)

			if onlyOne {
				break
			}

			continue
		}

		fmt.Println(string(msg))

		if onlyOne {
			break
		}
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
