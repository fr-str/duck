package actions

import (
	"bufio"
	ws "docker-project/api/server"
	"docker-project/docker"
	log "docker-project/logger"
	"docker-project/structs"
	"docker-project/wsc"
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
				w <- ws.Error(r, wsc.InternalServerError)
			}
			go c.HandleLive(r, w)

		case "logs":
			var l Logs
			if err := json.Unmarshal(v, &l); err != nil {
				log.Error(err)
				w <- ws.Error(r, wsc.InternalServerError)
			}
			go l.HandleLive(r, w)

		default:
			log.Error(wsc.Action.String() + wsc.NotFound.String())
			w <- ws.Error(r, wsc.Action+wsc.NotFound)
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

	contEvents := docker.ContainerMap.Register(r.Ctx)
	for v := range contEvents {
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
	allLogs := make([]docker.Log, 0)
	for _, cName := range a.ContainerNames {
		_, ok := docker.ContainerMap.GetFull(cName)
		if !ok {
			w <- ws.Error(r, wsc.NotFound+wsc.Container)
			continue
		}

		logs, _, err := docker.GetLogs(cName, a.Amount, a.Since, 0, false)
		if err != nil {
			if err == docker.ErrContNotExist {
				log.Debug(wsc.Container.String() + wsc.NotFound.String())
				w <- ws.Error(r, wsc.Container+wsc.NotFound)
				continue
			}

			log.Error(wsc.InternalServerError.String())
			w <- ws.Error(r, wsc.InternalServerError, "Reading "+cName+" logs")
			continue
		}
		allLogs = append(allLogs, logs...)
	}

	sort.Slice(allLogs, func(i, j int) bool {
		return allLogs[i].Timestamp > allLogs[j].Timestamp
	})

	w <- ws.Live("logs", allLogs)

	for _, cName := range a.ContainerNames {
		go a.streamLogs(r, w, cName)
	}
}

func (a *Logs) streamLogs(r *ws.Request, w chan<- ws.Response, containerName string) {
	if !docker.ContainerMap.Exists(containerName) {
		return
	}
	// FIXME: this is a mess

	var strip bool
	if !docker.ContainerMap.Get(containerName).Tty {
		strip = true
	}

	var rc io.ReadCloser
	var err error
	for {
		select {
		case <-r.Ctx.Done():
			if rc != nil {
				rc.Close()
			}
			return
		default:
		}

		cont, ok := docker.ContainerMap.GetFull(containerName)
		if !ok {
			w <- ws.Error(r, wsc.NotFound+wsc.Container)
			return
		}

		if cont.State == "exited" {
			time.Sleep(1 * time.Second)
			log.Debug(containerName, `cont.State == "exited" `)
			w <- ws.Error(r, wsc.Error+wsc.Container+wsc.Exited, containerName)
			return
		}

		_, rc, err = docker.GetLogs(containerName, 0, a.Since, 0, true)
		if err != nil {
			if err == docker.ErrContNotExist {
				log.Debug(wsc.Container.String() + wsc.NotFound.String())
				w <- ws.Error(r, wsc.Container+wsc.NotFound)
				return
			}

			log.Error(wsc.InternalServerError.String())
			w <- ws.Error(r, wsc.InternalServerError, "Reading "+containerName+" logs")
			return
		}

		log.Debug("reading", containerName)

		readL(r, rc, w, containerName, strip)
		time.Sleep(50 * time.Millisecond)
	}
}

func readL(r *ws.Request, rc io.ReadCloser, w chan<- ws.Response, containerName string, strip bool) {

	var line string
	var bline []byte
	scanChan := make(chan bool)
	canScanChan := make(chan bool)
	sc := bufio.NewScanner(rc)
	go func() {
		for {
			select {
			case <-r.Ctx.Done():
				close(scanChan)
				return
			case scanChan <- sc.Scan():
				if !<-canScanChan { // wait for the scan to be done
					return
				}
			}
		}
	}()

	for {
		select {
		case <-r.Ctx.Done():
			rc.Close()
			return
		case b := <-scanChan:
			if !b {
				canScanChan <- false
				return
			}
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
			Message:   docker.CutTimestamp(msg),
			Container: containerName,
		}})
		canScanChan <- true
	}
}
