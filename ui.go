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
var TickTime int

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
	w.Write([]byte("set next pattern:" + p + "\n"))
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
				// ui.logger.Info("GET", name, "alive:", alive)
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

func (ui *UI) HandleRaw(w http.ResponseWriter, r *http.Request) {
	ui.cacheRW.RLock()
	defer ui.cacheRW.RUnlock()
	w.Write([]byte(Grid))
}

func (ui *UI) HandleBrowser(w http.ResponseWriter, r *http.Request) {
	// msg := "<html><head><style>body {background-color: #000;}</style><meta http-equiv=\"refresh\" content=\"0.1\" /><body>\n"
	// msg := strings.ReplaceAll(Grid, "\n", "<br />\n")
	// msg += "\n</body></head></html>"
	msg := fmt.Sprintf(`<!doctype html>
<html>
	<head>
	  <title>HashiCorp's Game of Life</title>
	  <meta charset="utf-8">
	  <style>
	    body {background-color: #000}
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
