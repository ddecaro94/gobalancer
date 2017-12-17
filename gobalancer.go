package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type proxy struct {
	Backend *Backend
}

type frontend struct {
	Name    string
	Listen  string
	TLS     bool
	Pool    string
	Bounce  []int
	Logfile string
}

type server struct {
	Name   string
	Scheme string
	Host   string
	Port   int
}
type pool struct {
	Name    string   `json:"name"`
	Servers []server `json:"servers"`
}

//Config for the main program
type Config struct {
	Frontends []frontend `json:"frontends"`
	Pools     []pool     `json:"pools"`
}

var tr, client = &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    300 * time.Second,
	DisableCompression: true,
}, &http.Client{Transport: tr, Timeout: 60 * time.Second}

func (p proxy) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error
	iter, repeat, ttl, path := 0, true, p.Backend.hosts.Len(), req.RequestURI

	for repeat && iter < ttl {
		iter++
		req.URL, err = url.Parse(p.Backend.Next())
		req.URL.Path = path
		req.RequestURI = ""
		req.Host = ""
		res, httperr := client.Do(req)
		switch {

		case httperr != nil:
			repeat = true
			fmt.Printf("Calling %s %s, error: %s - redirecting\n", req.URL.Host, req.URL.Path, httperr.Error())
			if iter < ttl {
				panic("No valid host found for service " + req.URL.Path)
			}
		case res.StatusCode == 404:
			defer res.Body.Close()
			fmt.Printf("Calling %s %s, received %d\n", req.URL.Host, req.URL.Path, res.StatusCode)
			repeat = true
		default:
			defer res.Body.Close()
			fmt.Printf("Calling %s %s, received %d\n", req.URL.Host, req.URL.Path, res.StatusCode)
			repeat = false
			for name, header := range res.Header {
				for _, val := range header {
					resp.Header().Set(name, val)
				}
			}
			_, err = io.Copy(resp, res.Body)
			if err != nil {
				panic(err)
			}

		}

	}
}

func main() {

	var config Config
	file, e := ioutil.ReadFile("./config.json")
	if e != nil {
		panic(e)
	}

	if err := json.Unmarshal(file, &config); err != nil {
		panic(err)
	}

	b, err := NewBackend("roundrobin", "http://www.amazon.it:80", "http://www.facebook.com:80")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", config)
	http.Handle("/", proxy{b})
	http.ListenAndServe(":9000", nil)
}
