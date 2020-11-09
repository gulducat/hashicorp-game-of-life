package main

// TODO: api (or whatever) should represent all the cells as a multi-dimensional array

import (
	"fmt"
	"strconv"
	"strings"
)

func Coords(name string) (int, int) {
	// given "1-1", return: 1, 1
	bits := strings.Split(name, "-")
	// ^ bits = ["1", "1"]
	// make em integers for math reasons
	x, _ := strconv.Atoi(bits[0])
	y, _ := strconv.Atoi(bits[1])
	return x, y
}

func main() {
	name := "0-0"
	x, y := Coords(name)
	c := Cell{x: x, y: y}
	fmt.Println(c)
	fmt.Println(c.Neighbors(3, 3))
}
