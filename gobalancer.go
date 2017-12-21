package main

import (
	"github.com/ddecaro94/gobalancer/api"
	"github.com/ddecaro94/gobalancer/config"
)

func main() {

	c, err := config.ReadConfig("./config.json")
	if err != nil {
		panic(err)
	}

	manager := api.NewManager(c)

	manager.Start()
}
