package main

import (
	"context"
	"docker-project/api"
	"docker-project/config"
	"docker-project/docker"
	dcli "docker-project/docker/client"
	log "docker-project/logger"

	"github.com/docker/docker/api/types"
)

func main() {
	if config.Port == "" {
		config.Port = ":6666"
	}

	dcli.Init()

	log.Info("Getting container list...")
	list, err := dcli.Cli.ContainerList(context.TODO(), types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}
	log.Info("Got container list")

	for _, cont := range list {
		docker.SetContainer(docker.DockerContainer(cont))
	}

	go docker.UpdateMap(dcli.Cli)
	go docker.HandleEvents(dcli.Cli)
	log.Fatal(api.Start(config.Port))
}
