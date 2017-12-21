package api

import (
	"fmt"
	"net/http"

	"github.com/ddecaro94/gobalancer/balancer"
	"github.com/ddecaro94/gobalancer/config"
	"github.com/gorilla/mux"
)

//A Manager represents the management server object
type Manager struct {
	Config  *config.Config
	Server  *http.Server
	proxies map[string]*http.Server
}

//Start powers on the management api server
func (s *Manager) Start() {
	router := mux.NewRouter()
	router.HandleFunc("/reload", s.ReloadConfig).Methods("GET")
	s.Server.Handler = router
	for _, frontend := range s.Config.Frontends {
		go func(frontend config.Frontend) {
			cluster := s.Config.Clusters[frontend.Pool]
			s.proxies[frontend.Name] = &http.Server{Addr: frontend.Listen, Handler: balancer.New(s.Config, frontend.Name)}
			err := s.proxies[frontend.Name].ListenAndServe()
			if err != nil {
				panic(err)
			} else {
				fmt.Println("Listening on %s, frontend %+v, cluster %+v", frontend.Listen, frontend, cluster)
			}
		}(frontend)
	}
	s.Server.ListenAndServe()
}

//NewManager creates a management server from a configuration object
func NewManager(c *config.Config) (m *Manager) {
	servers := make(map[string]*http.Server)

	man := &Manager{c, &http.Server{Addr: ":9999"}, servers}

	return man
}

//ReloadConfig wraps config.reload
func (s *Manager) ReloadConfig(w http.ResponseWriter, r *http.Request) {
	err := s.Config.Reload()
	if err != nil {
		http.Error(w, "Could not reload config", 500)
		return
	}
}
