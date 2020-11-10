package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type API struct {
	BaseUrl string
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

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Println(resp.StatusCode, method, path) //string(body))
	return resp.StatusCode, body
}
