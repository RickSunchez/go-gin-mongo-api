package main

import (
	"bytes"
	"fmt"
	"gin-server/internal/errors"
	"gin-server/internal/provider"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

var pr provider.Hosts
var httpErr errors.HTTPErrors

const (
	addr string = "localhost:8080"
)

func main() {
	pr = provider.NewProvider()

	pr.Add("http://localhost:8000")
	pr.Add("http://localhost:9000")

	http.HandleFunc("/", handler)

	var httpServerError = make(chan error)
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		httpServerError <- http.ListenAndServe(addr, nil)
	}()

	select {
	case err := <-httpServerError:
		log.Fatalln(err)
	default:
		fmt.Printf("served on: http://%s\n", addr)
	}

	wg.Wait()
}

func handler(w http.ResponseWriter, r *http.Request) {
	reqBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
		return
	}

	defer r.Body.Close()

	host := pr.GetHost() + r.URL.Path

	fmt.Println(host)

	resp, err := proxyRedirect(w, r.Method, host, reqBytes)

	if err != nil {
		httpErr.InternalError(w, err)
		log.Fatalln(err)
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		httpErr.InternalError(w, err)
		log.Fatalln(err)

		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

func proxyRedirect(w http.ResponseWriter, method, host string, body []byte) (*http.Response, error) {
	fmt.Printf("Redirect to %s\n", host)

	var resp *http.Response
	var err error

	client := &http.Client{}

	req, err := http.NewRequest(method, host, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
