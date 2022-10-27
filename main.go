package main

import (
	"docker-project/api"
	"docker-project/config"
	"docker-project/docker"
	"log"
)

func main() {

	go docker.Init()
	if config.Port == "" {
		config.Port = ":6666"
	}

	log.Fatal(api.Start(config.Port))
}
