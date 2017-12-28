package api

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/ddecaro94/gobalancer/config"
	"github.com/gorilla/mux"
)

//ReloadConfig wraps config.reload
func (m *Manager) ReloadConfig(w http.ResponseWriter, r *http.Request) {
	m.logger.Info("Reloading config...")
	m.mutex.Lock()
	defer m.mutex.Unlock()
	err := m.Config.Reload()
	if err != nil {
		m.logger.Error("Unable to reload configuration",
			zap.String("error", err.Error()))
		http.Error(w, "Could not reload config", 500)
		return
	}
	/*	TODO
		handle static properties:
			-frontend.active
			-frontend.listen
			-frontend.tls
			-frontend.logfile
	*/
	m.logger.Info("Config successfully reloaded")
}

//GetFrontends returns a list of currently enabled frontends
func (m *Manager) GetFrontends(w http.ResponseWriter, r *http.Request) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	obj, err := json.MarshalIndent(m.Config.Frontends, "", "  ")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	} else {
		w.Write(obj)
	}
}

//GetFrontend returns a specific frontend
func (m *Manager) GetFrontend(w http.ResponseWriter, r *http.Request) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	vars := mux.Vars(r)
	name := vars["name"]
	if m.Config.Frontends[name].Name == "" {
		http.Error(w, "Resource Not Found", 404)
		return
	}
	obj, err := json.MarshalIndent(m.Config.Frontends[name], "", "  ")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	} else {
		w.Write(obj)
	}
}

//PatchFrontend updates a specific frontend's properties
func (m *Manager) PatchFrontend(w http.ResponseWriter, r *http.Request) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	vars := mux.Vars(r)
	name := vars["name"]
	var status *config.Frontend
	if m.Config.Frontends[name].Name == "" {
		http.Error(w, "Resource Not Found", 404)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&status)
	if err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}
	if m.Config.Frontends[name].Active {
		switch status.Active {
		case false:
			m.proxies[name].Shutdown(context.Background())
			m.Config.Frontends[name].Active = false
			w.Write([]byte("Frontend successfully stopped"))
		case true:
			w.Write([]byte("Already active"))
		default:
			http.Error(w, "Bad Request", 400)
		}
	} else {
		switch status.Active {
		case true:
			var err error
			go func(frontend *config.Frontend) {
				if frontend.TLS.Enabled {
					err = m.proxies[frontend.Name].ListenAndServeTLS(frontend.TLS.Cert, frontend.TLS.Key)

				} else {
					err = m.proxies[frontend.Name].ListenAndServe()
				}
				if err != nil {
					m.logger.Warn("Frontend has been stopped",
						zap.String("name", frontend.Name))
				}
			}(m.Config.Frontends[name])
			m.Config.Frontends[name].Active = true
			m.logger.Info("Frontend has been started",
				zap.String("name", m.Config.Frontends[name].Name))
			w.Write([]byte("Frontend successfully started"))
		case false:
			w.Write([]byte("Already stopped"))
		default:
			http.Error(w, "Bad Request", 400)
		}
	}

}

//GetClusters returns a list of currently available clusters
func (m *Manager) GetClusters(w http.ResponseWriter, r *http.Request) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	obj, err := json.MarshalIndent(m.Config.Clusters, "", "  ")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	} else {
		w.Write(obj)
	}
}

//GetCluster returns specific clusters
func (m *Manager) GetCluster(w http.ResponseWriter, r *http.Request) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	vars := mux.Vars(r)
	name := vars["name"]
	if len(m.Config.Clusters[name].Servers) == 0 {
		http.Error(w, "Resource Not Found", 404)
		return
	}
	obj, err := json.MarshalIndent(m.Config.Clusters[name], "", "  ")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
	} else {
		w.Write(obj)
	}
}

//LogLevel handles GET and POST request to hot modify log config
func (m *Manager) LogLevel(w http.ResponseWriter, r *http.Request) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	vars := mux.Vars(r)
	name := vars["name"]
	m.loggers[name].ServeHTTP(w, r)
}
