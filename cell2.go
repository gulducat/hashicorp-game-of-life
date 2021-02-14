package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

var Statuses = make(map[string]bool) // TODO: not this
var Mut sync.RWMutex

type Cell2 struct {
	x     int
	y     int
	alive bool
	addr  string
	n     map[string]*Cell2

	mut sync.RWMutex
	// server *UDPServer
}

func NewCell2(name string) Cell2 {
	x, y := Coords(name)
	// udpServer := NewUDPServer2()
	return Cell2{
		x:     x,
		y:     y,
		alive: true,
		// n:     make(map[string]*Cell2),
	}
}

func (c *Cell2) Name() string {
	return fmt.Sprintf("%d-%d", c.x, c.y)
}

func (c *Cell2) Service() string {
	if c.IsSeed() {
		return c.Name()
	}
	return fmt.Sprintf("CELL-%d", c.Index())
}

func (c *Cell2) Index() int {
	width := MaxWidth
	if c.x == 0 && c.y == 0 {
		return 0
	}
	idx := c.x + width*c.y - width
	return idx
}

func (c *Cell2) IsSeed() bool {
	return c.Name() == "0-0"
}

func (c *Cell2) Address() (string, error) {
	// TODO: retry if stale addr; somewhere... Update()?
	if c.addr != "" {
		return c.addr, nil
	}

	// addr := ""
	// svc, err := Consul.Service(c.Name())
	// if err != nil {
	// 	return "", err
	// }
	// addr = fmt.Sprintf("%s:%d", svc.Address, svc.Port)

	// turns out dns is wayyyy faster than http
	cdns := NewConsulDNS()
	addr, err := cdns.GetServiceAddr(c.Service())
	if err != nil {
		log.Println(err)
		c.addr = ""
		return "", err
	}

	c.addr = addr
	return addr, err
}

func (c *Cell2) Exists() bool {
	return Consul.ServiceExists(c.Service())
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
		&Cell2{x: c.x - 1, y: c.y - 1}, // 1-1
		&Cell2{x: c.x, y: c.y - 1},     // 2-1
		&Cell2{x: c.x + 1, y: c.y - 1}, // 3-1

		// middle row
		&Cell2{x: c.x - 1, y: c.y}, // 1-2
		// 2-2 is self.
		&Cell2{x: c.x + 1, y: c.y}, // 3-2

		// bottom row
		&Cell2{x: c.x - 1, y: c.y + 1}, // 1-3
		&Cell2{x: c.x, y: c.y + 1},     // 2-3
		&Cell2{x: c.x + 1, y: c.y + 1}, // 3-3

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

func (c *Cell2) Tick(seed *Cell2) {
	tickStart := time.Now()

	// avoid race: wait for all cells to get the tick.
	// it takes up to ~20ms for seed to finish 49 cells on laptop
	// NOTE: SendUDP's Deadline must be longer than this*8.
	sleep := time.Duration(MaxWidth * MaxHeight / 30) // TODO: hmmm.. magic.
	time.Sleep(sleep * time.Millisecond)

	c.alive = c.GetNextLiveness()
	c.UpdateNeighbors()
	go c.Update(seed)

	tickEnd := time.Now()
	tickDelta := tickEnd.Sub(tickStart)
	log.Println("tickDelta:", tickDelta)
}

func (c *Cell2) Listen() (err error) {
	// addr := os.Getenv("NOMAD_ADDR_udp")
	addr := "0.0.0.0:" + UdpPort
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Println("Error starting server:", err)
		return
	}
	defer conn.Close()

	seed := NewCell2("0-0")

	errChan := make(chan error, 1)
	buf := make([]byte, 512)
	go func() {
		// twixtTicks := 0
		// var mut sync.Mutex

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
			// log.Printf("serv recv %s", msg)

			// start := time.Now()
			Mut.Lock()

			parts := strings.Split(msg, " ")
			switch parts[0] {

			case "ping":
				break

			case "tick":
				c.Tick(&seed)

			case "pattern":
				p := parts[1]
				if p == "random" {
					rand.Seed(time.Now().UnixNano())
					c.alive = rand.Intn(3) > 1 // ~1/3 of the time
				} else {
					ApplyPattern(c, p)
				}
				c.Tick(&seed)

			default: // updates from neighbors
				// twixtTicks++
				nName := parts[0]
				nAlive := parts[1] == "true"
				Statuses[nName] = nAlive

				// rebuild the whole grid every update? is that any kind of improvement?
				// if c.IsSeed() {
				// 	grid := ""
				// 	for y := 1; y <= MaxHeight; y++ {
				// 		for x := 1; x <= MaxWidth; x++ {
				// 			val := "🌑"
				// 			name := fmt.Sprintf("%d-%d", x, y)
				// 			alive, ok := Statuses[name]
				// 			if ok {
				// 				if alive {
				// 					val = "🟢"
				// 				} else {
				// 					val = "⭕️"
				// 				}
				// 			}
				// 			grid += val
				// 		}
				// 		grid += "\n"
				// 	}
				// 	Grid = grid
				// }

			}

			Mut.Unlock()

			// respond to client
			out := append(buf[:n], '\n') // add newline so Readline() in client SendUDP() knows end of response
			i, err := conn.WriteTo(out, dst)
			if err != nil {
				log.Println("i:", i, ":: err: ", err)
			}

			// end := time.Now()
			// log.Printf("listen duration: %s (%q)", end.Sub(start), buf[:n])
		}
	}()
	fmt.Println("server started on", addr)

	select { // TODO: learn more about select
	// case <-ctx.Done():
	// 	fmt.Println("cancelled")
	// 	err = ctx.Err()
	case err = <-errChan:
	}
	return
}

