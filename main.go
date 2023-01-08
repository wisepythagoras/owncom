package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"go.bug.st/serial"
)

var wg *sync.WaitGroup

func read(portPtr *serial.Port) {
	defer wg.Done()

	port := *portPtr

	for {
		msg := make([]byte, 256)

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
		}

		fmt.Println(string(msg))
	}
}

func main() {
	device := flag.String("device", "", "The serial device to open")
	flag.Parse()

	if len(*device) == 0 {
		log.Fatal("No serial device was selected")
	}

	ports, err := serial.GetPortsList()

	if err != nil {
		log.Fatal(err)
	}

	if len(ports) == 0 {
		log.Fatal("No serial ports found")
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

	msg := []byte("Test message")
	msg = append(msg, 0)
	port.Write(msg)

	wg = new(sync.WaitGroup)
	wg.Add(1)

	go read(&port)

	wg.Wait()
}
