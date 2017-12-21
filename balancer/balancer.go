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
)

var tr, client = &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    300 * time.Second,
	DisableCompression: true,
}, &http.Client{Transport: tr, Timeout: 60 * time.Second}

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
	iter, repeat, ttl, path := 0, true, p.conf.Clusters[p.conf.Frontends[p.frontend].Pool].Size, req.RequestURI

	//read all request body because making a client request closes the body
	bodyBytes, _ := ioutil.ReadAll(req.Body)
	req.Body.Close()

	for repeat && iter < ttl {
		iter++
		server := p.Next()
		req.URL.Host = server.Host
		req.URL.Scheme = server.Scheme
		req.URL.Path = path
		req.RequestURI = ""
		req.Host = ""
		req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		res, httperr := client.Do(req)

		switch {

		case httperr != nil:
			repeat = true
			fmt.Printf("Calling %s %s, error: %s - redirecting\n", req.URL.Host, req.URL.Path, httperr.Error())
			if iter >= ttl {
				fmt.Printf("No valid host found for service: %s \n", req.URL.Path)
				http.Error(resp, "Service Unavailable", 503)
			}
		case codeToBounce(res.StatusCode, p.conf.Frontends[p.frontend].Bounce):
			fmt.Printf("Calling %s %s, received %d\n", req.URL.Host, req.URL.Path, res.StatusCode)
			repeat = true
			if iter >= ttl {
				forward(resp, res)
				defer res.Body.Close()
			}
		default:
			fmt.Printf("Calling %s %s, received %d\n", req.URL.Host, req.URL.Path, res.StatusCode)
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
