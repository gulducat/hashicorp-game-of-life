package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
)

func Coords(name string) (int, int) {
	// given "1-1", return: 1, 1
	bits := strings.Split(name, "-")
	x, _ := strconv.Atoi(bits[0])
	y, _ := strconv.Atoi(bits[1])
	return x, y
}

type CellStatus struct {
	rw     *sync.RWMutex
	x      int
	y      int
	name   string
	status Status
}

func NewCellStatus(x, y int) *CellStatus {
	return &CellStatus{
		rw: new(sync.RWMutex),
		x:  x,
		y:  y,
	}
}

func (c *CellStatus) Name() string {
	if c.name == "" {
		c.name = fmt.Sprintf("%d-%d", c.x, c.y)
	}
	return c.name
}

func (c *CellStatus) Neighbors(maxX int, maxY int) map[string]*CellStatus {
	neighbors := make(map[string]*CellStatus)
	for x := c.x - 1; x <= c.x+1; x++ {
		for y := c.y - 1; y <= c.y+1; y++ {
			if x < 1 || y < 1 || x > maxX || y > maxY {
				continue
			}
			if x == c.x && y == c.y {
				continue
			}
			neighbor := &CellStatus{
				x: x,
				y: y,
			}
			neighbors[neighbor.Name()] = neighbor
		}
	}
	return neighbors
}

func (c *CellStatus) GetStatus() Status {
	var status Status
	consulStatus := Consul.GetKV(c.Name())
	switch consulStatus {
	case "alive":
		status = Alive
	case "dead":
		status = Dead
	default:
		status = Nonexistent
	}
	c.rw.Lock()
	c.status = status
	c.rw.Unlock()
	return status
}

type CellRunner struct {
	*CellStatus
	neighbors map[string]*CellStatus
	logger    hclog.Logger
}

func NewCellRunner(x, y int, logger hclog.Logger) *CellRunner {
	name := fmt.Sprintf("%d-%d", x, y)
	return &CellRunner{
		CellStatus: NewCellStatus(x, y),
		logger:     logger.Named(name),
	}
}

func (c *CellStatus) GetJobspec() NomadJob {
	var job NomadJob
	spec := strings.Replace(DefaultJob, "0-0", c.Name(), -1)
	json.Unmarshal([]byte(spec), &job)
	return job
}

func (c *CellRunner) Exists() bool {
	return Consul.ServiceExists(c.Name()) // maybe dumb to check the whole catalog...
}

func (c *CellRunner) Alive() bool {
	healthy := Consul.ServiceHealth(c.Name())
	c.logger.Info(c.Name() + " is healthy:")
	return healthy
}

func (c *CellRunner) TmpFile() string {
	return fmt.Sprintf("%s/%s", TmpDir, c.Name())
}

func (c *CellRunner) SetStatus(alive bool) {
	status := "alive"
	if !alive {
		status = "dead"
	}
	Consul.SetKV(c.Name(), status)
	// _ = os.Mkdir(TmpDir, 0755)
	// err := ioutil.WriteFile(c.TmpFile(), []byte(status), 0644)
	// if err != nil {
	// 	panic(err)
	// }
}

func (c *CellRunner) Destroy() {
	Nomad.DeleteJob(c)
	Consul.DeleteKV(c.Name())
}
