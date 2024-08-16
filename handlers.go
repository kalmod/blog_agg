package main

import (
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	type statusJSON struct {
		Status string `json:"status"`
	}
	respondWithJSON(w, 200, statusJSON{Status: "ok"})
}

func errorHandler(w http.ResponseWriter, r *http.Request) {

	respondWithError(w, 500, "Internal Server Error")
}
