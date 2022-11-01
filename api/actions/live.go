package actions

import (
	"bufio"
	ws "docker-project/api/server"
	"docker-project/docker"
	"docker-project/er"
	log "docker-project/logger"
	"docker-project/structs"
	"encoding/json"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/timoni-io/go-utils/types"
)

type Live map[string]json.RawMessage

func (a Live) HandleSub(r *ws.Request, w chan<- ws.Response) {
	for k, v := range a {
		switch strings.ToLower(k) {
		case "containers":
			var c Containers
			if err := json.Unmarshal(v, &c); err != nil {
				log.Error(err)
				w <- ws.Error(r, er.InternalServerError)
			}
			go c.HandleLive(r, w)

		case "logs":
			var l Logs
			if err := json.Unmarshal(v, &l); err != nil {
				log.Error(err)
				w <- ws.Error(r, er.InternalServerError)
			}
			go l.HandleLive(r, w)

		default:
			log.Error(er.Action.String() + er.NotFound.String())
			w <- ws.Error(r, er.Action+er.NotFound)
		}
	}
}

func (a *Containers) HandleLive(r *ws.Request, w chan<- ws.Response) {
	for el := range docker.ContainerMap.Iter() {
		w <- ws.Live("containers", types.WatchMsg[string, structs.Container]{
			Event: types.PutEvent,
			Item:  el,
		})
	}

	for v := range docker.ContainerMap.Register() {
		select {
		case <-r.Ctx.Done():
			return
		case w <- ws.Live("containers", v):
		}
	}
}

// What a mess lol
func (a *Logs) HandleLive(r *ws.Request, w chan<- ws.Response) {
	log.Debug("streaming logs for", a.ContainerNames)
	for _, cName := range a.ContainerNames {
		go a.streamLogs(r, w, cName)
	}
}

func (a *Logs) streamLogs(r *ws.Request, w chan<- ws.Response, containerName string) {
	cont, ok := docker.ContainerMap.GetFull(containerName)
	if !ok {
		w <- ws.Error(r, er.NotFound+er.Container)
		return
	}

	logs, _, err := docker.GetLogs(containerName, a.Amount, a.Since, 0, false)
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
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp > logs[j].Timestamp
	})
	w <- ws.Live("logs", logs)

	var strip bool
	if !docker.ContainerMap.Get(containerName).Tty {
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
		_, rc, _ = docker.GetLogs(containerName, 0, a.Since, 0, true)
		log.Debug("reading", containerName)
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
		if len(bline) < 8 {
			continue
		}

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

		w <- ws.Live("logs", []docker.Log{{
			Timestamp: ti.UnixNano(),
			Message:   msg,
		}})
	}
}
