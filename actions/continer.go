package actions

import (
	"docker-project/api"
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

func (a *Containers) Handle(r *api.Request) api.Response {
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
		return api.Error(r, er.Action+er.NotFound)
	}
}

func (a *Containers) HandleSub(r *api.Request, w chan<- api.Response) {
	for el := range docker.ContainerMap.Iter() {
		w <- api.Ok(r, types.WatchMsg[string, structs.Container]{
			Event: types.PutEvent,
			Item:  el,
		})
	}

	for v := range docker.ContainerMap.Register() {
		select {
		case <-r.Ctx.Done():
			return
		case w <- api.Ok(r, v):
		}
	}
}

func (a *Containers) Stop(r *api.Request) api.Response    { return api.Ok(r, nil) }
func (a *Containers) Start(r *api.Request) api.Response   { return api.Ok(r, nil) }
func (a *Containers) Restart(r *api.Request) api.Response { return api.Ok(r, nil) }
func (a *Containers) List(r *api.Request) api.Response {
	m := map[string]structs.Container{}
	for v := range docker.ContainerMap.Iter() {
		m[v.Key] = v.Value
	}
	return api.Ok(r, m)
}
