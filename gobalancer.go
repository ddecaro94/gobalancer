package main

import (
	"log"
	_ "net/http/pprof"

	"github.com/ddecaro94/gobalancer/api"
	"github.com/ddecaro94/gobalancer/config"
	"github.com/spf13/viper"
)

func main() {
	c := &config.Config{}
	var errJSON, errYAML error
	//c, errJSON = config.ReadConfigJSON("./config.json")
	if errJSON != nil {
		log.Println("Not a valid JSON config. Tying to read as YAML...")
		c, errYAML = config.ReadConfigYAML("./config.yml")
		if errYAML != nil {
			log.Println("Not a valid YAML config. Aborting...")
			panic(errYAML)
		}
		log.Println("Read config in yaml format.")
	}
	log.Println("Read config in json format.")

	viper.SetConfigName("config")         // name of config file (without extension)
	viper.AddConfigPath("/etc/appname/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.appname") // call multiple times to add many search paths
	viper.AddConfigPath(".")              // optionally look for config in the working directory
	err := viper.ReadInConfig()           // Find and read the config file
	if err != nil {                       // Handle errors reading the config file
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	errnew := viper.UnmarshalKey("Frontends", &c.Frontends)
	if errnew != nil {
		return
	}

	manager := api.NewManager(c)
	manager.Start()
}
