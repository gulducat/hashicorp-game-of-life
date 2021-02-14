package main

// TODO: need a better global ticker, Sleepy() is too fuzzy and depends on system clock... seed could send udp to all cells? or a separate periodic batch job?
// TODO: concurrent error handling: https://hashicorp.slack.com/archives/C01A1M2QQ1Z/p1610135235145900

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
)

var MaxWidth = 0
var MaxHeight = 0

const TmpDir = "/tmp/hgol"

// lazy "global" api clients
var logger = hclog.New(nil)
var Consul = NewConsul(logger)

func main() {
	// HACK - only let one seed job run, the one we care about for waypoint (?)
	// only our main seed job will have an "http" port.
	if os.Getenv("NOMAD_HOST_PORT_http") == "" && os.Getenv("NOMAD_ALLOC_INDEX") == "0" {
		select {} // block forever instead of killing so nomad doesn't try to replace.
	}

	MaxWidth, _ = strconv.Atoi(os.Getenv("MAX_W"))
	MaxHeight, _ = strconv.Atoi(os.Getenv("MAX_H"))
	// logger := hclog.New(nil)
	arg := "seed"
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}

	fmt.Println(os.Args)

	CacheAllCells()

	seed := NewCell2("1-1")
	switch arg {

	case "test":
		SendUDP("hello there", &seed)

	case "run":
		Run()

	case "api":
		ApiListen() // "api" is gone, long live "0-0"

	case "pattern":
		p := os.Args[2]
		cdns := NewConsulDNS()
		addr, err := cdns.GetServiceAddr("0-0-http")
		if err != nil {
			log.Fatal("OH NO:", err)
		}
		a := NewAPI("http://"+addr, logger)
		_, body := a.Get(fmt.Sprintf("/p/%s", p))
		fmt.Println(string(body))

	case "dnstest":
		// fmt.Println(Consul.Service("0-0"))
		cdns := NewConsulDNS()
		addr, err := cdns.GetServiceAddr("0-0")
		if err != nil {
			log.Println(err)
		}
		fmt.Println(addr)

	}
}

func GetName() (name string) {
	idx, err := strconv.Atoi(os.Getenv("NOMAD_ALLOC_INDEX"))
	if err != nil {
		fmt.Println("ERR getting alloc idx:", err)
		return name
	}
	width, err := strconv.Atoi(os.Getenv("MAX_W"))
	if err != nil {
		fmt.Println("ERR getting MAX_W")
	}

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

func GetSelf() *Cell2 {
	name := GetName()
	if name == "" {
		log.Fatal("aint a nomad job")
	}
	self := NewCell2(name)
	return &self
}

func Ticker() {
	ui := NewUI(logger, 0)
	sleep := time.Duration(MaxWidth * MaxHeight / 2)
	if sleep < 100 {
		sleep = 100
	}
	for {
		ui.UpdateGrid()
		// TODO: collapse to a single SendToAll()
		if NextPattern != "" {
			SendToAll("pattern " + NextPattern)
			NextPattern = ""
		}
		SendToAll("tick tock")
		time.Sleep(sleep * time.Millisecond)
	}
}

func Run() {
	seed := NewCell2("0-0")
	self := GetSelf()
	isSeed := seed.Name() == self.Name()
	fmt.Println("self:", self.Name())

	if isSeed {
		go self.Listen()
		go Ticker()
		ApiListen()
	} else {
		self.Listen()
	}

}

var AllCells []*Cell2

func CacheAllCells() {
	for x := 1; x <= MaxWidth; x++ {
		for y := 1; y <= MaxHeight; y++ {
			c := Cell2{x: x, y: y}
			AllCells = append(AllCells, &c)
		}
	}
}

func SendToAll(msg string) {
	// TODO: move logic from cli to "api" so it can cache service addresses.
	// TODO: WaitGroup is only needed for running from laptop. (?)
	var wg sync.WaitGroup // TODO: does this concurrency really help?  yes, yes it most certainly does.
	start := time.Now()
	for _, c := range AllCells {
		wg.Add(1)
		go func(c *Cell2) {
			SendUDP(msg, c)
			wg.Done()
		}(c)
	}
	wg.Wait()
	end := time.Now()
	log.Println("SendToAll duration:", end.Sub(start))
}
