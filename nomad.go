package main

// TODO: un-hard-code the bin path (hashicorp-game-of-life)

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
)

func NewNomad(logger hclog.Logger) *NomadAPI {
	addr := os.Getenv("NOMAD_ADDR")
	if addr == "" {
		addr = "http://localhost:4646"
	}
	api := NewAPI(fmt.Sprintf("%s/v1", addr), logger)
	return &NomadAPI{
		api: api,
	}
}

type NomadAPI struct {
	api *API
}

func (n *NomadAPI) CreateJob(cell *Cell) {
	job := cell.GetJobspec()
	spec, err := json.Marshal(job)
	if err != nil {
		panic(err)
	}
	status, body := n.api.Post("/jobs", spec)
	fmt.Println(status, string(body))
}

func (n *NomadAPI) DeleteJob(cell *Cell) {
	path := fmt.Sprintf("/job/%s?purge=true", cell.Name())
	n.api.Delete(path)
}

type NomadJob struct {
	Job struct {
		ID          string   `json:"ID"`
		Name        string   `json:"Name"`
		Type        string   `json:"Type"`
		Datacenters []string `json:"Datacenters"`
		TaskGroups  []struct {
			Name          string `json:"Name"`
			Count         int    `json:"Count"`
			EphemeralDisk struct {
				SizeMB int `json:"SizeMB"`
			} `json:"EphemeralDisk"`
			Tasks []struct {
				Name   string `json:"Name"`
				Driver string `json:"Driver"`
				Config struct {
					Command string   `json:"command"`
					Args    []string `json:"args"`
				} `json:"Config"`
				Env       interface{} `json:"Env"`
				Resources struct {
					CPU      int `json:"CPU"`
					MemoryMB int `json:"MemoryMB"`
					DiskMB   int `json:"DiskMB"`
					Networks []struct {
						DynamicPorts []struct {
							Label string `json:"Label"`
						} `json:"DynamicPorts"`
					} `json:"Networks"`
				} `json:"Resources"`
				Services []struct {
					Name      string   `json:"Name"`
					PortLabel string   `json:"PortLabel"`
					Tags      []string `json:"Tags"`
					Checks    []struct {
						Name          string   `json:"Name"`
						Type          string   `json:"Type"`
						Command       string   `json:"Command"`
						Args          []string `json:"Args"`
						Interval      int64    `json:"Interval"`
						Timeout       int64    `json:"Timeout"`
						InitialStatus string   `json:"InitialStatus"`
					} `json:"Checks"`
				} `json:"Services"`
			} `json:"Tasks"`
		} `json:"TaskGroups"`
	} `json:"Job"`
}

var DefaultJob = fmt.Sprintf(`{
	"Job": {
	  "ID": "0-0",
	  "Name": "0-0",
	  "Type": "service",
	  "Datacenters": ["dc1"],
	  "TaskGroups": [{
		  "Name": "cell",
		  "Count": 1,
		  "EphemeralDisk": {
			"SizeMB": 150
		  },
		  "Tasks": [{
			  "Name": "cell",
			  "Driver": "raw_exec",
			  "Config": {
				"command": "hashicorp-game-of-life",
				"args": ["run"]
			  },
			  "Env": {
				  "CONSUL_HTTP_ADDR": "http://localhost:8500"
			  },
			  "Resources": {
				"CPU": 60,
				"MemoryMB": 35,
				"DiskMB": 10,
				"Networks": [{
					"DynamicPorts": [{
						"Label": "udp"
					}]
				}]
			  },
			  "Services": [{
				  "Name": "0-0",
				  "PortLabel": "udp",
				  "Checks": [{
					  "Name": "check",
					  "Type": "script",
					  "Command": "hashicorp-game-of-life",
					  "Args": ["check"],
					  "Interval": 1000000000,
					  "Timeout": 2000000000,
					  "InitialStatus": "passing"
					}]
				}]
			}]
		}]
	}
  }`)
