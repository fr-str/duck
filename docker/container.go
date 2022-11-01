package docker

import (
	"context"
	dcli "docker-project/docker/client"
	log "docker-project/logger"
	"docker-project/structs"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/timoni-io/go-utils/maps"
	"github.com/timoni-io/go-utils/slice"
)

var ContainerMap = maps.New(make(map[string]structs.Container)).Safe().Eventfull(context.TODO(), 10)

type DockerContainer types.Container

func SetContainer(dcont types.Container) {
	cont := DockerContainer(dcont)
	name := strings.TrimLeft(cont.Names[0], "/")
	ev := slice.NewSafeSlice[structs.Event](5)
	if co, ok := ContainerMap.GetFull(name); ok {
		black := /* false // */ strings.Contains(cont.Status, "second") && !strings.Contains(cont.Status, "Less")
		magic := co.State == cont.State && co.Status == cont.Status
		if black || magic {
			return
		}
		ev = co.Events
	}

	t, err := dcli.Cli.ContainerInspect(context.TODO(), cont.ID)
	if err != nil {
		if strings.Contains(err.Error(), "No such container") {
			return
		}
		log.Error(err)
	}

	log.Debug("Setting", name)
	c := structs.Container{
		ID:      cont.ID,
		Name:    name,
		Tty:     t.Config.Tty,
		Image:   cont.Image,
		State:   cont.State,
		Status:  cont.Status,
		Network: cont.setNet(),
		IP:      cont.setIP(),
		CMD:     cont.Command,
		Mounts:  cont.getMounts(),
		Ports:   cont.getPorts(),
		Created: cont.Created,
		Events:  ev,
	}
	ContainerMap.Set(name, c)
}

func UpdateMap(cli *client.Client /* , name string */) error {
	tic := time.NewTicker(100 * time.Millisecond)
	defer tic.Stop()
	for range tic.C {
		l, err := cli.ContainerList(context.TODO(), types.ContainerListOptions{
			All: true,
		})
		if err != nil {
			log.Error(err)
		}

		for _, cont := range l {
			go SetContainer(cont)
		}
	}
	return nil
}

func (dcont DockerContainer) setIP() string {
	var ip string
	for _, v := range dcont.NetworkSettings.Networks {
		ip = v.IPAddress
		return ip
	}

	return ip
}

func (dcont DockerContainer) setNet() string {
	var netw string
	for k := range dcont.NetworkSettings.Networks {
		netw = k
		break
	}

	return netw
}
func (dcont DockerContainer) getMounts() (mounts []structs.Mount) {
	for _, mount := range dcont.Mounts {
		mounts = append(mounts, structs.Mount(mount))
	}
	sort.Slice(mounts, func(i, j int) bool {
		return mounts[i].Destination < mounts[j].Destination
	})

	return
}

func (dcont DockerContainer) getPorts() (Ports []structs.Port) {
	for _, port := range dcont.Ports {
		Ports = append(Ports, structs.Port(port))
	}

	sort.Slice(Ports, func(i, j int) bool {
		return Ports[i].PrivatePort < Ports[j].PrivatePort
	})
	return
}
