package main

import (
	"docker-project/actions"
	"docker-project/api"
	"docker-project/config"
	"docker-project/docker"
)

func main() {

	go docker.Init()
	if config.Port == "" {
		config.Port = ":6666"
	}
	api.RegisterAction[actions.Containers]("container")
	api.RegisterSubscription[actions.Containers]("live.containers")
	api.RegisterSubscription[actions.Logs]("live.logs")
	api.RegisterAction[actions.Logs]("logs")
	api.Start(config.Port)
	select {}
}
