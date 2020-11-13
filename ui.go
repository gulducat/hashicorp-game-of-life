package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
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
	out := ""
	var mutex = &sync.Mutex{}

	bits := make(map[string]string)
	services := Consul.ServiceCatalog()

	for y := 1; y <= MaxHeight; y++ {
		wg := sync.WaitGroup{}
		// concurrent-ize each row
		for x := 1; x <= MaxWidth; x++ {
			c := Cell{x: x, y: y}
			wg.Add(1)

			go func(cell *Cell) {
				exists := false
				for name, _ := range services {
					if name == c.Name() {
						exists = true
						break
					}
				}
				var val string
				if exists {
					if cell.Alive() {
						val = "🟢"
					} else {
						val = "⭕️"
					}
				} else {
					val = "🌑"

				}
				mutex.Lock()
				bits[c.Name()] = val
				mutex.Unlock()

				wg.Done()
			}(&c)

		}
		wg.Wait()
	}

	for y := 1; y <= MaxHeight; y++ {
		for x := 1; x <= MaxWidth; x++ {
			c := Cell{x: x, y: y}
			out += bits[c.Name()]
			// out += fmt.Sprintf(" %s", bits[c.Name()])
		}
		out += "\n"
	}
	return out
}
