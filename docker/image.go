package docker

import (
	"bufio"
	"context"
	dcli "docker-project/docker/client"
	log "docker-project/logger"
	"encoding/json"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

func ListImages() any {
	images, err := dcli.Cli.ImageList(context.TODO(), types.ImageListOptions{All: true})
	if err != nil {
		log.Error(err)
		return nil
	}

	return images
}

func PruneImages() any {
	prune, err := dcli.Cli.ImagesPrune(context.TODO(), filters.Args{})
	if err != nil {
		log.Error(err)
		return nil
	}

	return prune
}

type Status struct {
	Status         string
	Progress       string
	ProgressDetail struct {
		Current int
		Total   int
	}
}

func PullImage(name string) chan Status {
	reader, err := dcli.Cli.ImagePull(context.TODO(), name, types.ImagePullOptions{})
	if err != nil {
		log.Error(err)
		return nil
	}
	result := make(chan Status)
	go func() {
		defer reader.Close()
		sc := bufio.NewScanner(reader)
		for sc.Scan() {
			st := Status{}
			json.Unmarshal(sc.Bytes(), &st)
			result <- st
			// log.PrintJSON(st)
		}
		close(result)
	}()
	return result
}

func RemoveImage(id string) error {
	_, err := dcli.Cli.ImageRemove(context.TODO(), id, types.ImageRemoveOptions{})
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
