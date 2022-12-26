package docker

import (
	"context"
	dcli "docker-project/docker/client"
	log "docker-project/logger"
	"docker-project/structs"
	"strings"
	"time"

	ty "github.com/timoni-io/go-utils/types"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

type eventHandler struct {
	ctx    context.Context
	cancel context.CancelFunc
	er     <-chan error
	ev     <-chan events.Message
}

func HandleEvents(cli *client.Client) {
	eh := eventHandler{}
	eh.ctx, eh.cancel = context.WithCancel(context.Background())
	eh.ev, eh.er = cli.Events(eh.ctx, types.EventsOptions{})
	go eh.readEr(cli)
	go eh.readEv(cli)
}

func (eh eventHandler) readEr(cli *client.Client) {
	for err := range eh.er {
		log.Error(err)
		eh.cancel()
		time.Sleep(time.Second)
		HandleEvents(cli)
	}
}

func (eh eventHandler) readEv(cli *client.Client) {
	for {
		select {
		case <-eh.ctx.Done():
			return
		case ev := <-eh.ev:
			// TODO handle other event types
			switch ev.Type {
			case events.ContainerEventType:
				handleContainer(cli, ev)
			default:
				log.Debug(ev.Type, ev.Action)
			}
		}
	}
}

func handleContainer(cli *client.Client, ev events.Message) {
	// fmt.Println(ev.Actor.Attributes["name"],"event: ", ev.Action)
	// ignore exec events, untill i fix container info
	// switch {
	// case strings.HasPrefix(ev.Action, "exec_"):
	// 	return
	// }
	updateChan <- struct{}{}

	name := ev.Actor.Attributes["name"]
	cont, ok := ContainerMap.GetFull(name)
	if !ok && ev.Action != "rename" {
		return
	}
	if ev.Action == "destroy" {
		ContainerMap.Delete(name)
		DockerContainerMap.Delete(name)
		return
	}

	if ev.Action == "rename" {
		_, oldName, ok := strings.Cut(name, "_")
		if !ok {
			return
		}
		cont, ok := ContainerMap.GetFull(oldName)
		if !ok {
			return
		}

		ContainerMap.Delete(oldName)
		DockerContainerMap.Delete(oldName)
		ContainerMap.Set(name, cont)
		go func() {
			t, err := dcli.Cli.ContainerInspect(context.TODO(), cont.ID)
			if err != nil {
				if strings.Contains(err.Error(), "No such container") {
					return
				}
				log.Error(err)
			}
			log.Debug("Set container", t.Name)
			DockerContainerMap.Set(strings.TrimLeft(t.Name, "/"), t)
		}()
		return
	}

	contEvent := structs.Event{
		Action: ev.Action,
		Status: ev.Status,
		Type:   ev.Type,
		Time:   ev.TimeNano,
	}

	switch ev.Action {
	case "die":
		contEvent.ExitCode = ev.Actor.Attributes["exitCode"]
	default:
	}

	// cont.Events.Add(contEvent)
	ContainerMap.Broadcast(ty.WatchMsg[string, structs.Container]{
		Event: ty.PutEvent,
		Item: ty.Item[string, structs.Container]{
			Key:   cont.Name,
			Value: cont,
		},
	})
}
