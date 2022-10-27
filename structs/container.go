package structs

import (
	"github.com/docker/docker/api/types"
	"github.com/timoni-io/go-utils/slice"
)

type Container struct {
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
