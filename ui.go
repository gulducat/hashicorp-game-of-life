package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/hashicorp/go-hclog"
)

// var httpPort = os.Getenv("NOMAD_PORT_waypoint")
var httpPort = os.Getenv("NOMAD_PORT_http")
var Grid string
var NextPattern string

func ApiListen() {
	logger.Info("running api")
	ui := NewUI(logger, time.Second/2)
	if httpPort == "" {
		httpPort = "80"
	}
	logger.Info("listening on " + ":" + httpPort)
	if err := ui.ListenAndServe(":" + httpPort); err != nil {
		// logger.Info("listening on " + ":80")
		// if err := ui.ListenAndServe(":80"); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

type UI struct {
	cacheRW    sync.RWMutex
	cachedGrid []byte

	logger hclog.Logger
}

func NewUI(logger hclog.Logger, refreshRate time.Duration) *UI {
	return &UI{
		logger: logger.Named("ui"),
	}
}

func (ui *UI) ListenAndServe(address string) error {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Get("/", ui.HandleGet)
	r.Get("/p/{pattern}", ui.HandlePattern)
	return http.ListenAndServe(address, r)
}

func (ui *UI) HandlePattern(w http.ResponseWriter, r *http.Request) {
	p := chi.URLParam(r, "pattern")
	_, ok := Patterns[p]
	if !ok {
		msg := fmt.Sprintf("Invalid pattern %q", p)
		http.Error(w, msg, 404)
		return
	}
	NextPattern = p
	w.Write([]byte("set next pattern:" + p + "\n"))
}

func (ui *UI) UpdateGrid() {
	var val string
	var name string
	Mut.RLock()
	defer Mut.RUnlock()
	Grid = ""
	for y := 1; y <= MaxHeight; y++ {
		for x := 1; x <= MaxWidth; x++ {
			val = "🌑"
			name = fmt.Sprintf("%d-%d", x, y)
			alive, ok := Statuses[name]
			if ok {
				// ui.logger.Info("GET", name, "alive:", alive)
				if alive {
					val = "🟢"
				} else {
					val = "⭕️"
				}
			}
			Grid += val
		}
		Grid += "\n"
	}
}

func (ui *UI) HandleGet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(Grid))
	// TODO: separate function to build the grid on a ticker, store as a var for cache
	// w.Write([]byte(Grid))
	// var val string
	// var name string
	// // for name, alive := range Statuses {
	// // 	w.Write([]byte(fmt.Sprintf("%s %t\n", name, alive)))
	// // }
	// Mut.RLock()
	// defer Mut.RUnlock()
	// for y := 1; y <= MaxHeight; y++ {
	// 	for x := 1; x <= MaxWidth; x++ {
	// 		val = "🌑"
	// 		name = fmt.Sprintf("%d-%d", x, y)
	// 		alive, ok := Statuses[name]
	// 		if ok {
	// 			// ui.logger.Info("GET", name, "alive:", alive)
	// 			if alive {
	// 				val = "🟢"
	// 			} else {
	// 				val = "⭕️"
	// 			}
	// 		}
	// 		// w.Write([]byte(name + val))
	// 		w.Write([]byte(val))
	// 	}
	// 	w.Write([]byte("\n"))
	// }
}

// type cellStatus struct {
// 	cell   *Cell
// 	status string
// }

// func StatusGrid() []byte {
// 	var wg sync.WaitGroup
// 	services := Consul.ServiceCatalog()
// 	cellStatCh := make(chan *cellStatus, MaxHeight*MaxWidth)
// 	for x := 1; x <= MaxWidth; x++ {
// 		wg.Add(1)
// 		go func(x int) {
// 			defer wg.Done()
// 			for y := 1; y <= MaxHeight; y++ {
// 				c := &Cell{x: x, y: y}
// 				var exists bool
// 				for name := range services {
// 					if name == c.Name() {
// 						exists = true
// 						break
// 					}
// 				}
// 				val := "🌑"
// 				if exists {
// 					if c.Alive() {
// 						val = "🟢"
// 					} else {
// 						val = "⭕️"
// 					}
// 				}
// 				cellStatCh <- &cellStatus{
// 					cell:   c,
// 					status: val,
// 				}
// 			}
// 		}(x)
// 	}
// 	wg.Wait()
// 	close(cellStatCh)

// 	cellStats := make([]*cellStatus, 0, MaxHeight*MaxWidth)
// 	for cs := range cellStatCh {
// 		cellStats = append(cellStats, cs)
// 	}
// 	sort.Slice(cellStats, func(i, j int) bool {
// 		iY, iX := cellStats[i].cell.y, cellStats[i].cell.x
// 		jY, jX := cellStats[j].cell.y, cellStats[j].cell.x
// 		return iY < jY || (iY == jY && iX < jX)
// 	})
// 	var out bytes.Buffer
// 	var count int
// 	for _, cs := range cellStats {
// 		out.WriteString(cs.status)
// 		count++
// 		if count%MaxWidth == 0 {
// 			out.WriteString("\n")
// 		}
// 	}
// 	return out.Bytes()
// }
