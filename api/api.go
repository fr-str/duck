package api

import (
	"docker-project/api/actions"
	ws "docker-project/api/server"
	log "docker-project/logger"
	"net/http"
)

func Start(port string) error {

	ws.RegisterAction[actions.Containers]("container")
	ws.RegisterSubscription[actions.Containers]("live.containers")
	ws.RegisterSubscription[actions.Logs]("live.logs")
	ws.RegisterAction[actions.Logs]("logs")

	http.HandleFunc("/api", ws.Handler)
	log.Info("Listening on", port)

	return http.ListenAndServe(port, nil)
}
