package main

// TODO: api (or whatever) should represent all the cells as a multi-dimensional array

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

const MaxSize = 3

func main() {
	name := os.Getenv("NOMAD_JOB_NAME")
	if name == "" {
		name = "0-0"
	}
	self := NewCell(name)
	fmt.Printf("self: %v\n", self)

	if os.Args[1] == "check" {
		// TODO: actually check something
		os.Exit(0)
	}

	neighbors := self.Neighbors(MaxSize, MaxSize)

	// dynamically generate neighbor jobs
	for _, n := range neighbors {
		// wait a tick to avoid multiple cells attempting to create the same neighbor.
		randomSleep := rand.Intn(3) + 1
		time.Sleep(time.Duration(randomSleep) * time.Second)

		if !n.Exists() {
			n.Create()
		}
	}

	for {
		// TODO: actual game of life rules
		totalAlive := 0
		for _, n := range neighbors {
			if n.Alive() {
				totalAlive += 1
			}
		}
		if totalAlive == 3 {
			self.Set(false)
		}
		time.Sleep(1 * time.Second)
	}
}
