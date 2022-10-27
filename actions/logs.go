package actions

import (
	"bufio"
	"docker-project/api"
	"docker-project/docker"
	"docker-project/er"
	log "docker-project/logger"
	"strings"
	"time"
)

type Logs struct {
	ContainerName string
	Amount        int
	Since         int64
	Until         int64
}

func (a *Logs) Handle(r *api.Request) api.Response {
	if a.ContainerName == "" {
		return api.Error(r, er.Missing+er.ContainerName)
	}
	act := strings.TrimPrefix(r.Action, "logs.")
	switch act {
	case "before":
		return a.Before(r)
	default:
		return api.Error(r, er.Action+er.NotFound)
	}
}

func (a *Logs) HandleSub(r *api.Request, w chan<- api.Response) {
	if a.Until != 0 {
		w <- api.Error(r, er.Forbbiden+er.UntilInLive)
		return
	}

	logs, _, err := docker.GetLogs(a.ContainerName, a.Amount, a.Since, 0, false)
	if err != nil {
		if err == docker.ErrContNotExist {
			w <- api.Error(r, er.Container+er.NotFound)
			return
		}
		w <- api.Error(r, er.InternalServerError)
		return
	}
	w <- api.Ok(r, logs)

	var strip bool
	if !docker.ContainerMap.Get(a.ContainerName).Tty {
		strip = true
	}

	_, rc, _ := docker.GetLogs(a.ContainerName, 0, a.Since, 0, true)
	var line string
	var bline []byte
	sc := bufio.NewScanner(rc)
	for {
		if !sc.Scan() {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		select {
		case <-r.Ctx.Done():
			rc.Close()
			return
		default:
		}

		bline = sc.Bytes()

		line = string(bline)
		if strip {
			line = string(bline[8:])
		}

		msg := line[30:]
		if len(msg) != 0 {
			msg = msg[1:]
		}

		t := line[:30]
		ti, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(t))
		if err != nil {
			log.Error(err)
		}

		w <- api.Ok(r, []docker.Log{{
			Timestamp: ti.UnixNano(),
			Message:   msg,
		}})
	}
}

func (a *Logs) Before(r *api.Request) api.Response {
	logs, _, err := docker.GetLogs(a.ContainerName, a.Amount, a.Since, a.Until, false)
	if err != nil {
		if err == docker.ErrContNotExist {
			return api.Error(r, er.Container+er.NotFound)
		}
		return api.Error(r, er.InternalServerError)

	}
	return api.Ok(r, logs)
}
