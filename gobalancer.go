package main

import (
	"log"
	_ "net/http/pprof"

	"github.com/ddecaro94/gobalancer/api"
	"github.com/ddecaro94/gobalancer/config"
)

func main() {
	var c *config.Config
	var errJSON, errYAML error
	c, errJSON = config.ReadConfigJSON("./config.yml")
	if errJSON != nil {
		log.Println("Not a valid JSON config. Tying to read as YAML...")
		c, errYAML = config.ReadConfigYAML("./config.yml")
		if errYAML != nil {
			log.Println("Not a valid YAML config. Aborting...")
			panic(errYAML)
		}
		log.Println("Read config in yaml format.")
	}

	manager := api.NewManager(c)
	manager.Start()
}
