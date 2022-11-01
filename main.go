package main

import (
	"docker-project/api"
	"docker-project/config"
	log "docker-project/logger"
)

func main() {
	if config.Port == "" {
		config.Port = ":6666"
	}

	log.Fatal(api.Start(config.Port))
}
