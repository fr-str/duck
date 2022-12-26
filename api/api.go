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

	ws.RegisterAction[actions.Containers]("container")
	ws.RegisterSubscription[actions.Live]("live")
	ws.RegisterAction[actions.Logs]("logs")
	time.Sleep(500 * time.Millisecond)
	r := mux.NewRouter()
	r.HandleFunc("/api", ws.Handler)

	// r.PathPrefix("/").Handler(http.FileServer(http.Dir("./front/build")))
	log.Info("Listening on", port)

	return http.ListenAndServe(port, r)
}
