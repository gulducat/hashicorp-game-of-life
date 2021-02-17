package main

import (
	"bufio"
	"net"
	"time"
)

func SendUDP(daters string, cell *Cell) (err error) {
	for i := 0; i < 5; i++ {
		err = SendUDPOnce(daters, cell)
		if err == nil {
			break
		}
	}
	return
}

func SendUDPOnce(daters string, cell *Cell) (err error) {
	addr, err := cell.Address()
	if err != nil {
		return
	}

	start := time.Now()

	conn, err := net.Dial("udp", addr)
	if err != nil {
		logger.Error("getting UDP port",
			"name", cell.Name(),
			"err", err)
		return
	}
	defer conn.Close()

	// remember: this fixed my "jobs dont die becuause frozen" bug
	now := time.Now()
	err = conn.SetDeadline(now.Add(150 * time.Millisecond)) // TODO: longer when run on a cluster?
	if err != nil {
		logger.Error("setting deadline", "err", err)
	}

	_, err = conn.Write([]byte(daters))
	if err != nil {
		logger.Error("sending",
			"daters", daters,
			"name", cell.Name(),
			"err", err)
	}

	_, _, err = bufio.NewReader(conn).ReadLine()
	if err != nil {
		logger.Error("reading response",
			"name", cell.Name(),
			"err", err)
		return
	}

	end := time.Now()
	logger.Debug("SendUDP",
		"daters", daters,
		"target", cell.Name(),
		"duration", end.Sub(start))

	return
}
