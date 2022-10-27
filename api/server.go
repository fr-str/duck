package api

import (
	log "docker-project/logger"
	"net/http"
)

func Start(port string) error {

	http.HandleFunc("/api", Handler)
	log.Info("Listening on", port)
	return http.ListenAndServe(port, nil)
}
