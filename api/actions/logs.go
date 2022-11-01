package actions

import (
	ws "docker-project/api/server"
	"docker-project/docker"
	"docker-project/er"
	log "docker-project/logger"
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
		log.Debug(er.Missing.String() + er.ContainerName.String())
		return ws.Error(r, er.Missing+er.ContainerName)
	}
	act := strings.TrimPrefix(r.Action, "logs.")
	switch act {
	case "get":
		return a.Before(r)
	default:
		log.Debug(er.Action.String() + er.NotFound.String())
		return ws.Error(r, er.Action+er.NotFound)
	}
}

func (a *Logs) Before(r *ws.Request) ws.Response {
	var logs []docker.Log
	for _, cName := range a.ContainerNames {
		lgs, _, err := docker.GetLogs(cName, a.Amount, a.Since, a.Until, false)
		if err != nil {
			if err == docker.ErrContNotExist {
				log.Debug(er.Container.String() + er.NotFound.String())
				return ws.Error(r, er.Container+er.NotFound)
			}
			log.Error(er.InternalServerError.String())
			return ws.Error(r, er.InternalServerError)
		}
		logs = append(logs, lgs...)
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp > logs[j].Timestamp
	})
	return ws.Ok(r, logs)
}
