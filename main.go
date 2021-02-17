package main

import (
	"fmt"
	"os"
	"sync"
	"time"
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

	SetVars()
	CacheAllCells()

	seed := NewCell("0-0")
	switch arg {

	case "test":
		SendUDP("hello there", &seed)

	case "run":
		Run()

	case "pattern":
		p := os.Args[2]
		cdns := NewConsulDNS()
		addr, err := cdns.GetServiceAddr("0-0-http")
		if err != nil {
			logger.Error("getting address", "err", err)
			return
		}
		a := NewAPI("http://" + addr)
		_, body := a.Get(fmt.Sprintf("/p/%s", p))
		logger.Info(string(body))

	}
}

func GetName() (name string) {
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

func GetSelf() *Cell {
	self := NewCell(GetName())
	return &self
}

func Ticker() {
	ui := NewUI()
	inc := MaxWidth * MaxHeight / 5
	if inc < 300 {
		inc = 300
	}
	TickTime = inc
	sleep := time.Duration(inc)
	for {
		ui.UpdateGrid()
		if NextPattern != "" {
			SendToAll("pattern " + NextPattern)
			NextPattern = ""
		} else {
			SendToAll("tick tock")
		}
		logger.Info("Ticker sleep", "ms", TickTime)
		time.Sleep(sleep * time.Millisecond)
	}
}

func Run() {
	seed := NewCell("0-0")
	self := GetSelf()
	isSeed := seed.Name() == self.Name()
	logger.Info("self: " + self.Name())

	if isSeed {
		go self.Listen()
		go Ticker()
		ApiListen()
	} else {
		self.Listen()
	}

}

func CacheAllCells() {
	for x := 1; x <= MaxWidth; x++ {
		for y := 1; y <= MaxHeight; y++ {
			c := Cell{x: x, y: y}
			AllCells = append(AllCells, &c)
		}
	}
}

func SendToAll(msg string) {
	var wg sync.WaitGroup
	start := time.Now()
	for _, c := range AllCells {
		wg.Add(1)
		go func(c *Cell) {
			SendUDP(msg, c)
			wg.Done()
		}(c)
	}
	wg.Wait()
	end := time.Now()
	logger.Info("SendToAll",
		"msg", msg,
		"duration", end.Sub(start))
}
