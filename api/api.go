package api

import (
	"net/http"

	"github.com/ddecaro94/gobalancer/config"
)

//A Manager represents the management server object
type Manager struct {
	Config *config.Config
}

//Start powers on the management api server
func (s Manager) Start() {
	http.ListenAndServe(":9999", nil)
}

//NewManager creates a management server from a configuration object
func NewManager(c *config.Config) (m *Manager) {
	man := &Manager{c}
	return man
}
