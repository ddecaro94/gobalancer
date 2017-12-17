package main

import (
	"container/ring"
	"net/url"
	"sync"
)

//A Backend represents a list of servers to be balanced
type Backend struct {
	hosts *ring.Ring
	mutex *sync.Mutex
}

//Next returns the next address according to the balancing algorithm
func (b *Backend) Next() (host string) {
	b.mutex.Lock()
	b.hosts = b.hosts.Next()
	defer b.mutex.Unlock()
	return b.hosts.Value.(string)
}

//NewBackend returns a setof servers using the balancing algorithm
func NewBackend(algorithm string, hosts ...string) (*Backend, error) {
	var err error
	b := &Backend{}
	b.mutex = &sync.Mutex{}
	switch algorithm {
	case "roundrobin":
		b.hosts = ring.New(len(hosts))
		for _, host := range hosts {
			b.hosts.Value, err = url.Parse(host)
			if err != nil {
				return nil, err
			}
			b.hosts = b.hosts.Next()
		}
		return b, nil
	default:
		return nil, err
	}

}
