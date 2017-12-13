package main

import (
	"container/ring"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

var tr, client = &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    300 * time.Second,
	DisableCompression: true,
}, &http.Client{Transport: tr}

var hosts *ring.Ring

func main() {
	runtime.GOMAXPROCS(2)
	proxied := []string{"http://ibmcollib01:7800", "http://ibmcollib02:7800"}
	hosts = ring.New(len(proxied))
	for index := 0; index < hosts.Len(); index++ {
		hosts.Value = proxied[index]
		hosts = hosts.Next()
	}

	http.HandleFunc("/", proxy)
	http.ListenAndServe(":9000", nil)

}

func proxy(resp http.ResponseWriter, req *http.Request) {
	iter := 0
	ttl := hosts.Len()

	req.Host = ""
	u, err := url.Parse(getNextHost())
	req.URL = u
	req.URL.Path = req.RequestURI
	req.RequestURI = ""
	res, err := client.Do(req)
	iter++
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s %s %d\n", req.URL.Host, req.URL.Path, res.StatusCode)
	for res.StatusCode == 404 && iter < ttl {
		res, err = client.Do(req)
		iter++
	}
	resdata, err := ioutil.ReadAll(res.Body)
	for name, header := range res.Header {
		for _, val := range header {
			resp.Header().Set(name, val)
		}
	}
	resp.Write(resdata)
}

func getNextHost() (host string) {
	hosts = hosts.Next()
	return hosts.Value.(string)
}
