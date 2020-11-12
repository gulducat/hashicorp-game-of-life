package main

// TODO: api (or whatever) should represent all the cells as a multi-dimensional array

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

const MaxWidth = 7
const MaxHeight = 8

const TmpDir = "/tmp/hgol"

// lazy "global" api clients
var Consul = NewConsul()
var Nomad = NewNomad()

// more lazy globals
// var ThisDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))

func main() {
	arg := "seed"
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}
	switch arg {

	case "run":
		Run()

	case "api":
		ServeWeb()

	case "seed":
		seed := NewCell("0-0")
		seed.Create()

	case "check":
		self := GetSelf()
		if self.GetStatus() {
			os.Exit(0)
		} else {
			os.Exit(1)
		}

	case "more":
		Reset()

	}
}

func GetSelf() *Cell {
	name := os.Getenv("NOMAD_JOB_NAME")
	if name == "" {
		panic("aint a nomad job")
	}
	self := NewCell(name)
	return &self
}

func Run() {
	seed := NewCell("0-0")

	self := GetSelf() // TODO: rename "self" ?
	log.Printf("self: %v\n", self)

	rando := rand.Intn(2)
	startingStatus := false
	if rando == 1 {
		startingStatus = true
	}
	self.SetStatus(startingStatus)
	// self.SetStatus(true)

	neighbors := self.Neighbors(MaxWidth, MaxHeight)
	EnsureJobs(neighbors)

	for {
		// sleep first to give the job(s) a chance to be created.
		time.Sleep(1 * time.Second)

		if !seed.Exists() {
			self.Destroy()
			return
		}

		selfStatus := self.GetStatus()

		totalAlive := 0
		for _, n := range neighbors {
			if n.Alive() {
				totalAlive += 1
			}
		}

		//Any live cell with two or three live neighbors lives on to the next generation.
		if selfStatus == true && (totalAlive == 2 || totalAlive == 3) {
			continue
		}

		//Any dead cell with exactly three live neighbors becomes a live cell, as if by reproduction.
		if selfStatus == false && totalAlive == 3 {
			self.SetStatus(true)
			continue
		}

		//Any live cell with fewer than two live neighbors dies, as if by underpopulation.
		if selfStatus == true && totalAlive < 2 {
			self.SetStatus(false)
			continue
		}

		//Any live cell with more than three live neighbors dies, as if by overpopulation.
		if selfStatus == true && totalAlive > 3 {
			self.SetStatus(false)
		}
	}
}

func EnsureJobs(neighbors []*Cell) {
	for _, n := range neighbors {
		fmt.Println("Creating job:", n.Name())
		n.Create()
	}
}

func Reset() {
	var c Cell
	for y := 1; y <= MaxHeight; y++ {
		for x := 1; x <= MaxWidth; x++ {
			c = Cell{x: x, y: y}
			c.SetStatus(true)
		}
	}
}
