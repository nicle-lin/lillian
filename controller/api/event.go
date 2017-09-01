package api

import "net/http"

func (a *Api) events(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	//TODO: show all event
}

func (a *Api) purgeEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	//TODO: purge all event

	w.WriteHeader(http.StatusNoContent)
}
