package main

import (
	"fmt"
	"log"
	"net/http"
)

func ServeWeb() {
	http.Handle("/", new(handler))
	log.Fatal(http.ListenAndServe(":80", nil))
}

type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, StatusGrid())

	// log for some reason
	fmt.Printf("%s %s %s %s \"%s\"\n",
		r.RemoteAddr, r.Host, r.Method, r.URL, r.UserAgent())
}

func StatusGrid() string {
	var c Cell
	var out string
	for y := 1; y <= MaxHeight; y++ {
		for x := 1; x <= MaxWidth; x++ {
			c = Cell{x: x, y: y}
			if !c.Exists() {
				out = fmt.Sprintf("%s 🌑", out)
				continue
			}
			if c.Alive() {
				out = fmt.Sprintf("%s 🟢", out)
			} else {
				out = fmt.Sprintf("%s ⭕️", out)
			}
		}
		out = out + "\n"
	}
	return out
}
