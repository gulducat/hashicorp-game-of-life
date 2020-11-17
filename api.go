package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-hclog"
)

type API struct {
	BaseUrl string
	client  *http.Client
	logger  hclog.Logger
}

func NewAPI(baseUrl string, logger hclog.Logger) *API {
	return &API{
		BaseUrl: baseUrl,
		client:  http.DefaultClient,
		logger:  logger,
	}
}

func (a *API) Get(path string) (int, []byte) {
	return a.Request("GET", path, []byte{})
}

func (a *API) Post(path string, data []byte) (int, []byte) {
	return a.Request("POST", path, data)
}

func (a *API) Put(path string, data []byte) (int, []byte) {
	return a.Request("PUT", path, data)
}

func (a *API) Delete(path string) (int, []byte) {
	return a.Request("DELETE", path, []byte{})
}

func (a *API) Request(method string, path string, data []byte) (int, []byte) {
	url := fmt.Sprintf("%s%s", a.BaseUrl, path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	a.logger.Info("submitted http request",
		"status_code", resp.StatusCode,
		"method", method,
		"path", path)
	return resp.StatusCode, body
}
