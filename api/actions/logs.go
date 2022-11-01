package actions

import (
	"bufio"
	ws "docker-project/api/server"
	"docker-project/docker"
	"docker-project/er"
	log "docker-project/logger"
	"io"
	"strings"
	"time"
)

type Logs struct {
	ContainerName string
	Amount        int
	Since         int64
	Until         int64
}

func (a *Logs) Handle(r *ws.Request) ws.Response {
	if a.ContainerName == "" {
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

// What a mess lol
func (a *Logs) HandleSub(r *ws.Request, w chan<- ws.Response) {
	cont, ok := docker.ContainerMap.GetFull(a.ContainerName)
	if !ok {
		w <- ws.Error(r, er.NotFound+er.Container)
		return
	}

	logs, _, err := docker.GetLogs(a.ContainerName, a.Amount, a.Since, 0, false)
	if err != nil {
		if err == docker.ErrContNotExist {
			log.Debug(er.Container.String() + er.NotFound.String())
			w <- ws.Error(r, er.Container+er.NotFound)
			return
		}

		log.Error(er.InternalServerError.String())
		w <- ws.Error(r, er.InternalServerError)
		return
	}
	w <- ws.Ok(r, logs)

	var strip bool
	if !docker.ContainerMap.Get(a.ContainerName).Tty {
		strip = true
	}
	var rc io.ReadCloser
	for {
		select {
		case <-r.Ctx.Done():
			if rc != nil {
				rc.Close()
			}
			return
		default:
		}
		if cont.State == "exited" {
			time.Sleep(1 * time.Second)
			log.Debug(`cont.State == "exited" `)
			continue
		}
		_, rc, _ = docker.GetLogs(a.ContainerName, 0, a.Since, 0, true)
		log.Debug("reading", a.ContainerName)
		readL(r, rc, w, strip)
		time.Sleep(50 * time.Millisecond)
	}
}

func readL(r *ws.Request, rc io.ReadCloser, w chan<- ws.Response, strip bool) {
	var line string
	var bline []byte
	sc := bufio.NewScanner(rc)
	for {
		select {
		case <-r.Ctx.Done():
			rc.Close()
			return
		default:
		}

		if !sc.Scan() {
			return
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

		w <- ws.Ok(r, []docker.Log{{
			Timestamp: ti.UnixNano(),
			Message:   msg,
		}})
	}
}

func (a *Logs) Before(r *ws.Request) ws.Response {
	logs, _, err := docker.GetLogs(a.ContainerName, a.Amount, a.Since, a.Until, false)
	if err != nil {
		if err == docker.ErrContNotExist {
			log.Debug(er.Container.String() + er.NotFound.String())
			return ws.Error(r, er.Container+er.NotFound)
		}
		log.Error(er.InternalServerError.String())
		return ws.Error(r, er.InternalServerError)

	}
	return ws.Ok(r, logs)
}
