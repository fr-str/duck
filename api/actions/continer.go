package actions

import (
	ws "docker-project/api/server"
	log "docker-project/logger"

	"docker-project/docker"
	"docker-project/er"
	"docker-project/structs"
	"strings"

	"github.com/timoni-io/go-utils/types"
)

type Containers struct {
	EnvID string
	Key   string
	Value string
}

func (a *Containers) Handle(r *ws.Request) ws.Response {
	act := strings.TrimPrefix(r.Action, "container.")
	switch act {
	case "stop":
		return a.Stop(r)
	case "start":
		return a.Start(r)
	case "restart":
		return a.Restart(r)
	case "list":
		return a.List(r)

	default:
		log.Error(er.Action.String() + er.NotFound.String())
		return ws.Error(r, er.Action+er.NotFound)
	}
}

func (a *Containers) HandleSub(r *ws.Request, w chan<- ws.Response) {
	for el := range docker.ContainerMap.Iter() {
		w <- ws.Ok(r, types.WatchMsg[string, structs.Container]{
			Event: types.PutEvent,
			Item:  el,
		})
	}

	for v := range docker.ContainerMap.Register() {
		select {
		case <-r.Ctx.Done():
			return
		case w <- ws.Ok(r, v):
		}
	}
}

func (a *Containers) Stop(r *ws.Request) ws.Response    { return ws.Ok(r, nil) }
func (a *Containers) Start(r *ws.Request) ws.Response   { return ws.Ok(r, nil) }
func (a *Containers) Restart(r *ws.Request) ws.Response { return ws.Ok(r, nil) }
func (a *Containers) List(r *ws.Request) ws.Response {
	m := map[string]structs.Container{}
	for v := range docker.ContainerMap.Iter() {
		m[v.Key] = v.Value
	}
	return ws.Ok(r, m)
}
