package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var UdpPort = os.Getenv("NOMAD_HOST_PORT_udp")

func SendUDP(daters string, cell *Cell) (err error) {
	for i := 0; i < 5; i++ {
		err = SendUDPOnce(daters, cell)
		if err == nil {
			break
		}
	}
	return
}

func SendUDPOnce(daters string, cell *Cell) (err error) {
	addr, err := cell.Address()
	if err != nil {
		return
	}

	start := time.Now()

	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Printf("Error getting UDP port for %s: %s", cell.Name(), err)
		return
	}
	defer conn.Close()

	// remember: this fixed my "jobs dont die becuause frozen" bug
	now := time.Now()
	err = conn.SetDeadline(now.Add(150 * time.Millisecond)) // TODO: longer when run on a cluster?
	if err != nil {
		log.Println("Error setting deadline:", err)
	}

	fmt.Println("Sending", daters, "to", cell.Name())
	_, err = conn.Write([]byte(daters))
	if err != nil {
		log.Printf("Error sending %q to %s: %s", daters, cell.Name(), err)
	}

	_, _, err = bufio.NewReader(conn).ReadLine()
	if err != nil {
		log.Printf("Error reading response from %s: %s", cell.Name(), err)
		return
	}

	end := time.Now()
	log.Printf("SendUDP %q to %q duration: %s", daters, cell.Name(), end.Sub(start))

	return
}
