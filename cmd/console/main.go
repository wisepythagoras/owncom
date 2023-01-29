package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/wisepythagoras/owncom/core"
	"go.bug.st/serial"
)

var wg *sync.WaitGroup

func main() {
	device := flag.String("device", "", "The serial device to open")
	baudRate := flag.Int("baud-rate", 9600, "The baud rate")
	data := flag.String("data", "", "The raw data to send")
	flag.Parse()

	if len(*device) == 0 {
		log.Fatalln("No serial device was selected")
	} else if len(*data) == 0 {
		log.Fatalln("No data passed")
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

	handler := core.Handler{WG: wg}

	if err = handler.ConnectToSerial(*device, *baudRate); err != nil {
		log.Fatal(err)
	}

	handler.SendRaw([]byte(*data + "\r\n"))
	msg, err := handler.GetOne()

	if err != nil {
		log.Fatalln(err)
	}

	stringResp := string(msg)
	stringResp = strings.Trim(stringResp, "\r\n")

	fmt.Println(stringResp)
}
