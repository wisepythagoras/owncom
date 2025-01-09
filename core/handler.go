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
	OutChan chan string
	Module  *Module
	lockWg  *sync.WaitGroup
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

		numOfSegments := len(msgBytes) / 50
		remainderBytes := len(msgBytes) % 50

		for i := 0; i < numOfSegments; i++ {
			m := msgBytes[50*i : 50*(i+1)]

			if h.Module != nil {
				m = h.Module.Marshal(m)
			}

			h.Port.Drain()

			if err = h.SendRaw(m); err != nil {
				return err
			}

			if h.Module != nil {
				msg, ok := <-h.OutChan

				if !ok {
					return fmt.Errorf("failed to read")
				} else if msg != "+OK\r\n" {
					return fmt.Errorf("unknown error %q", string(m))
				}
			}
		}

		if remainderBytes > 0 {
			m := msgBytes[len(msgBytes)-remainderBytes:]

			if h.Module != nil {
				m = h.Module.Marshal(m)
			}

			h.Port.Drain()

			if err = h.SendRaw(m); err != nil {
				return err
			}

			if h.Module != nil {
				msg, ok := <-h.OutChan

				if !ok {
					return fmt.Errorf("failed to read")
				} else if msg != "+OK\r\n" {
					return fmt.Errorf("unknown error %q", string(m))
				}
			}
		}
	}

	return nil
}

// GetOne gets one full message from the device until it encounters "\r\n".
func (h *Handler) GetOne() ([]byte, error) {
	msg := make([]byte, 0)

	for {
		buff := make([]byte, 1)
		_, err := h.Port.Read(buff)

		if err != nil {
			return nil, err
		}

		msg = append(msg, buff...)
		l := len(msg)

		if l >= 2 && msg[l-2] == 13 && msg[l-1] == 10 {
			break
		}
	}

	return msg, nil
}

func (h *Handler) ListenRaw(onlyOne bool) {
	defer h.WG.Done()

	if h.lockWg == nil {
		h.lockWg = new(sync.WaitGroup)
	}

	for {
		// Wait in case something else is trying to read from the device.
		h.lockWg.Wait()

		msg, err := h.GetOne()

		if err != nil {
			fmt.Println(err)

			if onlyOne {
				break
			}

			continue
		}

		h.OutChan <- string(msg)

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
