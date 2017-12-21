package api

import (
	"encoding/json"
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
	router := mux.NewRouter().StrictSlash(false)
	router.HandleFunc("/reload", m.ReloadConfig).Methods("GET")
	router.HandleFunc("/frontends", m.GetFrontends).Methods("GET")
	router.HandleFunc("/clusters", m.GetClusters).Methods("GET")
	router.HandleFunc("/frontends/{name}", m.GetFrontend).Methods("GET")
	router.HandleFunc("/clusters/{name}", m.GetCluster).Methods("GET")
	m.Server.Handler = router
	for _, frontend := range m.Config.Frontends {
		go func(frontend config.Frontend) {
			m.proxies[frontend.Name] = &http.Server{Addr: frontend.Listen, Handler: balancer.New(m.Config, frontend.Name)}
			err := m.proxies[frontend.Name].ListenAndServe()
			if err != nil {
				panic(err)
			}
		}(frontend)
	}
	m.Server.ListenAndServe()
}

//NewManager creates a management server from a configuration object
func NewManager(c *config.Config) (m *Manager) {
	servers := make(map[string]*http.Server)

	man := &Manager{c, &http.Server{Addr: ":9999"}, servers, &sync.Mutex{}}

	return man
}

//ReloadConfig wraps config.reload
func (m *Manager) ReloadConfig(w http.ResponseWriter, r *http.Request) {
	err := m.Config.Reload()
	if err != nil {
		http.Error(w, "Could not reload config", 500)
		return
	}
}

//GetFrontends returns a list of currently enabled frontends
func (m *Manager) GetFrontends(w http.ResponseWriter, r *http.Request) {
	obj, err := json.MarshalIndent(m.Config.Frontends, "", "\t")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	} else {
		w.Write(obj)
	}
}

//GetFrontend returns a specific frontend
func (m *Manager) GetFrontend(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	obj, err := json.MarshalIndent(m.Config.Frontends[name], "", "\t")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	} else {
		w.Write(obj)
	}
}

//GetClusters returns a list of currently available clusters
func (m *Manager) GetClusters(w http.ResponseWriter, r *http.Request) {
	obj, err := json.MarshalIndent(m.Config.Clusters, "", "\t")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	} else {
		w.Write(obj)
	}
}

//GetCluster returns specific clusters
func (m *Manager) GetCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	obj, err := json.MarshalIndent(m.Config.Clusters[name], "", "\t")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	} else {
		w.Write(obj)
	}
}
