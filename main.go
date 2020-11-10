package main

// TODO: api (or whatever) should represent all the cells as a multi-dimensional array

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

const MaxSize = 8

// lazy "global" api clients
var Consul = NewConsul()
var Nomad = NewNomad()

func main() {
	name := os.Getenv("NOMAD_JOB_NAME")
	if name == "" {
		name = "0-0"
	}
	self := NewCell(name) // TODO: rename "self" ?
	log.Printf("self: %v\n", self)

	switch os.Args[1] {
	case "grid":
		Grid()
	case "check":
		Check(&self)
	case "kill":
		Kill(&self)
	}

	// start alive (TODO: randomize)
	self.SetStatus(true)

	neighbors := self.Neighbors(MaxSize, MaxSize)
	EnsureJobs(neighbors)

	for {
		// sleep first to give the job(s) a chance to be created.
		time.Sleep(1 * time.Second)

		// TODO: actual game of life rules
		totalAlive := 0
		for _, n := range neighbors {
			if n.Alive() {
				totalAlive += 1
			}
		}

		// lazy, and probably wrong.
		if totalAlive == 2 || totalAlive == 3 {
			self.SetStatus(true)
		} else {
			self.SetStatus(false)
		}
	}
}

func Grid() {
	var c Cell
	for y := 1; y <= MaxSize; y++ {
		for x := 1; x <= MaxSize; x++ {
			c = Cell{x: x, y: y}
			if c.Alive() {
				fmt.Printf("0 ")
				// fmt.Printf("%d-%d: O ", x, y)
			} else {
				fmt.Printf("X ")
				// fmt.Printf("%d-%d: X ", x, y)
			}
		}
		fmt.Printf("\n")
	}
	os.Exit(0)
}

func Check(cell *Cell) {
	if cell.GetStatus() {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func Kill(cell *Cell) {
	cell.SetStatus(false)
	log.Println(cell.GetStatus())
	os.Exit(0)
}

func EnsureJobs(neighbors []*Cell) {
	for _, n := range neighbors {
		// wait a bit to avoid multiple cells attempting to create the same neighbor.
		randomSleep := rand.Intn(3) + 1
		time.Sleep(time.Duration(randomSleep) * time.Second)

		if !n.Exists() {
			n.Create()
		}
	}
}
