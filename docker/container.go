package docker

import (
	"context"
	dcli "docker-project/docker/client"
	log "docker-project/logger"
	"docker-project/structs"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/timoni-io/go-utils/maps"
	"github.com/timoni-io/go-utils/slice"
)

var ContainerMap = maps.New(make(map[string]structs.Container)).Safe().Eventfull(context.TODO(), 10)
var DockerContainerMap = maps.New(make(map[string]types.ContainerJSON)).Safe()

type DockerContainer types.Container

func Patch(obj1, obj2 any) ([]byte, error) {
	objb1, err := json.Marshal(obj1)
	if err != nil {
		return nil, err
	}
	objb2, err := json.Marshal(obj2)
	if err != nil {
		return nil, err
	}
	return jsonpatch.CreateMergePatch(objb1, objb2)
}

func SetContainer(dcont types.Container) {
	cont := DockerContainer(dcont)
	name := strings.TrimLeft(cont.Names[0], "/")
	// var x structs.Container
	cmap, ok := ContainerMap.GetFull(name)
	cdoc := structs.Container{
		ID:      cont.ID,
		Name:    name,
		Image:   cont.Image,
		State:   cont.State,
		Status:  cont.Status,
		Network: cont.setNet(),
		IP:      cont.setIP(),
		CMD:     cont.Command,
		Mounts:  cont.getMounts(),
		Ports:   cont.getPorts(),
		Created: cont.Created,
	}

	if ok {
		cdoc.Tty = cmap.Tty
		cdoc.Events = cmap.Events

		patch, err := Patch(cmap, cdoc)
		if err != nil {
			log.Error(err)
		}
		if len(patch) == 2 {
			return
		}

		// log.Debug(len(patch))
		// x = cmap

	} else {
		t, err := dcli.Cli.ContainerInspect(context.TODO(), cont.ID)
		if err != nil {
			if strings.Contains(err.Error(), "No such container") {
				return
			}
			log.Error(err)
		}
		DockerContainerMap.Set(strings.TrimLeft(cont.Names[0], "/"), t)
		cdoc.Events = slice.NewSafeSlice[structs.Event](5)
		cdoc.Tty = t.Config.Tty
	}

	// log.Debug("Setting", name)
	// patch, err := Patch(cdoc, x)
	// if err != nil {
	// 	log.Error(err)
	// }
	// patch2, err := Patch(x, cdoc)
	// if err != nil {
	// 	log.Error(err)
	// }
	// log.Debug(cdoc.Name, string(patch), string(patch2))

	ContainerMap.Set(name, cdoc)
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
			SetContainer(cont)
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
