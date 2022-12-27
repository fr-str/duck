package actions

import (
	ws "docker-project/api/server"
	"docker-project/docker"
	log "docker-project/logger"
	"docker-project/wsc"
	"strings"
)

type Image struct {
	ID, Name string
	Prune    bool
}

func (a *Image) Handle(r *ws.Request) ws.Response {
	act := strings.TrimPrefix(r.Action, "image.")
	switch act {
	case "new":
		return a.New(r)
	case "pull":
		return a.Pull(r)
	case "list":
		return a.List(r)
	case "delete":
		return a.Delete(r)
	default:
		log.Error(wsc.Action.String() + wsc.NotFound.String())
		return ws.Error(r, nil, wsc.Action, wsc.NotFound)
	}
}

func (a *Image) New(r *ws.Request) ws.Response {
	return ws.Ok(r, "not ok")
}

func (a *Image) List(r *ws.Request) ws.Response {
	return ws.Ok(r, docker.ListImages())
}

func (a *Image) Pull(r *ws.Request) ws.Response {
	if a.Name == "" {
		return ws.Error(r, nil, wsc.Missing, wsc.Name)
	}

	r.ResultCh <- ws.Response{
		RequestID: r.RequestID,
		Data:      "STREAM START",
	}

	for st := range docker.PullImage(a.Name) {
		r.ResultCh <- ws.Response{
			RequestID: r.RequestID,
			Code:      wsc.OK.Get(),
			Data:      st.Status,
		}

	}

	return ws.Ok(r, "STREAM END")
}

// \u003cnone\u003e@\u003cnone\u003e
func (a *Image) Delete(r *ws.Request) ws.Response {
	if a.ID == "" && !a.Prune {
		return ws.Error(r, nil, wsc.Missing, wsc.ID)
	}

	if a.Prune {
		return ws.Ok(r, docker.PruneImages())
	}

	if err := docker.RemoveImage(a.ID); err != nil {
		return ws.Error(r, nil, wsc.Image, wsc.Delete, wsc.Error)
	}

	return ws.Ok(r, "not ok")
}
