package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"strings"
)

// TODO: write to and have cells watch a consul kv?  edit: watches are only agent config or cli

// https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life#Examples_of_patterns
// http://www.radicaleye.com/lifepage/glossary.html

/* cool things
10x9 lwss turns into a glider when it hits the edge

*/

var Patterns = map[string]string{
	"random": "",

	// oscillators
	"blinker": `
.*.
.*.
.*.
`,
	"toad": `
....
.***
***.
....
`,
	"beacon": `
**..
**..
..**
..**
`,
	// spaceships
	"glider": `
.*.
..*
***
`,
	// attempt at light weight space ship
	// best at 9x9
	"ifail1": `
.*..*
....*
*...*
.****
`,
	"lwss": `
.****
*...*
....*
*..*.
`,
	"mwss": `
..*****
.*....*
......*
.*...*.
...*...`,
	// need a bigger grid for these...
	"hwss": `
..******
.*.....*
.......*
.*....*.
...**...
`,
	"bunnies": `
.*.....*.
...*...*.
...*..*.*
..*.*....
`,
	"cross": `
...****..
...*..*..
.***..***
.*......*
.*......*
.***..***
...*..*..
...****..
`,
	"pulsar": `
...***...***..
..............
.*....*.*....*
.*....*.*....*
.*....*.*....*
...***...***..
..............
...***...***..
.*....*.*....*
.*....*.*....*
.*....*.*....*
..............
...***...***..
`,
}

func ParsePattern(name string, offsetX, offsetY int) (map[string]bool, error) {
	p, ok := Patterns[name]
	if !ok {
		msg := fmt.Sprintf("Error invalid pattern: %q", name)
		log.Println(msg)
		err := errors.New(msg)
		return nil, err
	}

	m := make(map[string]bool)
	y := 1 + offsetY
	scanner := bufio.NewScanner(strings.NewReader(p[1:]))
	for scanner.Scan() {
		for x, val := range scanner.Text() {
			cell := fmt.Sprintf("%d-%d", x+1+offsetX, y)

			m[cell] = val == 42 // 42 is "*"
		}
		y++
	}
	log.Println("mapped pattern:", m)
	return m, nil
}

func ApplyPattern(cell *Cell2) bool {
	if cell.pattern == "" {
		return false
	}

	pat, err := ParsePattern(cell.pattern, 1, 1)
	if err != nil {
		return false
	}

	val, ok := pat[cell.Name()]
	if ok {
		cell.alive = val
	} else {
		cell.alive = false
	}

	for k := range Statuses {
		val, ok := pat[k]
		if ok {
			Statuses[k] = val
		} else {
			Statuses[k] = false
		}
	}
	return true
}
