package main

import (
	"os"
)

var ConsulAddr = os.Getenv("CONSUL_HTTP_ADDR")

func SetVars() {
	if ConsulAddr == "" {
		ConsulAddr = "http://localhost:8500"
	}
}
