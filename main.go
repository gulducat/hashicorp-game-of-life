package main

// TODO: need a better global ticker, Sleepy() is too fuzzy and depends on system clock... seed could send udp to all cells? or a separate periodic batch job?
// TODO: concurrent error handling: https://hashicorp.slack.com/archives/C01A1M2QQ1Z/p1610135235145900

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
)

// 15*16 = 240
// 24*21 = 504 / 7 clients = 72 per node
// const MaxWidth = 24
// const MaxHeight = 21

const MaxWidth = 7
const MaxHeight = 7

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
	switch arg {

	case "run":
		Run()

	case "api":
		ApiListen() // "api" is gone, long live "0-0" (for now)

	case "seed":
		seed := NewCell2("0-0")
		seed.Create()

	case "more":
		SendToAll("random xxx")

	case "pattern":
		// Sleepy()
		// time.Sleep(200 * time.Millisecond)
		p := os.Args[2]
		_, ok := Patterns[p]
		if !ok {
			log.Fatalf("Invalid pattern %q", p)
		}
		SendToAll("pattern " + p)

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

func Sleepy() {
	// lazy attempt to sync up cells... on even seconds.
	// it works pretty well!  at least on a single computer (my laptop)
	now := time.Now()
	// these aren't really half seconds...
	toHalfSecond := (1000000000 - now.Nanosecond()) // this one's a second
	// toHalfSecond := (1000000000 - now.Nanosecond()) / 2 // these
	// toHalfSecond := (1000000000 - now.Nanosecond()) - 500000000 // are failures of the imagination.
	duration, err := time.ParseDuration(fmt.Sprintf("%dns", toHalfSecond))
	if err != nil {
		log.Println("Error getting duration...", err)
		return
	}
	log.Printf("Now: %s ; sleeping for: %s", now, duration)
	time.Sleep(duration)
}

func Run() {
	seed := NewCell2("0-0")
	self := GetSelf()
	isSeed := seed.Name() == self.Name()
	fmt.Println("self:", self.Name())

	defer self.Destroy()

	if !isSeed {
		if !seed.WaitUntilExists(10) { // try for 10 seconds
			fmt.Println("no service for seed 0-0 at launch")
			return
		}
	}

	neighbors := self.Neighbors()
	// time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond) // TODO: add back Exists() check in EnsureJobs()?
	EnsureJobs(neighbors)

	go self.Listen()
	if isSeed {
		ApiListen() // blocks, seed stops here.
	}

	// main loop
	var err error
	deadSeeds := 0      // to stay alive if 0-0 only *temporarily* goes down
	for deadSeeds < 5 { // TODO: 15
		// sleep first to give the job(s) a chance to be created.
		// time.Sleep(500 * time.Millisecond)
		Sleepy() // try to keep things in sync

		CheckPattern(self)

		self.SetAlive()
		self.UpdateNeighbors()

		err = self.Update(&seed)
		if err == nil {
			deadSeeds = 0
		} else {
			deadSeeds++
		}
		fmt.Println("deadSeeds:", deadSeeds)
	}

}

func EnsureJobs(cells map[string]*Cell2) {
	for name, c := range cells {
		// if !c.Exists() {
		fmt.Println("Creating job:", name)
		c.Create()
		// }
	}
}

func SendToAll(msg string) {
	// TODO: move logic to "api" so it can cache service addresses.

	// Sleepy()
	// time.Sleep(100 * time.Millisecond)

	var wg sync.WaitGroup // TODO: does this concurrency really help?
	for x := 1; x <= MaxWidth; x++ {
		for y := 1; y <= MaxHeight; y++ {
			wg.Add(1)
			go func(x, y int) {
				defer wg.Done()
				c := Cell2{x: x, y: y, alive: true, cdns: NewConsulDNS()}
				SendUDP(msg, &c)
				// c.UpdateNeighbors()
			}(x, y)
		}
	}
	wg.Wait()
}
