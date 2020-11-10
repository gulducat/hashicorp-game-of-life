package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

func NewConsul() *ConsulAPI {
	addr := os.Getenv("CONSUL_HTTP_ADDR")
	if addr == "" {
		addr = "http://localhost:8500"
	}
	return &ConsulAPI{
		api: &API{
			BaseUrl: fmt.Sprintf("%s/v1", addr),
		},
	}
}

type ConsulAPI struct {
	api *API
}

type ServiceHealth []struct {
	Checks []struct {
		Status string `json:"Status"`
	} `json:"Checks"`
}

func (c *ConsulAPI) ServiceHealth(name string) bool {
	var health ServiceHealth
	path := fmt.Sprintf("/health/service/%s", name)
	code, body := c.api.Get(path)
	if code != 200 {
		return false
	}
	json.Unmarshal(body, &health)

	if len(health) < 1 {
		return false
	}
	if len(health[0].Checks) < 1 {
		return false
	}
	return health[0].Checks[0].Status == "passing"
}

func (c *ConsulAPI) ServiceCatalog() map[string][]string {
	var catalog map[string][]string
	_, body := c.api.Get("/catalog/services")
	json.Unmarshal(body, &catalog)
	return catalog
}

func (c *ConsulAPI) ServiceExists(name string) bool {
	catalog := c.ServiceCatalog()
	for svc, _ := range catalog {
		if svc == name {
			return true
		}
	}
	return false
}

type ConsulKV []struct {
	Value string `json:"Value"`
}

func (c *ConsulAPI) GetKV(name string) string {
	var resp ConsulKV
	path := fmt.Sprintf("/kv/%s", name)
	code, body := c.api.Get(path)
	if code != 200 {
		return ""
	}
	json.Unmarshal(body, &resp)
	if len(resp) < 1 {
		return ""
	}
	b64value := resp[0].Value
	decoded, err := base64.StdEncoding.DecodeString(b64value)
	if err != nil {
		panic(err)
	}
	return string(decoded)
}

func (c *ConsulAPI) SetKV(name string, value string) bool {
	path := fmt.Sprintf("/kv/%s", name)
	code, _ := c.api.Put(path, []byte(value))
	if code != 200 {
		return false
	}
	return true
}
