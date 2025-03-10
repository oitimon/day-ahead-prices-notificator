package controller

import "net/http"

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("healthy"))
}
