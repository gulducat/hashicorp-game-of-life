package main

import "time"

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
