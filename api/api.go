package api

import (
	"net/http"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ddecaro94/gobalancer/balancer"
	"github.com/ddecaro94/gobalancer/config"
	"github.com/gorilla/mux"
)

//A Manager represents the management server object
type Manager struct {
	Config  *config.Config
	Server  *http.Server
	proxies map[string]*http.Server
	loggers map[string]*zap.AtomicLevel
	logger  *zap.Logger
	mutex   *sync.Mutex
}

//Start powers on the management api server
func (m *Manager) Start() {
	router := mux.NewRouter().StrictSlash(true)

	m.Server.Handler = router
	atomM := zap.NewAtomicLevelAt(zapcore.InfoLevel)

	zapcM := zap.Config{
		Level:            atomM,
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	loggerM, _ := zapcM.Build()
	m.logger = loggerM

	for _, frontend := range m.Config.Frontends {
		go func(frontend config.Frontend) {

			atom := zap.NewAtomicLevelAt(zapcore.InfoLevel)

			zapc := zap.Config{
				Level:            atom,
				Development:      false,
				Encoding:         "json",
				EncoderConfig:    zap.NewProductionEncoderConfig(),
				OutputPaths:      []string{"stderr", frontend.Logfile},
				ErrorOutputPaths: []string{"stderr", frontend.Logfile},
			}
			logger, _ := zapc.Build()

			m.loggers[frontend.Name] = &atom
			m.proxies[frontend.Name] = &http.Server{Addr: frontend.Listen, Handler: balancer.New(m.Config, logger, frontend.Name)}
			if frontend.Active {
				err := m.proxies[frontend.Name].ListenAndServe()
				if err != nil {
					panic(err)
				}
			}
		}(frontend)
	}
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	router.HandleFunc("/reload", m.ReloadConfig).Methods("GET")
	router.HandleFunc("/frontends", m.GetFrontends).Methods("GET")
	router.HandleFunc("/clusters", m.GetClusters).Methods("GET")
	router.HandleFunc("/frontends/{name}", m.GetFrontend).Methods("GET")
	router.HandleFunc("/frontends/{name}/log", m.LogLevel).Methods("GET", "PUT")
	router.HandleFunc("/log", atomM.ServeHTTP).Methods("GET", "PUT")
	router.HandleFunc("/clusters/{name}", m.GetCluster).Methods("GET")
	m.Server.ListenAndServe()
}

//NewManager creates a management server from a configuration object
func NewManager(c *config.Config) (m *Manager) {
	servers := make(map[string]*http.Server)
	loggers := make(map[string]*zap.AtomicLevel)

	man := &Manager{c, &http.Server{Addr: ":9999"}, servers, loggers, &zap.Logger{}, &sync.Mutex{}}

	return man
}
