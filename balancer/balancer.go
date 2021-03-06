package balancer

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ddecaro94/gobalancer/config"
	"github.com/google/uuid"
)

var random = rand.New(rand.NewSource(int64(time.Now().Nanosecond())))

//A Balancer stores structures used to compute the host to be used at runtime
type Balancer struct {
	index     int
	frontend  string
	mutex     *sync.Mutex
	conf      *config.Config
	logger    *zap.Logger
	transport *http.Transport
	client    *http.Client
}

//New returns a Balancer instance
func New(c *config.Config, logger *zap.Logger, frontend string) (p *Balancer) {

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    120 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr, Timeout: 120 * time.Second}
	b := &Balancer{0, frontend, &sync.Mutex{}, c, logger, tr, client}
	return b
}

func (p *Balancer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	begin := time.Now()
	reqID, err := uuid.NewUUID()

	p.logger.Debug("Received request",
		zap.String("reqID", reqID.String()),
		zap.String("from", req.RemoteAddr),
		zap.String("method", req.Method),
		zap.String("host", req.URL.Host),
		zap.String("path", req.URL.Path),
	)
	servers := p.conf.Clusters[p.conf.Frontends[p.frontend].Pool].Servers
	iter, repeat, ttl, path, body := 0, true, len(servers), req.RequestURI, []byte{}

	forbidden := make(map[string]bool)
	bouncedCodes := p.conf.Frontends[p.frontend].Bounce
	if err != nil {
		panic(err)
	}
	if len(bouncedCodes) != 0 {
		//in case bounce is activated, replace reader with nopcloser in order to not close the request body
		b := req.Body
		defer req.Body.Close()
		req.Body = ioutil.NopCloser(b)
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
		reqTime := time.Now()
		res, httperr := p.client.Do(req)

		switch {

		case httperr != nil:
			p.logger.Warn("Error calling service",
				zap.String("reqID", reqID.String()),
				zap.String("from", req.RemoteAddr),
				zap.String("method", req.Method),
				zap.String("host", req.URL.Host),
				zap.String("path", req.URL.Path),
				zap.String("error", httperr.Error()),
			)
			//fmt.Printf("%s - %s - Calling %s %s, error: %s - redirecting\n", req.RemoteAddr, reqID, req.URL.Host, req.URL.Path, httperr.Error())
			repeat = true
			//add url to forbidden
			forbidden[host] = true
			if iter >= ttl {
				p.logger.Error("No valid host found for service",
					zap.String("reqID", reqID.String()),
					zap.String("from", req.RemoteAddr),
					zap.String("method", req.Method),
					zap.String("host", req.URL.Host),
					zap.String("path", req.URL.Path),
				)
				//fmt.Printf("%s - %s - No valid host found for service: %s \n", req.RemoteAddr, reqID, req.URL.Path)
				http.Error(resp, "Service Unavailable", 503)
			}
		case codeToBounce(res.StatusCode, p.conf.Frontends[p.frontend].Bounce):
			p.logger.Warn("Received error code - retrying",
				zap.String("reqID", reqID.String()),
				zap.String("from", req.RemoteAddr),
				zap.String("method", req.Method),
				zap.String("host", req.URL.Host),
				zap.String("path", req.URL.Path),
				zap.Int("code", res.StatusCode),
			)
			//fmt.Printf("%s - %s - Calling %s %s, received %d\n", req.RemoteAddr, reqID, req.URL.Host, req.URL.Path, res.StatusCode)
			repeat = true
			forbidden[host] = true
			if iter >= ttl {
				forward(resp, res)
			}
		default:
			p.logger.Info("Service request OK",
				zap.String("reqID", reqID.String()),
				zap.String("from", req.RemoteAddr),
				zap.String("method", req.Method),
				zap.String("host", req.URL.Host),
				zap.String("path", req.URL.Path),
				zap.Int("code", res.StatusCode),
				zap.Duration("requestDuration", time.Since(reqTime)),
			)
			//fmt.Printf("%s - %s - Calling %s %s, received %d\n", req.RemoteAddr, reqID, req.URL.Host, req.URL.Path, res.StatusCode)
			repeat = false
			e := forward(resp, res)
			if e != nil {
				p.logger.Error("Error occurred during response",
					zap.String("reqID", reqID.String()),
					zap.String("from", req.RemoteAddr),
					zap.String("method", req.Method),
					zap.String("host", req.URL.Host),
					zap.String("path", req.URL.Path),
					zap.String("error", e.Error()),
				)
			} else {
				p.logger.Debug("End reply",
					zap.String("reqID", reqID.String()),
					zap.String("from", req.RemoteAddr),
					zap.String("method", req.Method),
					zap.String("host", req.URL.Host),
					zap.String("path", req.URL.Path),
					zap.Int("code", res.StatusCode),
					zap.Duration("totalElapsed", time.Since(begin)),
				)
			}
		}
	}
}

//Next returns the next address according to the balancing algorithm
func (p *Balancer) Next() (server config.Server) {
	cluster := p.conf.Clusters[p.conf.Frontends[p.frontend].Pool]
	switch cluster.Algorithm {
	case "roundrobin":
		p.index = (p.index + 1) % len(cluster.Servers)

	case "weighted":
		p.index = getWeightedIndex(cluster)
	}
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

func forward(w http.ResponseWriter, res *http.Response) (err error) {
	for name, header := range res.Header {
		for _, val := range header {
			w.Header().Set(name, val)
		}
	}
	_, e := io.Copy(w, res.Body)
	defer res.Body.Close()

	return e
}

func getWeightedIndex(cluster *config.Cluster) (index int) {
	r := random.Intn(cluster.CDF)
	for i, s := range cluster.Servers {
		r -= s.Weight
		if r <= 0 {
			return i
		}
	}
	return -1
}
