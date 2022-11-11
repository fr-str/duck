package actions

import (
	ws "docker-project/api/server"
	"docker-project/docker"
	log "docker-project/logger"
	"docker-project/wsc"
	"sort"
	"strings"
)

type Logs struct {
	ContainerNames []string
	Amount         int
	Since          int64
	Until          int64
}

func (a *Logs) Handle(r *ws.Request) ws.Response {
	if len(a.ContainerNames) == 0 {
		log.Debug(wsc.Missing.String() + wsc.ContainerName.String())
		return ws.Error(r, wsc.Missing+wsc.ContainerName)
	}
	act := strings.TrimPrefix(r.Action, "logs.")
	switch act {
	case "get":
		return a.Before(r)
	default:
		log.Debug(wsc.Action.String() + wsc.NotFound.String())
		return ws.Error(r, wsc.Action+wsc.NotFound)
	}
}

func (a *Logs) Before(r *ws.Request) ws.Response {
	var logs []docker.Log
	for _, cName := range a.ContainerNames {
		lgs, _, err := docker.GetLogs(cName, a.Amount, a.Since, a.Until, false)
		if err != nil {
			if err == docker.ErrContNotExist {
				log.Debug(wsc.Container.String() + wsc.NotFound.String())
				return ws.Error(r, wsc.Container+wsc.NotFound)
			}
			log.Error(wsc.InternalServerError.String())
			return ws.Error(r, wsc.InternalServerError)
		}
		logs = append(logs, lgs...)
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp > logs[j].Timestamp
	})
	return ws.Ok(r, logs)
}
