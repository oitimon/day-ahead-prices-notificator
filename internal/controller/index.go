package controller

import "net/http"

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Welcome to DA price notificator!"))
}
