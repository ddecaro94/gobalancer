package config

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"

	"github.com/go-yaml/yaml"
)

//TLS contains the frontend TLS settings
type TLS struct {
	Enabled bool   `json:"enabled"`
	Key     string `json:"key"`
	Cert    string `json:"cert"`
}

//Config for the main program
type Config struct {
	file      string
	format    string
	Frontends map[string]*Frontend `json:"frontends"`
	Clusters  map[string]*Cluster  `json:"clusters"`
}

//A Frontend is the proxy interface
type Frontend struct {
	Name     string           `json:"name"`
	Active   bool             `json:"active"`
	Listen   string           `json:"listen"`
	TLS      TLS              `json:"tls,omitempty"`
	Pool     string           `json:"pool"`
	Bounce   []int            `json:"bounce,omitempty"`
	Logfile  string           `json:"logfile,omitempty"`
	Proxy    *http.Server     `json:"-"`
	Logger   *zap.Logger      `json:"-"`
	LogLevel *zap.AtomicLevel `json:"-"`
}

//A Server represents a service provider listener
type Server struct {
	Name   string `json:"name"`
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Port   string `json:"port"`
	Weight int    `json:"weight,omitempty"`
}

//A Cluster aggregates more Servers into a pool
type Cluster struct {
	Algorithm string   `json:"algorithm"`
	CDF       int      `json:"-"` //cumulative density function (sum of weights)
	Servers   []Server `json:"servers"`
}

//ReadConfigJSON reads a json configuration file returning the Config object
func ReadConfigJSON(path string) (c *Config, err error) {
	var config Config
	file, e := ioutil.ReadFile(path)
	if e != nil {
		return nil, e
	}

	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}
	config.file = path
	config.format = "json"
	cdf := 0
	for key, cluster := range config.Clusters {
		for _, server := range cluster.Servers {
			cdf += server.Weight
		}
		cluster.CDF = cdf
		config.Clusters[key] = cluster
	}
	return &config, nil
}

//ReadConfigYAML reads a yaml configuration file returning the Config object
func ReadConfigYAML(path string) (c *Config, err error) {
	var config Config
	file, e := ioutil.ReadFile(path)
	if e != nil {
		return nil, e
	}

	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, err
	}
	config.file = path
	config.format = "yaml"
	cdf := 0
	for key, cluster := range config.Clusters {
		for _, server := range cluster.Servers {
			cdf += server.Weight
		}
		cluster.CDF = cdf
		config.Clusters[key] = cluster
	}
	return &config, nil
}

//Reload reloads config file
func (c *Config) Reload() (err error) {
	var (
		new *Config
	)
	if c.format == "json" {
		new, err = ReadConfigJSON(c.file)
		if err != nil {
			return err
		}
	}
	if c.format == "yaml" {
		new, err = ReadConfigYAML(c.file)
		if err != nil {
			return err
		}
	}

	c.Frontends = new.Frontends
	c.Clusters = new.Clusters
	for name := range new.Clusters {
		c.Clusters[name] = new.Clusters[name]
	}

	return nil
}

//Add server to the cluster
func (c *Cluster) Add(s Server) (result bool) {
	c.Servers = append(c.Servers, s)
	c.CDF += s.Weight
	return true
}

//Update server in the cluster
func (c *Cluster) Update(s Server) (updated int) {
	count := 0
	for _, server := range c.Servers {
		if server.Name == s.Name {
			server.Host = s.Host
			server.Port = s.Port
			server.Scheme = s.Scheme
			server.Weight = s.Weight
			c.CDF = c.CDF - server.Weight + s.Weight
			count++
		}
	}
	return count
}
