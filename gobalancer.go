package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var tr, client = &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    300 * time.Second,
	DisableCompression: true,
}, &http.Client{Transport: tr, Timeout: 60 * time.Second}


type proxy struct{
	Backend *Backend
}

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

func main() {
	b := NewBackend("", "http://ibmcollib01:7800", "http://ibmcollib02:7800")
	http.Handle("/", proxy{b})
	http.ListenAndServe(":9000", nil)
}
