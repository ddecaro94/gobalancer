package main

import (
	"net/url"
	"sync"
)

//A Backend represents a list of servers to be balanced
type Backend struct {
	algorithm string
	len       int
	index     int
	hosts     []*url.URL
	mutex     *sync.Mutex
}

//Next returns the next address according to the balancing algorithm
func (b *Backend) Next() (host *url.URL) {
	b.mutex.Lock()
	switch b.algorithm {
	case "roundrobin":
		b.index = (b.index + 1) % b.len
	}
	defer b.mutex.Unlock()
	return b.hosts[b.index]
}

//NewBackend returns a setof servers using the balancing algorithm
func NewBackend(c Cluster) *Backend {
	b := &Backend{}
	b.mutex = &sync.Mutex{}
	b.len = c.Size
	b.algorithm = c.Algorithm
	b.index = 0
	for _, host := range c.Servers {
		b.hosts = append(b.hosts, &url.URL{Host: host.Host, Scheme: host.Scheme})
	}
	return b

}
