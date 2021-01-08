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

func ApiListen() {
	logger.Info("running api")
	ui := NewUI(logger, time.Second/2)
	logger.Info("listening on " + ":80")
	if err := ui.ListenAndServe(":80"); err != nil {
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
	return http.ListenAndServe(address, r)
}

func (ui *UI) HandleGet(w http.ResponseWriter, r *http.Request) {
	// TODO: separate function to build the grid on a ticker, store as a var for cache
	var val string
	var name string
	// for name, alive := range Statuses {
	// 	w.Write([]byte(fmt.Sprintf("%s %t\n", name, alive)))
	// }
	for x := 1; x <= MaxWidth; x++ {
		for y := 1; y <= MaxHeight; y++ {
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
			// w.Write([]byte(name + val))
			w.Write([]byte(val))
		}
		w.Write([]byte("\n"))
	}
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
