package main

import (
	"container/ring"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"time"
)

var tr, client = &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    300 * time.Second,
	DisableCompression: true,
}, &http.Client{Transport: tr}

var hosts *ring.Ring
var l int
var mutex = &sync.Mutex{}

func main() {
	runtime.GOMAXPROCS(2)
	proxied := []string{"http://ibmcollib01:7800", "http://ibmcollib02:7800"}
	hosts = ring.New(len(proxied))
	for index := 0; index < hosts.Len(); index++ {
		hosts.Value = proxied[index]
		hosts = hosts.Next()
	}
	l = hosts.Len()

	http.HandleFunc("/", proxy)
	http.ListenAndServe(":9000", nil)

}

func proxy(resp http.ResponseWriter, req *http.Request) {
	var err error
	iter, repeat, ttl, path := 0, true, l, req.RequestURI

	for repeat && iter < ttl {
		iter++
		req.URL, err = url.Parse(getNextHost())
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
			fmt.Printf("Calling %s %s, received %d\n", req.URL.Host, req.URL.Path, res.StatusCode)
			repeat = true
		default:
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

func getNextHost() (host string) {
	mutex.Lock()
	hosts = hosts.Next()
	defer mutex.Unlock()
	return hosts.Value.(string)
}
