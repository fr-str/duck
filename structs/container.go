package structs

import (
	"context"
	dcli "docker-project/docker/client"
	log "docker-project/logger"

	"github.com/docker/docker/api/types"
	"github.com/timoni-io/go-utils/slice"
)

type Container struct {
	ID      string
	Name    string
	Image   string
	State   string
	Status  string
	Network string
	IP      string
	CMD     string
	Created int64
	Exited  int64
	Mounts  []Mount
	Ports   []Port
	Events  *slice.Slice[Event]
	Tty     bool
}

type Mount types.MountPoint
type Port types.Port

type Event struct {
	Action   string
	Status   string
	Type     string
	ExitCode string
	Time     int64
}

func (cont Container) Start() error {
	log.Debug("Starting container", cont.Name)
	return dcli.Cli.ContainerStart(context.TODO(), cont.ID, types.ContainerStartOptions{})
}

func (cont Container) Stop() error {
	log.Debug("Stopping container", cont.Name)
	return dcli.Cli.ContainerStop(context.TODO(), cont.ID, nil)
}

func (cont Container) Restart() error {
	log.Debug("Restarting container", cont.Name)
	return dcli.Cli.ContainerRestart(context.TODO(), cont.ID, nil)
}

func (cont Container) Kill() error {
	log.Debug("Killing container", cont.Name)
	return dcli.Cli.ContainerKill(context.TODO(), cont.ID, "SIGKILL")
}

func (cont Container) Delete() error {
	log.Debug("Deleting container", cont.Name)
	return dcli.Cli.ContainerRemove(context.TODO(), cont.ID, types.ContainerRemoveOptions{})
}

func (cont Container) Attach() error {
	log.Debug("Attaching to container", cont.Name)
	//WIP
	return nil
	dcli.Cli.ContainerAttach(context.TODO(), cont.ID, types.ContainerAttachOptions{
		Stream:     false,
		Stdin:      false,
		Stdout:     false,
		Stderr:     false,
		DetachKeys: "",
		Logs:       false,
	})
	return nil
}
