package api

import (
	"docker-project/api/actions"
	ws "docker-project/api/server"
	"docker-project/docker"
	dcli "docker-project/docker/client"
	log "docker-project/logger"
	"net/http"
	"time"
	"github.com/gorilla/mux"
)

func Start(port string) error {
	dcli.Init()
	go docker.UpdateMap(dcli.Cli)
	go docker.HandleEvents(dcli.Cli)

	ws.RegisterAction[actions.Containers]("container")
	ws.RegisterSubscription[actions.Live]("live")
	ws.RegisterAction[actions.Logs]("logs")
	time.Sleep(time.Second)
	r := mux.NewRouter()
	r.HandleFunc("/api", ws.Handler)

	// r.PathPrefix("/").Handler(http.FileServer(http.Dir("./front/build")))
	log.Info("Listening on", port)

	return http.ListenAndServe(port, r)
}
