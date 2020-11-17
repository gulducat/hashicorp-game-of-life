package main

import (
	"bytes"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/hashicorp/go-hclog"
)

type UI struct {
	cacheRW    sync.RWMutex
	cachedGrid []byte

	logger hclog.Logger
}

func NewUI(logger hclog.Logger, refreshRate time.Duration) (*UI, error) {
	ui := &UI{
		logger: logger.Named("ui"),
	}
	ui.startGridWatcher(refreshRate)
	return ui, nil
}

func (ui *UI) updateGrid() {
	grid := StatusGrid()
	ui.cacheRW.Lock()
	ui.cachedGrid = grid
	ui.cacheRW.Unlock()
}

func (ui *UI) startGridWatcher(refreshRate time.Duration) {
	go func() {
		ui.updateGrid()
		tick := time.Tick(refreshRate)
		for range tick {
			ui.updateGrid()
		}
	}()
}

func (ui *UI) HandleGet(w http.ResponseWriter, r *http.Request) {
	ui.cacheRW.RLock()
	if _, err := w.Write(ui.cachedGrid); err != nil {
		ui.logger.Error(err.Error())
	}
	ui.cacheRW.RUnlock()
}

func (ui *UI) ListenAndServe(address string) error {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Get("/", ui.HandleGet)
	return http.ListenAndServe(address, r)
}

type cellStatus struct {
	cell   *Cell
	status string
}

// cellStatuses is used sort cellStatus types via the sort interface
type cellStatuses []*cellStatus

func (cs cellStatuses) Len() int {
	return len(cs)
}

func (cs cellStatuses) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}

func (cs cellStatuses) Less(i, j int) bool {
	iX, iY := cs[i].cell.x, cs[i].cell.y
	jX, jY := cs[j].cell.x, cs[j].cell.y
	return iX < jX || (iX == jX && iY < jY)
}

func StatusGrid() []byte {
	var wg sync.WaitGroup
	services := Consul.ServiceCatalog()
	cellStatCh := make(chan *cellStatus, MaxHeight*MaxWidth)
	for x := 1; x <= MaxWidth; x++ {
		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			for y := 1; y <= MaxHeight; y++ {
				c := &Cell{x: x, y: y}
				var exists bool
				for name := range services {
					if name == c.Name() {
						exists = true
						break
					}
				}
				val := "🌑"
				if exists {
					if c.Alive() {
						val = "🟢"
					} else {
						val = "⭕️"
					}
				}
				cellStatCh <- &cellStatus{
					cell:   c,
					status: val,
				}
			}
		}(x)
	}
	wg.Wait()
	close(cellStatCh)

	cellStats := make(cellStatuses, 0, MaxHeight*MaxWidth)
	for cs := range cellStatCh {
		cellStats = append(cellStats, cs)
	}
	sort.Sort(cellStats)
	var out bytes.Buffer
	var count int
	for _, cs := range cellStats {
		out.WriteString(cs.status)
		count++
		if count%MaxWidth == 0 {
			out.WriteString("\n")
		}
	}
	return out.Bytes()
}
