package main

import (
	"os"
	"strconv"
	"sync"

	"github.com/hashicorp/go-hclog"
)

const datacenter = "dc1"

var MaxWidth = 0
var MaxHeight = 0

var logger = hclog.New(nil)

var ConsulAddr = os.Getenv("CONSUL_HTTP_ADDR")
var UdpPort = os.Getenv("NOMAD_HOST_PORT_udp")
var httpPort = os.Getenv("NOMAD_PORT_http")

var AllCells []*Cell
var Statuses = make(map[string]bool) // TODO: not this
var Mut sync.RWMutex

var Grid string
var NextPattern string
var TickTime int

func SetVars() {
	MaxWidth, _ = strconv.Atoi(os.Getenv("MAX_W"))
	MaxHeight, _ = strconv.Atoi(os.Getenv("MAX_H"))

	if ConsulAddr == "" {
		ConsulAddr = "http://localhost:8500"
	}

	if httpPort == "" {
		httpPort = "80"
	}
}
