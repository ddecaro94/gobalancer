package api

import (
	"net/http"

	"github.com/ddecaro94/gobalancer/config"
	"github.com/gorilla/mux"
)

//A Manager represents the management server object
type Manager struct {
	Config *config.Config
	Server *http.Server
}

//Start powers on the management api server
func (s *Manager) Start() {
	router := mux.NewRouter()
	router.HandleFunc("/reload", s.ReloadConfig).Methods("GET")
	s.Server.Handler = router
	s.Server.ListenAndServe()
}

//NewManager creates a management server from a configuration object
func NewManager(c *config.Config) (m *Manager) {
	man := &Manager{c, &http.Server{Addr: ":9999"}}
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
