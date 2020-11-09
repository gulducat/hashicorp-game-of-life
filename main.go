package main

// TODO: api (or whatever) should represent all the cells as a multi-dimensional array

import (
	"fmt"
	"os"
	"time"
)

var Name = os.Getenv("NOMAD_JOB_NAME")

func main() {
	name := "0-0"
	if Name != "" {
		name = Name
	}
	x, y := Coords(name)
	self := Cell{x: x, y: y}
	fmt.Printf("self: %v\n", self)

	neighbors := self.Neighbors(3, 3)
	for _, n := range neighbors {
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
