package main

import (
	"os"
	"strconv"
	"sync"

	"github.com/hashicorp/go-hclog"
)

const datacenter = "dc1"

var MaxWidth int
var MaxHeight int

var logger = hclog.New(&hclog.LoggerOptions{
	Name:  "",
	Level: hclog.LevelFromString("DEBUG"),
})

var ConsulAddr = os.Getenv("CONSUL_HTTP_ADDR")
var UdpPort = os.Getenv("NOMAD_HOST_PORT_udp")
var httpPort = os.Getenv("NOMAD_PORT_http")

var AllocIdx int

var AllCells []*Cell
var Statuses = make(map[string]bool) // TODO: not this
var Mut sync.RWMutex

var Grid string
var NextPattern string
var TickTime int

func SetGlobals() {
	var err error
	MaxWidth, err = strconv.Atoi(os.Getenv("MAX_W"))
	if err != nil {
		logger.Debug("using default MaxWidth: 9")
		MaxWidth = 9
	}
	MaxHeight, err = strconv.Atoi(os.Getenv("MAX_H"))
	if err != nil {
		logger.Debug("using default MaxHeight: 7")
		MaxHeight = 7
	}

	if ConsulAddr == "" {
		ConsulAddr = "http://localhost:8500"
	}

	if httpPort == "" {
		httpPort = "80"
	}

	idx, err := strconv.Atoi(os.Getenv("NOMAD_ALLOC_INDEX"))
	if err == nil {
		AllocIdx = idx
	} else {
		logger.Debug("using default idx: 0")
	}

}
