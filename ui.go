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
	cellsRW sync.RWMutex
	cells   map[string]*Cell

	cacheRW    sync.RWMutex
	cachedGrid []byte

	logger hclog.Logger
}

func NewUI(logger hclog.Logger, refreshRate time.Duration) (*UI, error) {
	return &UI{
		logger: logger.Named("ui"),
	}, nil
}

func (ui *UI) HandleGet(w http.ResponseWriter, r *http.Request) {
	ui.cacheRW.RLock()
	w.Write(ui.cachedGrid)
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
	cellStatCh := make(chan *cellStatus, 4)
	for y := 1; y <= MaxHeight; y++ {
		for x := 1; x <= MaxWidth; x++ {
			wg.Add(1)
			c := &Cell{x: x, y: y}
			go func(c *Cell) {
				defer wg.Done()
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
			}(c)

		}
		wg.Wait()
		close(cellStatCh)
	}

	cellStats := make(cellStatuses, 0, MaxHeight*MaxWidth)
	for cs := range cellStatCh {
		cellStats = append(cellStats, cs)
	}
	sort.Sort(cellStats)
	var out bytes.Buffer
	for _, cs := range cellStats {
		out.WriteString(cs.status)
	}
	out.WriteString("\n")
	return out.Bytes()
}
