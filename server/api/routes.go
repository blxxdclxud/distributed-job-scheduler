package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

func RegisterRoutes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/submit_job", SubmitJobHandler).Methods(http.MethodPost)
	router.HandleFunc("/status/{id}", GetJobStatusHandler).Methods(http.MethodGet)

	return router
}
