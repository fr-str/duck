package docker

import (
	log "docker-project/logger"
	"time"

	"github.com/docker/docker/client"
)

var dcli *client.Client

func Init() {
	log.Info("Init...")
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	// temp solution, planned to suport multiple nodes in the future
	dcli = cli

	go UpdateMap(cli)

	go HandleEvents(cli)

	time.Sleep(500 * time.Millisecond)
	log.Info("Init complete,", "found", ContainerMap.Len(), "containers")
}
