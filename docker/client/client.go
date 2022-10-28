package dcli

import (
	log "docker-project/logger"

	"github.com/docker/docker/client"
)

var Cli *client.Client

func Init() {
	log.Info("Init...")
	var err error
	Cli, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
}
