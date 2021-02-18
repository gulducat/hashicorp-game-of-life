package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type HTTP struct {
	BaseUrl string
	client  *http.Client
}

func NewHTTP(baseUrl string) *HTTP {
	return &HTTP{
		BaseUrl: baseUrl,
		client:  http.DefaultClient,
	}
}

func (a *HTTP) Get(path string) (int, []byte) {
	return a.RetryRequest("GET", path, []byte{})
}

func (a *HTTP) Post(path string, data []byte) (int, []byte) {
	return a.RetryRequest("POST", path, data)
}

func (a *HTTP) Put(path string, data []byte) (int, []byte) {
	return a.RetryRequest("PUT", path, data)
}

func (a *HTTP) Delete(path string) (int, []byte) {
	return a.RetryRequest("DELETE", path, []byte{})
}

func (a *HTTP) RetryRequest(method string, path string, data []byte) (int, []byte) {
	for x := 0; x < 5; x++ {
		i, b, err := a.Request(method, path, data)
		if err == nil {
			return i, b
		}
		minSleep := x * 200
		sleep := time.Duration(minSleep + rand.Intn(500))
		time.Sleep(sleep * time.Millisecond)
	}
	return 0, nil
}

func (a *HTTP) Request(method string, path string, data []byte) (int, []byte, error) {
	url := fmt.Sprintf("%s%s", a.BaseUrl, path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		logger.Error("http.NewRequest",
			"method", method,
			"path", path,
			"err", err)
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		logger.Error("client.Do",
			"method", method,
			"path", path,
			"err", err)
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	logger.Info("submitted http request",
		"method", method,
		"url", url,
		"status_code", resp.StatusCode)
	return resp.StatusCode, body, nil
}
