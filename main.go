package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/wisepythagoras/owncom/core"
	"github.com/wisepythagoras/owncom/crypto"
	"go.bug.st/serial"
)

var wg *sync.WaitGroup

func main() {
	device := flag.String("device", "", "The serial device to open")
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

	msgChan := make(chan *core.Packet)
	handler := core.Handler{WG: wg, MsgChan: msgChan}

	if err = handler.ConnectToSerial(*device); err != nil {
		log.Fatal(err)
	}

	// Encrypt message. Normally it would be called as follows:
	// salt, _ := crypto.CreateSalt(32)
	salt, _ := hex.DecodeString("4a707de3f5e271e1dba2a56434e959324dd7309bdb8333c0827cbfc7f0b3af2b")
	msg := core.Message{Msg: []byte("Test message")}
	packets, err := msg.PacketsAESGCM([]byte("test key"), salt)

	if err != nil {
		log.Fatal(err)
	}

	handler.Send(packets)

	wg = new(sync.WaitGroup)
	wg.Add(1)

	go handler.Listen()
	go func(msgChan chan *core.Packet) {
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

				fmt.Printf("%s (%d)> %s (%d packets)", p.ID, p.Checksum, string(plaintext), p.Total)
			}
		}
	}(handler.MsgChan)

	wg.Wait()
}
