package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

// RegisterRoutes creates a router mux.Router that has endpoints of the API assigned to it
func RegisterRoutes(handler Handler) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/submit_job", handler.SubmitJobHandler).Methods(http.MethodPost)
	router.HandleFunc("/status/{id}", handler.GetJobStatusHandler).Methods(http.MethodGet)

	return router
}
