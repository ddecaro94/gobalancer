package balancer

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/ddecaro94/gobalancer/config"
	"github.com/google/uuid"
)

var tr, client = &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    300 * time.Second,
	DisableCompression: true,
}, &http.Client{Transport: tr, Timeout: 600 * time.Second}

//A Balancer stores structures used to compute the host to be used at runtime
type Balancer struct {
	index    int
	mutex    *sync.Mutex
	conf     *config.Config
	frontend string
}

//New returns a Balancer instance
func New(c *config.Config, frontend string) (p *Balancer) {
	b := &Balancer{0, &sync.Mutex{}, c, frontend}
	return b
}

func (p *Balancer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	iter, repeat, ttl, path, body := 0, true, p.conf.Clusters[p.conf.Frontends[p.frontend].Pool].Size, req.RequestURI, []byte{}
	reqID, err := uuid.NewUUID()
	forbidden := make(map[string]bool)
	bouncedCodes := p.conf.Frontends[p.frontend].Bounce
	if err != nil {
		panic(err)
	}
	if len(bouncedCodes) != 0 {
		//read all request body because making a client request closes the body
		body, _ = ioutil.ReadAll(req.Body)
		req.Body.Close()
	}
	for repeat && iter < ttl {
		iter++
		server := p.Next()
		host := server.Host + ":" + server.Port
		for forbidden[host] == true {
			server := p.Next()
			host = server.Host + ":" + server.Port
		}
		req.URL.Host = host
		req.URL.Scheme = server.Scheme
		req.URL.Path = path
		req.RequestURI = ""
		req.Host = ""
		if len(bouncedCodes) != 0 {
			req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}
		res, httperr := client.Do(req)

		switch {

		case httperr != nil:
			fmt.Printf("%s - Calling %s %s, error: %s - redirecting\n", reqID, req.URL.Host, req.URL.Path, httperr.Error())
			repeat = true
			//add url to forbidden
			forbidden[host] = true
			if iter >= ttl {
				fmt.Printf("%s - No valid host found for service: %s \n", reqID, req.URL.Path)
				http.Error(resp, "Service Unavailable", 503)
			}
		case codeToBounce(res.StatusCode, p.conf.Frontends[p.frontend].Bounce):
			fmt.Printf("%s - Calling %s %s, received %d\n", reqID, req.URL.Host, req.URL.Path, res.StatusCode)
			repeat = true
			forbidden[host] = true
			if iter >= ttl {
				forward(resp, res)
				defer res.Body.Close()
			}
		default:
			fmt.Printf("%s - Calling %s %s, received %d\n", reqID, req.URL.Host, req.URL.Path, res.StatusCode)
			repeat = false
			forward(resp, res)
			defer res.Body.Close()
		}
	}
}

//Next returns the next address according to the balancing algorithm
func (p *Balancer) Next() (server config.Server) {
	cluster := p.conf.Clusters[p.conf.Frontends[p.frontend].Pool]
	p.mutex.Lock()
	switch cluster.Algorithm {
	case "roundrobin":
		p.index = (p.index + 1) % cluster.Size
	}
	defer p.mutex.Unlock()
	return cluster.Servers[p.index]
}

func codeToBounce(code int, list []int) bool {
	for _, bounced := range list {
		if bounced == code {
			return true
		}
	}
	return false
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
