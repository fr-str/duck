package actions

import (
	"bufio"
	ws "docker-project/api/server"
	log "docker-project/logger"
	"io"
	"time"

	"docker-project/docker"
	"docker-project/structs"
	"docker-project/wsc"
	"strings"

	"github.com/timoni-io/go-utils/cmd"
	"github.com/timoni-io/go-utils/types"
)

type Containers struct {
	Name         string
	ComposePath  string
	ComposeBuild bool
	SoftForce    bool
	Force        bool
}

func (a *Containers) Handle(r *ws.Request) ws.Response {
	act := strings.TrimPrefix(r.Action, "container.")
	switch act {
	case "stop", "restart", "start", "kill":
		return a.SSRK(r)
	case "list":
		return a.List(r)
	case "compose":
		return a.ApplyDockerCompose(r)
	case "create":
		return a.Create(r)
	case "inspect":
		return a.Inspect(r)
	case "delete":
		return a.Delete(r)

	default:
		log.Error(wsc.Action.String() + wsc.NotFound.String())
		return ws.Error(r, wsc.Action+wsc.NotFound)
	}
}

func (a *Containers) SSRK(r *ws.Request) ws.Response {
	cont, ok := docker.ContainerMap.GetFull(a.Name)
	if !ok {
		return ws.Error(r, wsc.NotFound+wsc.Container)
	}

	switch strings.TrimPrefix(r.Action, "container.") {
	case "start":
		ws.GoError(r, wsc.InternalServerError, cont.Start)
	case "stop":
		ws.GoError(r, wsc.InternalServerError, cont.Stop)
	case "restart":
		ws.GoError(r, wsc.InternalServerError, cont.Restart)
	case "kill":
		ws.GoError(r, wsc.InternalServerError, cont.Kill)
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

func (a *Containers) Create(r *ws.Request) ws.Response {
	_, ok := docker.ContainerMap.GetFull(a.Name)
	if ok {
		return ws.Error(r, wsc.Exists+wsc.Container)
	}
	//TODO

	return ws.Ok(r, "ok")
}

func (a *Containers) Inspect(r *ws.Request) ws.Response {
	log.Debug("Inspect", a.Name)
	cont, ok := docker.DockerContainerMap.GetFull(a.Name)
	if !ok {
		return ws.Error(r, wsc.NotFound+wsc.Container)
	}
	return ws.Custom(r, wsc.OK+wsc.Inspect, cont)
}

func (a *Containers) ApplyDockerCompose(r *ws.Request) ws.Response {
	log.Debug("ApplyDockerCompose", a.ComposePath)
	//TODO support custom command
	args := []string{"up", "-d"}
	pr := func(stdout io.ReadCloser, stderr io.ReadCloser) {
		go readLogs(r, stderr)
		readLogs(r, stdout)
	}

	if a.ComposeBuild {
		args = append(args, "--build")
	}

	ws.GoError(r, wsc.InternalServerError, func() error {
		log.Info("docker-compose", args)
		log.Info(pr)
		return cmd.NewCommand("docker-compose", args...).Run(&cmd.RunOptions{
			Dir:        a.ComposePath,
			PipeReader: pr,
			Timeout:    0,
		})
	})

	return ws.Ok(r, "ok")
}

func readLogs(r *ws.Request, rc io.ReadCloser) {
	var line string
	sc := bufio.NewScanner(rc)
	for sc.Scan() {
		line = sc.Text()
		if line == "" {
			continue
		}
		log.Debug(line)

		r.ResultCh <- ws.Ok(r, []docker.Log{{
			Timestamp: time.Now().UnixNano(),
			Message:   line,
		}})
	}
	r.ResultCh <- ws.Ok(r, []docker.Log{{
		Timestamp: time.Now().UnixNano(),
		Message:   "End of stream",
	}})
}

func (a *Containers) Delete(r *ws.Request) ws.Response {
	cont, ok := docker.ContainerMap.GetFull(a.Name)
	if !ok {
		return ws.Error(r, wsc.NotFound+wsc.Container)
	}

	if !a.Force && !a.SoftForce && cont.State != "exited" {
		return ws.Error(r, wsc.Forbbiden+wsc.ContainerIsRunning)
	}

	if a.SoftForce {
		ws.GoError(r, wsc.InternalServerError, cont.Stop, cont.Delete)
	} else {
		ws.GoError(r, wsc.InternalServerError, cont.Kill, cont.Delete)
	}

	return ws.Ok(r, "ok")
}
