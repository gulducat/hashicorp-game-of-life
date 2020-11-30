package main

// TODO: api (or whatever) should represent all the cells as a multi-dimensional array

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
)

// 15*16 = 240
var (
	MaxWidth  = *flag.Int("max_width", 8, "set the max width")
	MaxHeight = *flag.Int("max_height", 8, "set the max height")
)

// const MaxWidth = 7
// const MaxHeight = 8

const TmpDir = "/tmp/hgol"

// lazy "global" api clients
var logger = hclog.New(nil)
var Consul = NewConsul(logger)
var Nomad = NewNomad()

// more lazy globals
// var ThisDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))

func main() {
	logger := hclog.New(nil)
	flag.Parse()
	arg := "seed"
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}
	switch arg {

	case "run":
		Run()

	case "api":
		logger.Info("running api")
		ui, err := NewUI(logger, time.Second)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		logger.Info("listening on " + ":80")
		if err := ui.ListenAndServe(":80"); err != nil {
			logger.Error(err.Error())
		}

	case "seed":
		seed := NewCellRunner(0, 0, logger)
		Nomad.CreateJob(seed.CellStatus)

	case "check":
		self := GetSelf()
		if self.GetStatus() == Alive {
			os.Exit(0)
		} else {
			os.Exit(1)
		}

	case "more":
		Reset()

	}
}

func GetSelf() *CellRunner {
	name := os.Getenv("NOMAD_JOB_NAME")
	if name == "" {
		panic("aint a nomad job")
	}
	x, y := Coords(name)
	return NewCellRunner(x, y, logger)
}

func Run() {
	seed := NewCellRunner(0, 0, logger)

	self := GetSelf() // TODO: rename "self" ?
	log.Printf("self: %v\n", self)

	rando := rand.Intn(2)
	startingStatus := false
	if rando == 1 {
		startingStatus = true
	}
	self.SetStatus(startingStatus)
	// self.SetStatus(true)

	neighbors := self.Neighbors(MaxWidth, MaxHeight)
	EnsureJobs(neighbors)

	for {
		// sleep first to give the job(s) a chance to be created.
		time.Sleep(1 * time.Second)

		if !seed.Exists() {
			self.Destroy()
			return
		}

		selfStatus := self.GetStatus()

		totalAlive := 0
		for _, n := range neighbors {
			if n.status == Alive {
				totalAlive += 1
			}
		}

		//Any live cell with two or three live neighbors lives on to the next generation.
		if selfStatus == Alive && (totalAlive == 2 || totalAlive == 3) {
			continue
		}

		//Any dead cell with exactly three live neighbors becomes a live cell, as if by reproduction.
		if selfStatus == Dead && totalAlive == 3 {
			self.SetStatus(true)
			continue
		}

		//Any live cell with fewer than two live neighbors dies, as if by underpopulation.
		if selfStatus == Alive && totalAlive < 2 {
			self.SetStatus(false)
			continue
		}

		//Any live cell with more than three live neighbors dies, as if by overpopulation.
		if selfStatus == Alive && totalAlive > 3 {
			self.SetStatus(false)
			continue
		}

	}
}

func EnsureJobs(neighbors map[string]*CellStatus) {
	for _, n := range neighbors {
		fmt.Println("Creating job:", n.Name())
		Nomad.CreateJob(n)
	}
}

func Reset() {
	nullLog := hclog.NewNullLogger()
	for y := 1; y <= MaxHeight; y++ {
		for x := 1; x <= MaxWidth; x++ {
			c := NewCellRunner(x, y, nullLog)
			c.SetStatus(true)
		}
	}
}
