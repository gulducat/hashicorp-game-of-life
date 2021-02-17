package main

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Cell struct {
	x     int
	y     int
	alive bool
	addr  string
	n     map[string]*Cell

	mut sync.RWMutex
}

func NewCell(name string) Cell {
	bits := strings.Split(name, "-")
	x, _ := strconv.Atoi(bits[0])
	y, _ := strconv.Atoi(bits[1])
	return Cell{
		x:     x,
		y:     y,
		alive: true,
	}
}

func (c *Cell) Name() string {
	return fmt.Sprintf("%d-%d", c.x, c.y)
}

func (c *Cell) Service() string {
	if c.IsSeed() {
		return c.Name()
	}
	return fmt.Sprintf("cell-%d", c.Index())
}

func (c *Cell) Index() int {
	width := MaxWidth
	if c.x == 0 && c.y == 0 {
		return 0
	}
	idx := c.x + width*c.y - width
	return idx
}

func (c *Cell) IsSeed() bool {
	return c.Name() == "0-0"
}

func (c *Cell) Address() (string, error) {
	// TODO: retry if stale addr; somewhere... Update()?
	if c.addr != "" {
		return c.addr, nil
	}

	// turns out dns is wayyyy faster than http
	cdns := NewConsulDNS()
	addr, err := cdns.GetServiceAddr(c.Service())
	if err != nil {
		logger.Error("getting address", "err", err)
		c.addr = ""
		return "", err
	}

	c.addr = addr
	return addr, err
}

func (c *Cell) Neighbors() map[string]*Cell {
	if c.n != nil {
		return c.n
	}
	all := [8]*Cell{
		// comments assuming cell "2-2"

		// top row
		&Cell{x: c.x - 1, y: c.y - 1}, // 1-1
		&Cell{x: c.x, y: c.y - 1},     // 2-1
		&Cell{x: c.x + 1, y: c.y - 1}, // 3-1

		// middle row
		&Cell{x: c.x - 1, y: c.y}, // 1-2
		// 2-2 is self.
		&Cell{x: c.x + 1, y: c.y}, // 3-2

		// bottom row
		&Cell{x: c.x - 1, y: c.y + 1}, // 1-3
		&Cell{x: c.x, y: c.y + 1},     // 2-3
		&Cell{x: c.x + 1, y: c.y + 1}, // 3-3

	}
	var valid = make(map[string]*Cell)
	for _, n := range all {
		if n.x < 1 || n.y < 1 || n.x > MaxWidth || n.y > MaxHeight {
			continue
		}
		valid[n.Name()] = n
	}
	if c.Name() != "0-0" { // HACK
		c.n = valid
	}
	return valid
}

func (c *Cell) Listen() (err error) {
	addr := "0.0.0.0:" + UdpPort
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		logger.Error("starting udp server", "err", err)
		return
	}
	defer conn.Close()

	seed := NewCell("0-0")

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

			start := time.Now()
			Mut.Lock()

			parts := strings.Split(msg, " ")
			switch parts[0] {

			case "ping":
				break

			case "tick":
				c.Tick(&seed, "")

			case "pattern":
				c.Tick(&seed, parts[1])

			default: // updates from neighbors
				nName := parts[0]
				nAlive := parts[1] == "true"
				Statuses[nName] = nAlive

			}

			Mut.Unlock()

			// respond to client
			out := append(buf[:n], '\n') // add newline so Readline() in client SendUDP() knows end of response
			i, err := conn.WriteTo(out, dst)
			if err != nil {
				logger.Error("responding to client",
					"i", i,
					"err", err)
			}

			end := time.Now()
			logger.Debug("Listen",
				"msg", string(buf[:n]),
				"duration", end.Sub(start))
		}
	}()
	logger.Info("udp listener started", "addr", addr)

	select { // TODO: learn more about select
	case err = <-errChan:
	}
	return
}

func (c *Cell) Tick(seed *Cell, p string) {
	tickStart := time.Now()

	// avoid race: wait for all cells to get the tick.
	// it takes up to ~20ms for seed to finish 49 cells on laptop
	// NOTE: SendUDP's Deadline must be longer than this*8.
	sleep := time.Duration(MaxWidth * MaxHeight / 30) // TODO: hmmm.. magic.
	if sleep < 25 {
		sleep = 25
	}
	time.Sleep(sleep * time.Millisecond)

	if p == "random" {
		rand.Seed(time.Now().UnixNano())
		c.alive = rand.Intn(3) > 1 // ~1/3 of the time
	} else {
		if !ApplyPattern(c, p) {
			c.alive = c.GetNextLiveness()
		}
	}

	c.UpdateNeighbors()
	go c.Update(seed)

	tickEnd := time.Now()
	tickDelta := tickEnd.Sub(tickStart)
	logger.Debug("tickDelta", "duration", tickDelta)
}

func (c *Cell) GetNextLiveness() bool {
	totalAlive := 0
	for _, n := range c.Neighbors() {
		alive, ok := Statuses[n.Name()]
		if !ok { // if any neighbor doesn't exist yet, default to alive
			return true
		}
		if alive {
			totalAlive++
		}
	}
	logger.Info("GetNextLiveness", "totalAlive", totalAlive)

	// Any live cell with two or three live neighbors lives on to the next generation.
	// Any dead cell with exactly three live neighbors becomes a live cell, as if by reproduction.
	// Any live cell with fewer than two live neighbors dies, as if by underpopulation.
	// Any live cell with more than three live neighbors dies, as if by overpopulation.
	beAlive := false
	if c.alive == true {
		beAlive = totalAlive == 2 || totalAlive == 3 // 2 or 3
	} else {
		beAlive = totalAlive == 3 // exactly 3
	}
	return beAlive
}

func (c *Cell) UpdateNeighbors() {
	for _, n := range c.Neighbors() {
		go c.Update(n)
	}
}

func (c *Cell) Update(n *Cell) (err error) {
	// send self status to a neighbor
	maxSleep := MaxWidth * MaxHeight / 6
	if maxSleep < 50 {
		maxSleep = 50
	}
	jitter := rand.Intn(maxSleep)
	sleep := time.Duration(jitter)
	time.Sleep(sleep * time.Millisecond)
	d := fmt.Sprintf("%s %t", c.Name(), c.alive)
	err = SendUDP(d, n)
	if err != nil {
		logger.Error("updating neighbor",
			"name", n.Name(),
			"err", err)
	}
	return
}
