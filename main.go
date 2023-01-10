package main

import (
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

	// Encrypt message.
	// salt, _ := crypto.CreateSalt(32)
	salt := []byte{74, 112, 125, 227, 245, 226, 113, 225, 219, 162, 165, 100, 52, 233, 89, 50, 77, 215, 48, 155, 219, 131, 51, 192, 130, 124, 191, 199, 240, 179, 175, 43}
	key, _ := crypto.PBKDF2Key([]byte("test key"), salt)
	ciphertext, err := crypto.EncryptGCM([]byte("Test message"), key)

	if err != nil {
		log.Fatal(err)
	}

	handler.Send(ciphertext)

	wg = new(sync.WaitGroup)
	wg.Add(1)

	go handler.Listen()
	go func(msgChan chan *core.Packet) {
		for {
			p, ok := <-msgChan

			if !ok {
				fmt.Println("Error reading from chan")
				continue
			}

			key, _ := crypto.PBKDF2Key([]byte("test key"), salt)
			plaintext, err := crypto.DecryptGCM(p.Content, key)
			fmt.Println(plaintext, p.Checksum, err)
		}
	}(handler.MsgChan)

	wg.Wait()
}
