package structs

import (
	"context"
	dcli "docker-project/docker/client"

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
	return dcli.Cli.ContainerStart(context.TODO(), cont.ID, types.ContainerStartOptions{})
}

func (cont Container) Stop() error {
	return dcli.Cli.ContainerStop(context.TODO(), cont.ID, nil)
}

func (cont Container) Restart() error {
	return dcli.Cli.ContainerRestart(context.TODO(), cont.ID, nil)
}

func (cont Container) Kill() error {
	return dcli.Cli.ContainerKill(context.TODO(), cont.ID, "SIGKILL")
}
