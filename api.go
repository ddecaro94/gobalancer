package api

import "net/http"

//APIServer handles request for management
type APIServer struct {
	Config *Config
}

func (s APIServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

}