func (c *Cell2) UpdateNeighbors() {
	// TODO: does a goroutine actually making anything faster? maybe... special patterns are sensitive.
	// var wg sync.WaitGroup
	for _, n := range c.Neighbors() {
		// c.Update(n)
		// wg.Add(1)
		// go func(n *Cell2) {
		// 	defer wg.Done()
		go c.Update(n)
		// }(n)
	}
	// wg.Wait()
}

func (c *Cell2) Update(n *Cell2) (err error) {
	// send self status to a neighbor
	maxSleep := MaxWidth * MaxHeight / 6
	jitter := rand.Intn(maxSleep)
	sleep := time.Duration(jitter)
	time.Sleep(sleep * time.Millisecond)
	d := fmt.Sprintf("%s %t", c.Name(), c.alive)
	err = SendUDP(d, n)
	if err != nil {
		log.Printf("Error updating neighbor %s: %s\n", n.Name(), err)
	}
	return
}

func (c *Cell2) GetNextLiveness() bool {
	// c.mut.Lock()
	// defer c.mut.Unlock()

	totalAlive := 0
	for _, n := range c.Neighbors() {
		alive, ok := Statuses[n.Name()]
		if !ok { // if any neighbor doesn't exist yet, default to alive
			// c.alive = true
			return true
		}
		// fmt.Println("Alive", alive, "ok", ok)
		if alive {
			totalAlive++
		}
	}
	fmt.Println("totalAlive", totalAlive)

	// // aliveBefore := c.alive

	// // Any live cell with two or three live neighbors lives on to the next generation.
	// // Any dead cell with exactly three live neighbors becomes a live cell, as if by reproduction.
	// beAlive := (totalAlive == 2 || totalAlive == 3)
	// fmt.Println("beAlive", beAlive)
	// // Any live cell with fewer than two live neighbors dies, as if by underpopulation.
	// // Any live cell with more than three live neighbors dies, as if by overpopulation.
	// beDead := (totalAlive < 2 || totalAlive > 3) // covered by beAlive
	// // changed := (c)
	// // if beAlive {
	// // 	c.alive = true
	// // }
	// // if beDead {
	// // 	c.alive = false
	// // }

	// return beAlive || !beDead
	// // return aliveBefore != c.alive
	beAlive := false

	if c.alive == true {
		beAlive = totalAlive == 2 || totalAlive == 3 // 2 or 3
	} else {
		beAlive = totalAlive == 3 // exactly 3
	}

	return beAlive

	// Any live cell with two or three live neighbors lives on to the next generation.
	// if c.alive == true && (totalAlive == 2 || totalAlive == 3) {
	// 	return true
	// }
	// // Any dead cell with exactly three live neighbors becomes a live cell, as if by reproduction.
	// if c.alive == false && totalAlive == 3 {
	// 	c.alive = true
	// }
	// // Any live cell with fewer than two live neighbors dies, as if by underpopulation.
	// if c.alive == true && totalAlive < 2 {
	// 	c.alive = false
	// }
	// // Any live cell with more than three live neighbors dies, as if by overpopulation.
	// if c.alive == true && totalAlive > 3 {
	// 	c.alive = false

	// }
	// return c.alive
}
