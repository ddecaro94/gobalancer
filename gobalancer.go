package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type proxy struct {
	Frontend *Frontend
	Backend  *Backend
}

var tr, client = &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    300 * time.Second,
	DisableCompression: true,
}, &http.Client{Transport: tr, Timeout: 60 * time.Second}

func (p proxy) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	iter, repeat, ttl, path := 0, true, p.Backend.len, req.RequestURI

	for repeat && iter < ttl {
		iter++
		req.URL = p.Backend.Next()
		req.URL.Path = path
		req.RequestURI = ""
		req.Host = ""
		res, httperr := client.Do(req)
		switch {

		case httperr != nil:
			repeat = true
			fmt.Printf("Calling %s %s, error: %s - redirecting\n", req.URL.Host, req.URL.Path, httperr.Error())
			if iter >= ttl {
				fmt.Printf("No valid host found for service: %s ", req.URL.Path)
				http.Error(resp, "Service Unavailable", 503)
			}
		case codeToBounce(res.StatusCode, p.Frontend.Bounce):
			defer res.Body.Close()
			fmt.Printf("Calling %s %s, received %d\n", req.URL.Host, req.URL.Path, res.StatusCode)
			repeat = true
			if iter < ttl {
				forward(resp, res)
			}
		default:
			defer res.Body.Close()
			fmt.Printf("Calling %s %s, received %d\n", req.URL.Host, req.URL.Path, res.StatusCode)
			repeat = false
			forward(resp, res)
		}
	}
}

func forward(w http.ResponseWriter, res *http.Response) {
	for name, header := range res.Header {
		for _, val := range header {
			w.Header().Set(name, val)
		}
	}
	_, err := io.Copy(w, res.Body)
	if err != nil {
		panic(err)
	}
}

func main() {

	config, err := ReadConfig("./config.json")
	if err != nil {
		panic(err)
	}
	frontend, cluster := config.Frontends["main"], config.Clusters["pool1"]
	b := NewBackend(cluster)
	fmt.Printf("%+v", config)

	http.ListenAndServe(":9000", proxy{&frontend, b})
}

func codeToBounce(code int, list []int) bool {
	for _, bounced := range list {
		if bounced == code {
			return true
		}
	}
	return false
}
