package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

// import (
// 	"bufio"
// 	"context"
// 	"fmt"
// 	"net"
// 	"os"
// 	"strings"
// )

var UdpPort = os.Getenv("NOMAD_HOST_PORT_udp")

// // func main() {
// // 	if port == "" {
// // 		port = "101"
// // 	}
// // 	u := UDPServer{
// // 		ctx:  context.Background(),
// // 		addr: "0.0.0.0:" + port,
// // 	}
// // 	u.serve()
// // }

// func NewUDPServer(cell *Cell) UDPServer {
// 	if UdpPort == "" {
// 		UdpPort = "101"
// 	}
// 	return UDPServer{
// 		ctx:  context.Background(),
// 		addr: "0.0.0.0:" + UdpPort,
// 	}
// }

// type UDPServer struct {
// 	ctx  context.Context
// 	addr string
// 	cell *Cell
// }

// func (u UDPServer) serve() (err error) {
// 	conn, err := net.ListenPacket("udp", u.addr)
// 	if err != nil {
// 		return
// 	}
// 	defer conn.Close()

// 	errChan := make(chan error, 1)
// 	buf := make([]byte, 1024)
// 	go func() {
// 		for {
// 			n, dst, err := conn.ReadFrom(buf)
// 			if err != nil {
// 				errChan <- err
// 			}
// 			if buf[:n] == nil {
// 				continue
// 			}
// 			fmt.Printf("serv recv %s", string(buf[:n]))
// 			parts := strings.Split(string(buf[:n]), " ")
// 			neighbor := parts[0]
// 			alive := parts[1] == "true"
// 			u.cell.n[neighbor].status = alive
// 			conn.WriteTo(buf, dst)
// 		}
// 	}()
// 	fmt.Println("server started")

// 	select {
// 	case <-u.ctx.Done():
// 		fmt.Println("cancelled")
// 		err = u.ctx.Err()
// 	case err = <-errChan:
// 	}
// 	return
// }

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
	// addr = strings.Replace(addr, "127.0.0.1", "host.docker.internal", -1) // TODO: undo this osx kludge.

	// start := time.Now()

	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Printf("Error getting UDP port for %s: %s", cell.Name(), err)
		return
	}
	defer conn.Close()

	// TODO: remember: this fixed my "jobs dont die becuause frozen" bug
	now := time.Now()
	err = conn.SetDeadline(now.Add(100 * time.Millisecond)) // TODO: longer when run on a cluster?
	if err != nil {
		log.Println("Error setting deadline:", err)
	}

	fmt.Println("Sending", daters, "to", cell.Name())
	_, err = conn.Write([]byte(daters))
	if err != nil {
		log.Printf("Error sending %q to %s: %s", daters, cell.Name(), err)
	}

	// buf, _, err := bufio.NewReader(conn).ReadLine()
	_, _, err = bufio.NewReader(conn).ReadLine()
	if err != nil {
		log.Printf("Error reading response from %s: %s", cell.Name(), err)
		return
	}
	// log.Println("clnt recv", string(buf))

	// end := time.Now()
	// log.Printf("SendUDP %q to %q duration: %s", daters, cell.Name(), end.Sub(start))

	return
}
