package main

// TODO: api (or whatever) should represent all the cells as a multi-dimensional array

import (
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

const MaxSize = 3

const TmpDir = "/tmp/hgol"

// lazy "global" api clients
var Consul = NewConsul()
var Nomad = NewNomad()

// more lazy globals
var ThisDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))

func main() {
	arg := "seed"
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}
	switch arg {
	case "seed":
		Seed()
	case "run":
		Run()
	case "ui":
		UI()
	case "check":
		Check()
	case "kill":
		Kill()
	}
}

func Seed() {
	seed := NewCell("0-0")
	seed.Create()
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
	// TODO: order of operations here is pretty bad.
	seed := NewCell("0-0")

	self := GetSelf() // TODO: rename "self" ?
	log.Printf("self: %v\n", self)

	// start alive (TODO: randomize)
	self.SetStatus(true)

	neighbors := self.Neighbors(MaxSize, MaxSize)
	EnsureJobs(neighbors)

	for {
		// sleep first to give the job(s) a chance to be created.
		time.Sleep(1 * time.Second)

		if self.Name() == seed.Name() {
			continue
		}
		if !seed.Exists() {
			self.Destroy()
		}

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

func Check() {
	self := GetSelf()
	if self.GetStatus() {
		os.Exit(0)
	} else {
		os.Exit(2)
	}
}

func Kill() {
	self := GetSelf()
	self.SetStatus(false)
	log.Println(self.GetStatus())
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
