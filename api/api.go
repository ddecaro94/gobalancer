package api

import (
	"net/http"
	"sync"

	"github.com/ddecaro94/gobalancer/balancer"
	"github.com/ddecaro94/gobalancer/config"
	"github.com/gorilla/mux"
)

//A Manager represents the management server object
type Manager struct {
	Config  *config.Config
	Server  *http.Server
	proxies map[string]*http.Server
	mutex   *sync.Mutex
}

//Start powers on the management api server
func (m *Manager) Start() {
	router := mux.NewRouter().StrictSlash(true)

	m.Server.Handler = router
	for _, frontend := range m.Config.Frontends {
		go func(frontend config.Frontend) {
			m.proxies[frontend.Name] = &http.Server{Addr: frontend.Listen, Handler: balancer.New(m.Config, frontend.Name)}
			if frontend.Active {
				err := m.proxies[frontend.Name].ListenAndServe()
				if err != nil {
					panic(err)
				}
			}
		}(frontend)
	}
	router.HandleFunc("/reload", m.ReloadConfig).Methods("GET")
	router.HandleFunc("/frontends", m.GetFrontends).Methods("GET")
	router.HandleFunc("/clusters", m.GetClusters).Methods("GET")
	router.HandleFunc("/frontends/{name}", m.GetFrontend).Methods("GET")
	router.HandleFunc("/clusters/{name}", m.GetCluster).Methods("GET")
	m.Server.ListenAndServe()
}

//NewManager creates a management server from a configuration object
func NewManager(c *config.Config) (m *Manager) {
	servers := make(map[string]*http.Server)

	man := &Manager{c, &http.Server{Addr: ":9999"}, servers, &sync.Mutex{}}

	return man
}
