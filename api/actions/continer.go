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
	Name      string
	SoftForce bool
	Force     bool
}

func (a *Containers) Handle(r *ws.Request) ws.Response {
	act := strings.TrimPrefix(r.Action, "container.")
	switch act {
	case "stop", "restart", "start", "kill":
		return a.SSRK(r)
	case "list":
		return a.List(r)
	case "create":
		return a.Create(r)
	case "delete":
		return a.Delete(r)

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

func (a *Containers) SSRK(r *ws.Request) ws.Response {
	cont, ok := docker.ContainerMap.GetFull(a.Name)
	if !ok {
		return ws.Error(r, er.NotFound+er.Container)
	}

	switch strings.TrimPrefix(r.Action, "container.") {
	case "start":
		ws.GoError(r, er.InternalServerError, cont.Start)
	case "stop":
		ws.GoError(r, er.InternalServerError, cont.Stop)
	case "restart":
		ws.GoError(r, er.InternalServerError, cont.Restart)
	case "kill":
		ws.GoError(r, er.InternalServerError, cont.Kill)
	}

	return ws.Ok(r, "ok")
}

func (a *Containers) Create(r *ws.Request) ws.Response {
	_, ok := docker.ContainerMap.GetFull(a.Name)
	if ok {
		return ws.Error(r, er.Exists+er.Container)
	}
	//TODO

	return ws.Ok(r, "ok")
}

func (a *Containers) Delete(r *ws.Request) ws.Response {
	cont, ok := docker.ContainerMap.GetFull(a.Name)
	if !ok {
		return ws.Error(r, er.NotFound+er.Container)
	}

	if !a.Force && !a.SoftForce && cont.State != "exited" {
		return ws.Error(r, er.Forbbiden+er.ContainerIsRunning)
	}

	if a.SoftForce {
		ws.GoError(r, er.InternalServerError, cont.Stop, cont.Delete)
	} else {
		ws.GoError(r, er.InternalServerError, cont.Kill, cont.Delete)
	}

	return ws.Ok(r, "ok")
}

func (a *Containers) List(r *ws.Request) ws.Response {
	m := map[string]structs.Container{}
	for v := range docker.ContainerMap.Iter() {
		m[v.Key] = v.Value
	}
	return ws.Ok(r, m)
}
