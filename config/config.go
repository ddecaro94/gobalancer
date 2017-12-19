package config

import (
	"encoding/json"
	"io/ioutil"
)

//Config for the main program
type Config struct {
	Frontends map[string]Frontend `json:"frontends"`
	Clusters  map[string]Cluster  `json:"clusters"`
}

//A Frontend is the proxy interface
type Frontend struct {
	Name    string `json:"name"`
	Listen  string `json:"listen"`
	TLS     bool   `json:"tls"`
	Pool    string `json:"pool"`
	Bounce  []int  `json:"bounce"`
	Logfile string `json:"logfile"`
}

//A Server represents a service provider listener
type Server struct {
	Name   string `json:"name"`
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
}

//A Cluster aggregates more Servers into a pool
type Cluster struct {
	Algorithm string   `json:"algorithm"`
	Size      int      `json:"size"`
	Servers   []Server `json:"servers"`
}

//ReadConfig reads the configuration file returning the Config object
func ReadConfig(path string) (c *Config, err error) {
	var config Config
	file, e := ioutil.ReadFile(path)
	if e != nil {
		return nil, e
	}
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
