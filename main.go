package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/wisepythagoras/owncom/core"
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

	handler.Send("Test message")

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

			fmt.Println(p.Content, p.Checksum)
		}
	}(handler.MsgChan)

	wg.Wait()
}
