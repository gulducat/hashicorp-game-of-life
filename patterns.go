package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"strings"
)

// TODO: write to and have cells atch a consul kv?

// https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life#Examples_of_patterns

var Patterns = map[string]string{
	// oscillators
	"blinker": `
xox
xox
xox
`,
	"toad": `
xxxx
xooo
ooox
xxxx
`,
	"beacon": `
ooxx
ooxx
xxoo
xxoo
`,
	// spaceships
	"glider": `
xox
xxo
ooo
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
			m[cell] = val == 111 // 111 is "o"
		}
		y++
	}
	log.Println("mapped pattern:", m)
	return m, nil
}

func CheckPattern(cell *Cell2) {
	if cell.pattern == "" {
		return
	}
	pat, err := ParsePattern(cell.pattern, 0, 0)
	if err != nil {
		return
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
}
