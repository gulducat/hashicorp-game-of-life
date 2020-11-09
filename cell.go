package main

import (
	"fmt"
	"strconv"
	"strings"
)

func NewCell(name string) Cell {
	x, y := Coords(name)
	return Cell{x: x, y: y}
}

func Coords(name string) (int, int) {
	// given "1-1", return: 1, 1
	bits := strings.Split(name, "-")
	x, _ := strconv.Atoi(bits[0])
	y, _ := strconv.Atoi(bits[1])
	return x, y
}

type Cell struct {
	x int
	y int
}

func (c *Cell) Name() string {
	return fmt.Sprintf("%d-%d", c.x, c.y)
}

func (c *Cell) Neighbors(maxX int, maxY int) []Cell {
	all := [8]Cell{
		// comments assuming cell "2-2"

		// top row
		Cell{x: c.x - 1, y: c.y - 1}, // 1-1
		Cell{x: c.x, y: c.y - 1},     // 2-1
		Cell{x: c.x + 1, y: c.y - 1}, // 3-1

		// middle row
		Cell{x: c.x - 1, y: c.y}, // 1-2
		// 2-2 is self.
		Cell{x: c.x + 1, y: c.y}, // 3-2

		// bottom row
		Cell{x: c.x - 1, y: c.y + 1}, // 1-3
		Cell{x: c.x, y: c.y + 1},     // 2-3
		Cell{x: c.x + 1, y: c.y + 1}, // 3-3

	}
	var valid []Cell
	for _, n := range all {
		if n.x < 1 || n.y < 1 || n.x > maxX || n.y > maxY {
			continue
		}
		valid = append(valid, n)
	}
	return valid
}

func (c *Cell) Create() {
	job := NewNomadJob(c)
	CreateJob(&job)
}

func (c *Cell) Exists() bool {
	// TODO: check consul catalog
	return false
}

func (c *Cell) Alive() bool {
	// TODO: check consul checks api
	return true
}

func (c *Cell) Set(alive bool) {
	// TODO: store state in consul k/v
	return
}
