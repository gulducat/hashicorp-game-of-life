package main

// TODO: un-hard-code the bin path (hashicorp-game-of-life)
// TODO: disable rescheduling for seed job? https://www.nomadproject.io/docs/job-specification/reschedule#disabling-rescheduling

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
	api := NewAPI(addr, logger)
	return &NomadAPI{
		api: api,
	}
}

type NomadAPI struct {
	api *API
}

func (n *NomadAPI) CreateJob(cell *Cell2) {
	job := cell.GetJobspec()
	spec, err := json.Marshal(job)
	if err != nil {
		panic(err)
	}
	status, body := n.api.Post("/v1/jobs", spec)
	fmt.Println(status, string(body))
}

func (n *NomadAPI) DeleteJob(cell *Cell2) {
	path := fmt.Sprintf("/v1/job/%s?purge=true", cell.Name())
	// path := fmt.Sprintf("/job/%s", cell.Name())
	n.api.Delete(path)
	// status, body := n.api.Delete(path)
	// log.Printf("DELETE %s status: %d body: %s\n", path, status, body)
}

// type NomadJob struct {
// 	Job struct {
// 		ID          string   `json:"ID"`
// 		Name        string   `json:"Name"`
// 		Type        string   `json:"Type"`
// 		Datacenters []string `json:"Datacenters"`
// 		TaskGroups  []struct {
// 			Name          string `json:"Name"`
// 			Count         int    `json:"Count"`
// 			EphemeralDisk struct {
// 				SizeMB int `json:"SizeMB"`
// 			} `json:"EphemeralDisk"`
// 			Networks []struct {
// 				DynamicPorts []struct {
// 					Label string `json:"Label"`
// 				} `json:"DynamicPorts"`
// 			} `json:"Networks"`
// 			Services []struct {
// 				Name      string `json:"Name"`
// 				PortLabel string `json:"PortLabel"`
// 			} `json:"Services"`
// 			RestartPolicy struct {
// 				Attempts int `json:"Attempts"`
// 				Delay    int `json:"Delay"`
// 			} `json:"RestartPolicy"`
// 			Tasks []struct {
// 				Name   string `json:"Name"`
// 				Driver string `json:"Driver"`
// 				Config struct {
// 					Command string   `json:"command"`
// 					Args    []string `json:"args"`
// 				} `json:"Config"`
// 				Env       interface{} `json:"Env"`
// 				Resources struct {
// 					CPU      int `json:"CPU"`
// 					MemoryMB int `json:"MemoryMB"`
// 					DiskMB   int `json:"DiskMB"`
// 				} `json:"Resources"`
// 				LogConfig struct {
// 					MaxFiles      int `json:"MaxFiles"`
// 					MaxFileSizeMB int `json:"MaxFileSizeMB"`
// 				} `json:"LogConfig"`
// 			} `json:"Tasks"`
// 		} `json:"TaskGroups"`
// 	} `json:"Job"`
// }

// var DefaultJob = fmt.Sprintf(`{
// 	"Job": {
// 	  "ID": "0-0",
// 	  "Name": "0-0",
// 	  "Type": "service",
// 	  "Datacenters": ["dc1"],
// 	  "TaskGroups": [{
// 		  "Name": "cell",
// 		  "Count": 1,
// 		  "EphemeralDisk": {
// 			"SizeMB": 10
// 		  },
// 		  "Networks": [{
// 			"DynamicPorts": [{
// 			  "Label": "udp"
// 			}]
// 		  }],
// 		  "Services": [{
// 			"Name": "0-0",
// 			"PortLabel": "udp"
// 		  }],
// 		  "RestartPolicy": {
// 			"Attempts": 5,
// 			"Delay": 2000000000
// 		  },
// 		  "Tasks": [{
// 			  "Name": "cell",
// 			  "Driver": "raw_exec",
// 			  "Config": {
// 				"command": "hashicorp-game-of-life",
// 				"args": ["run"]
// 			  },
// 			  "Env": {
// 				  "CONSUL_HTTP_ADDR": "http://localhost:8500"
// 			  },
// 			  "Resources": {
// 				"CPU": 160,
// 				"MemoryMB": 35,
// 				"DiskMB": 10
// 			  },
// 			  "LogConfig": {
// 				"MaxFiles": 2,
// 				"MaxFileSizeMB": 2
// 			  }
// 			}]
// 		}]
// 	}
//   }`)

// TODO: slacklink
// poststart hook "consul_services" failed: unable to get address for service "1-5": invalid port "udp": port label not found
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
			Networks []struct {
				DynamicPorts []struct {
					Label string `json:"Label"`
				} `json:"DynamicPorts"`
			} `json:"Networks"`
			RestartPolicy struct {
				Attempts int `json:"Attempts"`
				Delay    int `json:"Delay"`
			} `json:"RestartPolicy"`
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
				} `json:"Resources"`
				Services []struct {
					Name      string `json:"Name"`
					PortLabel string `json:"PortLabel"`
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
		  "Networks": [{
			"DynamicPorts": [{
				"Label": "udp"
			}]
		  }],
		  "RestartPolicy": {
			"Attempts": 5,
			"Delay": 2000000000
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
				"DiskMB": 10
			  },
			  "Services": [{
				  "Name": "0-0",
				  "PortLabel": "udp"
			  }]
			}]
		}]
	}
  }`)
