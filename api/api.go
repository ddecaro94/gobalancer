package api

import (
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ddecaro94/gobalancer/balancer"
	"github.com/ddecaro94/gobalancer/config"
	"github.com/gorilla/mux"
)

//A Manager represents the management server object
type Manager struct {
	Config *config.Config
	Server *http.Server
	logger *zap.Logger
	mutex  *sync.Mutex
}

//Start powers on the management api server
func (m *Manager) Start() {
	router := mux.NewRouter().StrictSlash(true)

	tlsConf := &tls.Config{
		// Causes servers to use Go's default ciphersuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,
		// Only use curves which have assembly implementations
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519, // Go 1.8 only
		},
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

			// Best disabled, as they don't provide Forward Secrecy,
			// but might be necessary for some clients
			// tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			// tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	m.Server.Handler = router
	atomM := zap.NewAtomicLevelAt(zapcore.DebugLevel)

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
		go func(frontend *config.Frontend) {

			atom := zap.NewAtomicLevelAt(zapcore.DebugLevel)

			zapc := zap.Config{
				Level:            atom,
				Development:      false,
				Encoding:         "json",
				EncoderConfig:    zap.NewProductionEncoderConfig(),
				OutputPaths:      []string{"stderr", frontend.Logfile},
				ErrorOutputPaths: []string{"stderr", frontend.Logfile},
			}
			logger, err := zapc.Build()
			if err != nil {
				m.logger.Error("Could not start frontend",
					zap.String("name", frontend.Name),
					zap.String("error", err.Error()))
				return
			}

			frontend.Proxy = &http.Server{
				Addr:         frontend.Listen,
				Handler:      balancer.New(m.Config, logger, frontend.Name),
				TLSConfig:    tlsConf,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
			}
			frontend.LogLevel = &atom
			if frontend.Active {
				var err error
				m.logger.Info("Frontend starting up...",
					zap.String("name", frontend.Name),
					zap.String("Addr", frontend.Listen),
					zap.Bool("HTTPS", frontend.TLS.Enabled),
				)
				if frontend.TLS.Enabled {
					err = frontend.Proxy.ListenAndServeTLS(frontend.TLS.Cert, frontend.TLS.Key)

				} else {
					err = frontend.Proxy.ListenAndServe()
				}
				if err != nil {
					m.logger.Warn("Frontend has been stopped",
						zap.String("name", frontend.Name))
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
	router.HandleFunc("/frontends/{name}", m.PatchFrontend).Methods("PATCH")
	router.HandleFunc("/log", atomM.ServeHTTP).Methods("GET", "PUT")
	router.HandleFunc("/clusters/{name}", m.GetCluster).Methods("GET")
	m.Server.ListenAndServe()
}

//NewManager creates a management server from a configuration object
func NewManager(c *config.Config) (m *Manager) {

	man := &Manager{c, &http.Server{Addr: ":9999"}, &zap.Logger{}, &sync.Mutex{}}

	return man
}
