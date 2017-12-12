package main

import (
	"fmt"
	"time" 
	"net/http"
	"io/ioutil"
	"container/ring"
)
var tr, client = &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    30 * time.Second,
	DisableCompression: true,
}, &http.Client{Transport: tr}

var hosts *ring.Ring

func main() {

	http.HandleFunc("/", proxy)
	http.ListenAndServe(":9000", nil)
	
	proxied := []string{"http://www.corriere.it/", "http://www.repubblica.it/"}
	hosts = ring.New(len(proxied))
	for index := 0; index < hosts.Len(); index++ {
		hosts.Value = proxied[index]
		hosts = hosts.Next()
	}

}

func proxy (resp http.ResponseWriter, req *http.Request) {

	request, err := http.NewRequest(req.Method, getNextHost(hosts) +req.RequestURI, req.Body)
	request.Header = req.Header

	time.Sleep(10);

	res, err := client.Do(request)
	if err != nil {
        panic(err)
    }
	fmt.Printf("%s\n%#v\n", req.URL, res)
	resdata, err := ioutil.ReadAll(res.Body)
	for name, headers := range res.Header {
		for _, h := range headers {
			resp.Header().Set(name, h)
		}
	}
	resp.Write(resdata)
}

func getNextHost(r *ring.Ring)  (host string){
	r = r.Next()
	return r.Value.(string)
}