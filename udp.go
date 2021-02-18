package main

import (
	"bufio"
	"net"
	"sync"
	"time"
)

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

func SendUDP(daters string, cell *Cell) (err error) {
	for i := 0; i < 5; i++ {
		err = SendUDPOnce(daters, cell)
		if err != nil {
			cell.addr = "" // trigger cache refresh, probably should be handled elsewhere, but eh.
		} else {
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
