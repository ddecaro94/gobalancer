package api

import "net/http"

//Manager handles request for management
type Manager struct {
	Config *Config
}

func (s Manager) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

}
