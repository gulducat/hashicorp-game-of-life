package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type NomadJob struct {
	Job struct {
		ID          string   `json:"ID"`
		Name        string   `json:"Name"`
		Type        string   `json:"Type"`
		Datacenters []string `json:"Datacenters"`
		TaskGroups  []struct {
			Name  string `json:"Name"`
			Count int    `json:"Count"`
			Tasks []struct {
				Name   string `json:"Name"`
				Driver string `json:"Driver"`
				Config struct {
					Command string   `json:"command"`
					Args    []string `json:"args"`
				} `json:"Config"`
				Env      interface{} `json:"Env"`
				Services []struct {
					Name   string   `json:"Name"`
					Tags   []string `json:"Tags"`
					Checks []struct {
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

var DefaultJob = `{
	"Job": {
	  "ID": "0-0",
	  "Name": "0-0",
	  "Type": "service",
	  "Datacenters": ["dc1"],
	  "TaskGroups": [
		{
		  "Name": "cell",
		  "Count": 1,
		  "Tasks": [
			{
			  "Name": "cell",
			  "Driver": "raw_exec",
			  "Config": {
				"command": "/Users/danielbennett/git/gulducat/hashicorp-game-of-life/hashicorp-game-of-life",
				"args": ["run"]
			  },
			  "Env": null,
			  "Services": [
				{
				  "Name": "0-0",
				  "Tags": ["cell"],
				  "Checks": [
					{
					  "Name": "check",
					  "Type": "script",
					  "Command": "/Users/danielbennett/git/gulducat/hashicorp-game-of-life/hashicorp-game-of-life",
					  "Args": ["check"],
					  "Interval": 10000000000,
					  "Timeout": 20000000000,
					  "InitialStatus": "passing"
					}
				  ]
				}
			  ]
			}
		  ]
		}
	  ]
	}
  }`

func NewNomadJob(cell *Cell) NomadJob {
	var job NomadJob
	spec := strings.Replace(DefaultJob, "0-0", cell.Name(), -1)
	json.Unmarshal([]byte(spec), &job)
	return job
}

func NewNomad() *Nomad {
	addr := os.Getenv("NOMAD_ADDR")
	if addr == "" {
		addr = "http://localhost:4646"
	}
	return &Nomad{
		BaseUrl: fmt.Sprintf("%s/v1", addr),
	}
}

type Nomad struct {
	BaseUrl string
}

func (n *Nomad) Request(method string, path string, data []byte) (int, string) {
	url := fmt.Sprintf("%s%s", n.BaseUrl, path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func CreateJob(job *NomadJob) {
	n := NewNomad()
	spec, err := json.Marshal(job)
	if err != nil {
		panic(err)
	}
	status, body := n.Request("POST", "/jobs", spec)
	fmt.Println(status, body)
}
