package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wisepythagoras/owncom/core"
	"github.com/wisepythagoras/owncom/crypto"
	"go.bug.st/serial"
)

var wg *sync.WaitGroup

func main() {
	device := flag.String("device", "", "The serial device to open")
	baudRate := flag.Int("baud-rate", 9600, "The baud rate")
	modulePath := flag.String("module", "", "Pass a module")
	flag.Parse()

	if len(*device) == 0 {
		log.Fatalln("No serial device was selected")
	}

	ports, err := serial.GetPortsList()

	if err != nil {
		log.Fatalln(err)
	}

	if len(ports) == 0 {
		log.Fatalln("No serial ports found")
	}

	var module *core.Module

	if *modulePath != "" {
		module, err = core.ReadModule(*modulePath)

		if err != nil {
			log.Fatalln(err)
		}
	}

	found := false

	for _, port := range ports {
		fmt.Println("Found port", port)

		if port == *device {
			found = true
		}
	}

	if !found {
		log.Fatalf("No such serial device: %q\n", *device)
	}

	// Encrypt message. Normally it would be called as follows:
	// salt, _ := crypto.CreateSalt(32)
	salt, _ := hex.DecodeString("4a707de3f5e271e1dba2a56434e959324dd7309bdb8333c0827cbfc7f0b3af2b")
	aesGcmKey := crypto.AESGCMKey{
		Key:  []byte("test key"),
		Salt: salt,
	}

	msgChan := make(chan *core.Packet)
	outputChan := make(chan string)
	handler := core.Handler{
		WG:      wg,
		MsgChan: msgChan,
		OutChan: outputChan,
		Module:  module,
	}
	program := tea.NewProgram(createModel(&handler, &aesGcmKey), tea.WithAltScreen())

	if err = handler.ConnectToSerial(*device, *baudRate); err != nil {
		log.Fatal(err)
	}

	wg = new(sync.WaitGroup)
	wg.Add(1)

	if module != nil {
		go handler.ListenRaw(false)
	} else {
		// The default behavior is to treat the serial device as if it accepts any data and it pipes
		// it to the transmitter as-is.
		go handler.Listen()
	}

	// go func(outChan chan string, program *tea.Program) {
	// 	for {
	// 		msg, ok := <-outChan

	// 		if !ok {
	// 			fmt.Println("Error reading from output chan")
	// 			continue
	// 		}

	// 		program.Send(UserMessage{
	// 			From:    "SYS: ",
	// 			Message: msg,
	// 		})
	// 	}
	// }(outputChan, program)

	go func(msgChan chan *core.Packet, program *tea.Program) {
		countMap := make(map[string]uint32)
		packetMap := make(map[string][]*core.Packet)

		for {
			p, ok := <-msgChan

			if !ok {
				fmt.Println("Error reading from chan")
				continue
			}

			var count uint32

			if count, ok = countMap[p.ID]; !ok {
				countMap[p.ID] = 0
				packetMap[p.ID] = make([]*core.Packet, 0)
				count = 0
			}

			packetMap[p.ID] = append(packetMap[p.ID], p)
			countMap[p.ID] += 1
			count += 1

			if count == p.Total {
				data := make([]byte, 0)

				for _, packet := range packetMap[p.ID] {
					data = append(data, packet.Content...)
				}

				key, _ := crypto.PBKDF2Key([]byte("test key"), salt)
				plaintext, err := crypto.DecryptGCM(data, key)

				delete(countMap, p.ID)
				delete(packetMap, p.ID)

				if err != nil {
					fmt.Println(err)
					continue
				}

				// fmt.Printf("%s (%d)> %s (%d packets)\n", p.ID, p.Checksum, string(plaintext), p.Total)
				program.Send(UserMessage{
					From:    "Someone: ",
					Message: string(plaintext),
				})
			}
		}
	}(handler.MsgChan, program)

	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}

	wg.Wait()
}
