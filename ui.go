package main

import "fmt"

func UI() {
	var c Cell
	for y := 1; y <= MaxSize; y++ {
		for x := 1; x <= MaxSize; x++ {
			c = Cell{x: x, y: y}
			if c.Alive() {
				fmt.Printf("✅ ")
				// fmt.Printf("%d-%d: O ", x, y)
			} else {
				fmt.Printf("🛑 ")
				// fmt.Printf("%d-%d: X ", x, y)
			}
		}
		fmt.Printf("\n")
	}
}
