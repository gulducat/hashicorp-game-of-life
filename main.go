package main

import (
	"fmt"
	"os"
)

func main() {
	// HACK - only let one seed job run, the one we care about for waypoint (?)
	// only our main seed job will have an "http" port.
	if os.Getenv("NOMAD_HOST_PORT_http") == "" && os.Getenv("NOMAD_ALLOC_INDEX") == "0" {
		select {} // block forever instead of killing so nomad doesn't try to replace.
	}

	arg := "seed"
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}

	SetGlobals()

	seed := NewCell("0-0")
	switch arg {
	case "test":
		SendUDP("hello there", &seed)
	case "run":
		Run(&seed)
	case "pattern":
		SetPattern(os.Args[2])
	}
}

func Run(seed *Cell) {
	self := NewCell(GetName())
	isSeed := seed.Name() == self.Name()
	logger.Info("self: " + self.Name())

	if isSeed {
		CacheAllCells()
		go self.Listen()
		go Ticker()
		ApiListen()
	} else {
		self.Listen()
	}
}

func GetName() (name string) {
	// see vars.go for these globals
	idx := AllocIdx
	width := MaxWidth

	var x, y int
	if idx == 0 {
		return "0-0"
	}
	x = idx % width
	if x == 0 {
		x = width
	}
	y = (idx-1)/width + 1

	name = fmt.Sprintf("%d-%d", x, y)
	return name
}

func CacheAllCells() {
	for x := 1; x <= MaxWidth; x++ {
		for y := 1; y <= MaxHeight; y++ {
			c := Cell{x: x, y: y}
			AllCells = append(AllCells, &c)
		}
	}
}

func SetPattern(p string) {
	cdns := NewConsulDNS()
	addr, err := cdns.GetServiceAddr("0-0-http")
	if err != nil {
		logger.Error("getting address", "err", err)
		return
	}
	h := NewHTTP("http://" + addr)
	_, body := h.Get(fmt.Sprintf("/p/%s", p))
	logger.Info(string(body))
}
