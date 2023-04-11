package docker

import (
	"bufio"
	"context"
	dcli "docker-project/docker/client"
	log "docker-project/logger"
	"docker-project/structs"
	"encoding/json"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/fr-str/go-utils"
	"github.com/fr-str/go-utils/maps"
)

var (
	// key: container name (string)
	Containers = maps.New(make(map[string]*structs.Container)).Safe().Eventfull(context.TODO(), 10)
	// key: container name (string)
	DockerContainerMap = maps.New(make(map[string]*types.ContainerJSON)).Safe()
	updateChan         = make(chan struct{}, 10)
)

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

func SetContainer(cont DockerContainer) {
	name := cont.getName()
	cmap, ok := Containers.GetFull(name)
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
		cdoc.Ctx = cmap.Ctx
		cdoc.Cancel = cmap.Cancel
		cdoc.Tty = cmap.Tty
		cdoc.Stats = cmap.Stats
		cdoc.Started = cmap.Started
		cdoc.Exited = cmap.Exited

		patch, err := Patch(cmap, cdoc)
		if err != nil {
			log.Error(err)
		}

		if len(patch) == 2 {
			return
		}

	}

	if cdoc.Stats == nil {
		cdoc.Stats = &structs.MiniStats{}
	}
	if cdoc.Ctx == nil {
		cdoc.Ctx, cdoc.Cancel = context.WithCancel(context.Background())
	}

	go func() {
		t, err := dcli.Cli.ContainerInspect(context.TODO(), cont.ID)
		if err != nil {
			if strings.Contains(err.Error(), "No such container") {
				return
			}
			log.Error(err)
		}

		DockerContainerMap.Set(cont.getName(), &t)
		cdoc.Tty = t.Config.Tty
		cdoc.Started = utils.Must(time.Parse(time.RFC3339Nano, t.State.StartedAt)).Unix()
		cdoc.Exited = utils.Must(time.Parse(time.RFC3339Nano, t.State.FinishedAt)).Unix()
		if t.State.Running {
			cdoc.Exited = 0
		}
	}()

	Containers.Set(name, &cdoc)
	if !cdoc.Stats.Monitored {
		go GetStats(cdoc.Stats, cdoc.ID, cdoc.Name, cdoc.Ctx)
		cdoc.Stats.Monitored = true
	}
}

func UpdateMap(cli *client.Client) {
	for range updateChan {
		l, err := cli.ContainerList(context.TODO(), types.ContainerListOptions{All: true})
		if err != nil {
			log.Error(err)
		}

		for _, cont := range l {
			SetContainer(DockerContainer(cont))
		}
	}

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

func (dcont DockerContainer) getName() string {
	return strings.TrimLeft(dcont.Names[0], "/")
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

func GetStats(c *structs.MiniStats, id, name string, ctx context.Context) {
	st, err := dcli.Cli.ContainerStats(ctx, id, true)
	if err != nil {
		log.Error(err)
		c.Monitored = false
		return
	}
	go func() {
		<-ctx.Done()
		st.Body.Close()
		log.Debug("ctx.Done(), body close ", id, "-------------------------")
	}()
	sc := bufio.NewScanner(st.Body)
	var lastCPU uint64 = 0
	const MiB = 2 << 20
	for sc.Scan() {
		s := structs.Stats{}
		json.Unmarshal(sc.Bytes(), &s)

		c.CPUUsage = roundFloat(float64(s.CPUStats.CPUUsage.TotalUsage-lastCPU)/10000000, 2)
		c.Memory = structs.MiniMem{
			Usage: roundFloat(float64(s.MemoryStats.Usage)/MiB, 2),
			Limit: roundFloat(float64(s.MemoryStats.Limit)/float64(MiB), 2),
		}
		c.Network = structs.MiniNet{
			I: roundFloat(float64(s.Networks.Eth0.RxBytes)/MiB, 2),
			O: roundFloat(float64(s.Networks.Eth0.TxBytes)/MiB, 2),
		}
		if len(s.BlkioStats.IoServiceBytesRecursive) > 0 {
			c.BlockIO = structs.MiniBlk{
				I: roundFloat(float64(s.BlkioStats.IoServiceBytesRecursive[0].Value)/MiB, 2),
				O: roundFloat(float64(s.BlkioStats.IoServiceBytesRecursive[1].Value)/MiB, 2),
			}
		}
		lastCPU = s.CPUStats.CPUUsage.TotalUsage
	}

	c.Monitored = false
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
