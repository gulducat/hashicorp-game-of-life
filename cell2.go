package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

var Statuses = make(map[string]bool)

type Cell2 struct {
	x     int
	y     int
	alive bool
	addr  string
	n     map[string]*Cell2

	pattern string

	cdns *ConsulDNS
	mut  sync.RWMutex
	// server *UDPServer
}

func NewCell2(name string) Cell2 {
	x, y := Coords(name)
	// udpServer := NewUDPServer2()
	return Cell2{
		x:     x,
		y:     y,
		alive: true,
		cdns:  NewConsulDNS(),
		// n:     make(map[string]*Cell2),
	}
}

func (c *Cell2) Name() string {
	return fmt.Sprintf("%d-%d", c.x, c.y)
}

func (c *Cell2) Address() (string, error) {
	// TODO: retry if stale addr; somewhere... Update()?

	// if c.addr != "" && c.Name() != "0-0" {
	if c.addr != "" {
		return c.addr, nil
	}

	addr := ""
	svc, err := Consul.Service(c.Name())
	if err != nil {
		return "", err
	}
	addr = fmt.Sprintf("%s:%d", svc.Address, svc.Port)

	// TODO: remove cdns, http is probably fine for how infrequently we'll be hitting it, and it would allow my laptop to `make more` etc
	// addr, err := c.cdns.GetServiceAddr(c.Name())
	// log.Printf("%s addr: %s\n", c.Name(), addr)
	// if err != nil {
	// 	log.Println("Err getting address: ", err)
	// 	// addr = ""
	// }

	c.addr = addr
	return addr, err
}

func (c *Cell2) Exists() bool {
	return Consul.ServiceExists(c.Name())
}

func (c *Cell2) WaitUntilExists(seconds int) bool {
	exists := false
	for x := 0; x < 10*seconds; x++ {
		if c.Exists() {
			exists = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return exists
}

func (c *Cell2) Neighbors() map[string]*Cell2 {
	if c.n != nil {
		return c.n
	}
	all := [8]*Cell2{
		// comments assuming cell "2-2"

		// top row
		&Cell2{x: c.x - 1, y: c.y - 1, cdns: c.cdns}, // 1-1
		&Cell2{x: c.x, y: c.y - 1, cdns: c.cdns},     // 2-1
		&Cell2{x: c.x + 1, y: c.y - 1, cdns: c.cdns}, // 3-1

		// middle row
		&Cell2{x: c.x - 1, y: c.y, cdns: c.cdns}, // 1-2
		// 2-2 is self.
		&Cell2{x: c.x + 1, y: c.y, cdns: c.cdns}, // 3-2

		// bottom row
		&Cell2{x: c.x - 1, y: c.y + 1, cdns: c.cdns}, // 1-3
		&Cell2{x: c.x, y: c.y + 1, cdns: c.cdns},     // 2-3
		&Cell2{x: c.x + 1, y: c.y + 1, cdns: c.cdns}, // 3-3

	}
	var valid = make(map[string]*Cell2)
	for _, n := range all {
		if n.x < 1 || n.y < 1 || n.x > MaxWidth || n.y > MaxHeight {
			continue
		}
		valid[n.Name()] = n
	}
	// c.n = valid
	if c.Name() != "0-0" { // HACK
		c.n = valid
	}
	return valid
}

func (c *Cell2) Listen() (err error) {
	addr := "0.0.0.0:" + UdpPort
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Println("Error starting server:", err)
		return
	}
	defer conn.Close()

	errChan := make(chan error, 1)
	buf := make([]byte, 512)
	go func() {
		for {
			n, dst, err := conn.ReadFrom(buf)
			// "Callers should always process the n > 0 bytes returned before considering the error err." hm.
			if err != nil {
				errChan <- err
			}
			if buf[:n] == nil {
				continue
			}
			msg := string(buf[:n])
			log.Printf("serv recv %s", msg)

			parts := strings.Split(msg, " ")

			c.pattern = ""
			c.mut.Lock()
			switch parts[0] {

			case "tick":
				continue // TODO: maybe something sends "tick tock" to all the cells, instead of the Run() loop trying to sync up with system time?

			case "random":
				rand.Seed(time.Now().UnixNano())
				for k := range Statuses {
					Statuses[k] = rand.Intn(2) > 0 // set 1/2 to alive
				}

			case "pattern":
				c.pattern = parts[1]

			default:
				parts := strings.Split(msg, " ")
				nName := parts[0]
				nAlive := parts[1] == "true"
				Statuses[nName] = nAlive

			}
			c.mut.Unlock()

			// respond to client
			out := append(buf[:n], '\n') // add newline so Readline() in client SendUDP() knows end of response
			i, err := conn.WriteTo(out, dst)
			if err != nil {
				log.Println("i:", i, ":: err: ", err)
			}
		}
	}()
	fmt.Println("server started")

	select {
	// case <-ctx.Done():
	// 	fmt.Println("cancelled")
	// 	err = ctx.Err()
	case err = <-errChan:
	}
	return
}

func (c *Cell2) UpdateNeighbors() {
	// TODO: does a goroutine actually making anything faster? maybe... special patterns are sensitive.
	var wg sync.WaitGroup
	for _, n := range c.Neighbors() {
		// c.Update(n)
		wg.Add(1)
		go func(n *Cell2) {
			defer wg.Done()
			c.Update(n)
		}(n)
	}
	wg.Wait()
}

func (c *Cell2) Update(n *Cell2) (err error) {
	// send self status to a neighbor
	d := fmt.Sprintf("%s %t", c.Name(), c.alive)
	err = SendUDP(d, n)
	if err != nil {
		log.Printf("Error updating neighbor %s: %s\n", n.Name(), err)
	}
	return
}

func (c *Cell2) SetAlive() bool {
	c.mut.RLock()
	defer c.mut.RUnlock()

	totalAlive := 0
	for _, n := range c.Neighbors() {
		alive, ok := Statuses[n.Name()]
		if !ok { // if any neighbor doesn't exist, default to alive
			c.alive = true
			return true
		}
		// fmt.Println("Alive", alive, "ok", ok)
		if alive && ok {
			totalAlive++
		}
	}
	// log.Println("unlocking in SetAlive()")
	fmt.Println("totalAlive", totalAlive)

	// Any live cell with two or three live neighbors lives on to the next generation.
	if c.alive == true && (totalAlive == 2 || totalAlive == 3) {
		return true
	}

	// Any dead cell with exactly three live neighbors becomes a live cell, as if by reproduction.
	if c.alive == false && totalAlive == 3 {
		c.alive = true
	}

	// Any live cell with fewer than two live neighbors dies, as if by underpopulation.
	if c.alive == true && totalAlive < 2 {
		c.alive = false
	}

	// Any live cell with more than three live neighbors dies, as if by overpopulation.
	if c.alive == true && totalAlive > 3 {
		c.alive = false

	}
	return c.alive
}

func (c *Cell2) Create() {
	Nomad.CreateJob(c)
}

func (c *Cell2) Destroy() {
	fmt.Println("destroying self")
	Nomad.DeleteJob(c)
}

func (c *Cell2) GetJobspec() NomadJob {
	var job NomadJob
	spec := strings.Replace(DefaultJob, "0-0", c.Name(), -1)
	json.Unmarshal([]byte(spec), &job)
	return job
}
