package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
)

var ConsulAddr = os.Getenv("CONSUL_HTTP_ADDR")

func NewConsul(logger hclog.Logger) *ConsulAPI {
	if ConsulAddr == "" {
		ConsulAddr = "http://localhost:8500"
	}
	api := NewAPI(fmt.Sprintf("%s/v1", ConsulAddr), logger)
	return &ConsulAPI{
		api: api,
	}
}

type ConsulAPI struct {
	api *API
}

type ConsulService struct {
	Name    string `json:"ServiceName"`
	Address string `json:"Address"`
	Port    int    `json:"ServicePort"`
}

type ConsulServiceResp struct {
	Instances []ConsulService
}

func (c *ConsulAPI) ServiceExists(name string) bool {
	_, err := c.Service(name)
	return err == nil
}

func (c *ConsulAPI) Service(name string) (*ConsulService, error) {
	var svc ConsulServiceResp
	code, body := c.api.Get("/catalog/service/" + name)
	if code != 200 {
		msg := fmt.Sprintf("Error getting service info for %s; code: %d; body: %s", name, code, body)
		log.Println(msg)
		return nil, errors.New(msg)
	}
	err := json.Unmarshal(body, &svc.Instances)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshaling into ConsulService: %s", err)
		log.Println(msg)
		return nil, errors.New(msg)
	}
	if len(svc.Instances) == 0 {
		msg := fmt.Sprintf("Error getting Consul service %q: not found", name)
		log.Println(msg)
		return nil, errors.New(msg)
	}
	return &svc.Instances[0], nil
}

// func (c *ConsulAPI) ServiceExists(name string) bool {
// 	catalog := c.ServiceCatalog()
// 	for svc, _ := range catalog {
// 		if svc == name {
// 			return true
// 		}
// 	}
// 	return false
// }

// type ServiceHealth []struct {
// 	Checks []struct {
// 		Status string `json:"Status"`
// 	} `json:"Checks"`
// }

// func (c *ConsulAPI) ServiceHealth(name string) bool {
// 	var health ServiceHealth
// 	path := fmt.Sprintf("/health/service/%s?stale", name)
// 	code, body := c.api.Get(path)
// 	if code != 200 {
// 		return false
// 	}
// 	json.Unmarshal(body, &health)

// 	if len(health) < 1 {
// 		return false
// 	}
// 	if len(health[0].Checks) < 2 {
// 		return false
// 	}
// 	return health[0].Checks[1].Status == "passing"
// }

// func (c *ConsulAPI) ServiceCatalog() map[string][]string {
// 	var catalog map[string][]string
// 	_, body := c.api.Get("/catalog/services")
// 	json.Unmarshal(body, &catalog)
// 	return catalog
// }

// type ConsulKV []struct {
// 	Value string `json:"Value"`
// }

// func (c *ConsulAPI) GetKV(name string) string {
// 	var resp ConsulKV
// 	path := fmt.Sprintf("/kv/%s", name)
// 	code, body := c.api.Get(path)
// 	if code != 200 {
// 		return ""
// 	}
// 	json.Unmarshal(body, &resp)
// 	if len(resp) < 1 {
// 		return ""
// 	}
// 	b64value := resp[0].Value
// 	decoded, err := base64.StdEncoding.DecodeString(b64value)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return string(decoded)
// }

// func (c *ConsulAPI) SetKV(name string, value string) bool {
// 	path := fmt.Sprintf("/kv/%s", name)
// 	code, _ := c.api.Put(path, []byte(value))
// 	if code != 200 {
// 		return false
// 	}
// 	return true
// }

// func (c *ConsulAPI) DeleteKV(name string) {
// 	path := fmt.Sprintf("/kv/%s", name)
// 	c.api.Delete(path)
// }
