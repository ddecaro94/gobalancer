package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/ddecaro94/gobalancer/api"
)

type proxy struct {
	index    int
	mutex    *sync.Mutex
	Frontend *Frontend
	Cluster  *Cluster
}

var tr, client = &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    300 * time.Second,
	DisableCompression: true,
}, &http.Client{Transport: tr, Timeout: 60 * time.Second}

func (p *proxy) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	iter, repeat, ttl, path := 0, true, p.Cluster.Size, req.RequestURI

	for repeat && iter < ttl {
		iter++
		server := p.Next()
		req.URL.Host = server.Host
		req.URL.Scheme = server.Scheme
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
				return
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
	servers := make(map[string]*http.Server)

	if err != nil {
		panic(err)
	}
	for _, frontend := range config.Frontends {
		go func() {
			cluster := config.Clusters[frontend.Pool]
			servers[frontend.Name] = &http.Server{Addr: frontend.Listen, Handler: &proxy{0, &sync.Mutex{}, &frontend, &cluster}}
			err := servers[frontend.Name].ListenAndServe()
			if err != nil {
				panic(err)
			} else {
				fmt.Printf("Listening on %s, frontend %+v, cluster %+v", frontend.Listen, frontend, cluster)
			}
		}()
	}
	//frontend, cluster := config.Frontends["main"], config.Clusters["pool1"]
	//b := NewBackend(cluster)
	//fmt.Printf("%+v", config)

	http.Handle("/", api.APIServer{config})
	http.ListenAndServe(":9898", nil)
}

func codeToBounce(code int, list []int) bool {
	for _, bounced := range list {
		if bounced == code {
			return true
		}
	}
	return false
}

//Next returns the next address according to the balancing algorithm
func (p *proxy) Next() (server Server) {
	p.mutex.Lock()
	switch p.Cluster.Algorithm {
	case "roundrobin":
		p.index = (p.index + 1) % p.Cluster.Size
	}
	defer p.mutex.Unlock()
	return p.Cluster.Servers[p.index]
}
