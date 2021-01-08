package main

import (
	"strconv"
	"strings"
)

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"strconv"
// 	"strings"
// )

// func NewCell(name string) Cell {
// 	x, y := Coords(name)
// 	return Cell{x: x, y: y}
// }

func Coords(name string) (int, int) {
	// given "1-1", return: 1, 1
	bits := strings.Split(name, "-")
	x, _ := strconv.Atoi(bits[0])
	y, _ := strconv.Atoi(bits[1])
	return x, y
}

// type Cell struct {
// 	x      int
// 	y      int
// 	status bool
// 	n      map[string]*Cell
// }

// func (c *Cell) Name() string {
// 	return fmt.Sprintf("%d-%d", c.x, c.y)
// }

// func (c *Cell) Neighbors() map[string]*Cell {
// 	if c.n != nil {
// 		return c.n
// 	}
// 	all := [8]*Cell{
// 		// comments assuming cell "2-2"

// 		// top row
// 		&Cell{x: c.x - 1, y: c.y - 1}, // 1-1
// 		&Cell{x: c.x, y: c.y - 1},     // 2-1
// 		&Cell{x: c.x + 1, y: c.y - 1}, // 3-1

// 		// middle row
// 		&Cell{x: c.x - 1, y: c.y}, // 1-2
// 		// 2-2 is self.
// 		&Cell{x: c.x + 1, y: c.y}, // 3-2

// 		// bottom row
// 		&Cell{x: c.x - 1, y: c.y + 1}, // 1-3
// 		&Cell{x: c.x, y: c.y + 1},     // 2-3
// 		&Cell{x: c.x + 1, y: c.y + 1}, // 3-3

// 	}
// 	var valid = make(map[string]*Cell)
// 	for _, n := range all {
// 		if n.x < 1 || n.y < 1 || n.x > MaxWidth || n.y > MaxHeight {
// 			continue
// 		}
// 		valid[n.Name()] = n
// 		// valid = append(valid, n)
// 	}
// 	c.n = valid
// 	return valid
// }

// func (c *Cell) Create() {
// 	Nomad.CreateJob(c)
// }

// func (c *Cell) GetJobspec() NomadJob {
// 	var job NomadJob
// 	spec := strings.Replace(DefaultJob, "0-0", c.Name(), -1)
// 	json.Unmarshal([]byte(spec), &job)
// 	return job
// }

// func (c *Cell) Exists() bool {
// 	return Consul.ServiceExists(c.Name()) // maybe dumb to check the whole catalog...
// }

// func (c *Cell) Alive() bool {
// 	healthy := Consul.ServiceHealth(c.Name())
// 	log.Println(c.Name(), "healthy:", healthy)
// 	return healthy
// }

// func (c *Cell) TmpFile() string {
// 	return fmt.Sprintf("%s/%s", TmpDir, c.Name())
// }

// func (c *Cell) SetStatus(alive bool) {
// 	status := "alive"
// 	if !alive {
// 		status = "dead"
// 	}
// 	Consul.SetKV(c.Name(), status)
// 	// _ = os.Mkdir(TmpDir, 0755)
// 	// err := ioutil.WriteFile(c.TmpFile(), []byte(status), 0644)
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// }

// func (c *Cell) GetStatus() bool {
// 	return Consul.GetKV(c.Name()) == "alive"

// 	// alive := 0
// 	// for _, n := range c.Neighbors() {
// 	// 	if n.status {
// 	// 		alive++
// 	// 	}
// 	// }

// 	// bts, err := ioutil.ReadFile(c.TmpFile())
// 	// if err != nil {
// 	// 	log.Println("ERR", err)
// 	// 	return false
// 	// }
// 	// return string(bts) == "alive"
// }

// func (c *Cell) Destroy() {
// 	Nomad.DeleteJob(c)
// 	Consul.DeleteKV(c.Name())
// }
