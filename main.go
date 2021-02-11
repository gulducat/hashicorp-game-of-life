package main

// TODO: need a better global ticker, Sleepy() is too fuzzy and depends on system clock... seed could send udp to all cells? or a separate periodic batch job?
// TODO: concurrent error handling: https://hashicorp.slack.com/archives/C01A1M2QQ1Z/p1610135235145900

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
)

// 15*16 = 240
// 24*21 = 504 / 7 clients = 72 per node
// const MaxWidth = 24
// const MaxHeight = 21

// const MaxWidth = 15
// const MaxHeight = 12

// TODO: why does the ui start being wobbly?
// good for 4 clients; 77 per node
const MaxWidth = 18
const MaxHeight = 17

// 6 clients, 77 per
// const MaxWidth = 23
// const MaxHeight = 20

// 11x10 is too big in vagrant (nomad gets real sad)
// *INTERESTINGLY* it runs better in a vagrant, way less host CPU
// const MaxWidth = 3
// const MaxHeight = 3

const TmpDir = "/tmp/hgol"

// lazy "global" api clients
var logger = hclog.New(nil)
var Consul = NewConsul(logger)
var Nomad = NewNomad(logger)

func main() {
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

	case "seed":
		seed := NewCell2("0-0")
		seed.Create()

	case "more":
		// SendToAll("random xxx")
		SendToAll("pattern random")
		// SendUDP("pattern random", &seed)

	case "pattern":
		// SendUDP("pattern "+os.Args[2], &seed)
		// Sleepy()
		// time.Sleep(200 * time.Millisecond)
		p := os.Args[2]
		_, ok := Patterns[p]
		if !ok {
			log.Fatalf("Invalid pattern %q", p)
		}
		SendToAll("pattern " + p)

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

func GetSelf() *Cell2 {
	name := os.Getenv("NOMAD_JOB_NAME")
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
		SendToAll("tick tock")
		time.Sleep(sleep * time.Millisecond)
	}
}

func WatchSeed(s, c *Cell2) {
	defer c.Destroy()
	dead := 0
	// allow seed going down for up to 10 seconds
	for dead < 5 {
		time.Sleep(2 * time.Second)
		err := SendUDP("ping pong", s)
		if err == nil {
			dead = 0
		} else {
			dead++
		}
	}
}

func Run() {
	seed := NewCell2("0-0")
	self := GetSelf()
	isSeed := seed.Name() == self.Name()
	fmt.Println("self:", self.Name())

	// defer self.Destroy()

	if !isSeed {
		if !seed.WaitUntilExists(10) { // try for 10 seconds
			fmt.Println("no service for seed 0-0 at launch")
			self.Destroy()
			return
		}
	}

	neighbors := self.Neighbors()
	EnsureJobs(self, neighbors)

	if isSeed {
		go self.Listen()
		go Ticker()
		ApiListen()
	} else {
		go WatchSeed(&seed, self)
		self.Listen()
	}

}

func EnsureJobs(self *Cell2, cells map[string]*Cell2) {
	for name, c := range cells {
		// cells above and below already exist, one of them created me.
		if self.x > c.x || self.y > c.y {
			continue
		}
		// try not to stampede nomad api
		sleep := time.Duration(100 + rand.Intn(200))
		time.Sleep(sleep * time.Millisecond)
		if !c.Exists() {
			fmt.Println("Creating job:", name)
			c.Create()
		}
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
