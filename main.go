package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/wisepythagoras/owncom/core"
	"go.bug.st/serial"
)

var wg *sync.WaitGroup

func read(portPtr *serial.Port) {
	defer wg.Done()

	port := *portPtr

	for {
		msg := make([]byte, 0)
		read := 0

		for {
			buff := make([]byte, 1)
			n, err := port.Read(buff)

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

		// fmt.Println(hex.EncodeToString(msg))

		raw, err := hex.DecodeString(string(msg))

		if err != nil {
			fmt.Println(err)
			continue
		}

		packet, err := core.UnmarshalPacket(raw)

		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println(packet.Content)
	}
}

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

	mode := &serial.Mode{
		BaudRate: 9600,
	}
	port, err := serial.Open(*device, mode)

	if err != nil {
		log.Fatal(err)
	}

	packet := &core.Packet{Content: "Test message"}
	hexMsg, err := packet.MarshalToHex()
	fmt.Println(hexMsg)
	msg := []byte(hexMsg)

	if err != nil {
		log.Fatalln(err)
	}

	// msg := []byte("Test message")
	// msg = append(msg, 0)
	msg = append(msg, 0)
	port.Write(msg)

	wg = new(sync.WaitGroup)
	wg.Add(1)

	go read(&port)

	wg.Wait()
}
