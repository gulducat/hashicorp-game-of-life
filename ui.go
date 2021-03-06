package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func ApiListen() {
	ui := NewUI()
	logger.Info("api listening", "port", httpPort)
	if err := ui.ListenAndServe(":" + httpPort); err != nil {
		logger.Error("ListenAndServe", "err", err)
		os.Exit(1)
	}
}

type UI struct {
	cacheRW    sync.RWMutex
	cachedGrid []byte
}

func NewUI() *UI {
	return &UI{}
}

func (ui *UI) ListenAndServe(address string) error {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.ThrottleBacklog(8, 8, 1*time.Second))
	r.Get("/", ui.HandleBrowser)
	r.Get("/raw", ui.HandleRaw)
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
	head := "<html><head><meta http-equiv='refresh' content='0; url=/' /></head><body>\n"
	foot := "\n<br></body></html>\n"
	w.Write([]byte(head + "set next pattern: " + p + foot))
}

func (ui *UI) HandleRaw(w http.ResponseWriter, r *http.Request) {
	ui.cacheRW.RLock()
	defer ui.cacheRW.RUnlock()
	w.Write([]byte(Grid))
}

func (ui *UI) HandleBrowser(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf(`<!doctype html>
<html>
	<head>
	  <title>HashiCorp's Game of Life</title>
	  <meta charset="utf-8">
	  <style>
	    html {background-color: #000}
	  </style>
	</head>
	<body>
	  <div id="grid"></div>
	  <script>
		function fetch(){
		  var xhr = new XMLHttpRequest();
		  xhr.open("GET", "/raw");
		  xhr.onload = function () {
			if (this.status==200) {
			  document.getElementById("grid").innerHTML = this.response.replaceAll("\n", "<br>\n");
			} else {
			  console.log(this.status);
			  console.log(this.response);
			}
		  };
		  xhr.send();
		}
		setInterval(fetch, %d);
	  </script>
	</body>
</html>
`, TickTime/2)
	w.Write([]byte(msg))
}

func (ui *UI) UpdateGrid() {
	var val string
	var name string
	Mut.RLock()
	defer Mut.RUnlock()
	ui.cacheRW.Lock()
	defer ui.cacheRW.Unlock()
	Grid = ""
	for y := 1; y <= MaxHeight; y++ {
		for x := 1; x <= MaxWidth; x++ {
			val = "🥶"
			name = fmt.Sprintf("%d-%d", x, y)
			alive, ok := Statuses[name]
			if ok {
				if alive {
					val = "🟢"
				} else {
					val = "🌑"
				}
			}
			Grid += val
		}
		Grid += "\n"
	}
}
