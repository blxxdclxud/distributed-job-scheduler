package server

import (
	"net/http"

	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/pkg/logger"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/api"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/scheduler"
)

// RunServer initializes all components of the server: API, scheduler, etc...
// ---- just placeholder now !!! ----
func RunServer() {
	sched := scheduler.NewScheduler()

	apiHandler := api.Handler{Scheduler: sched}
	router := api.RegisterRoutes(apiHandler)

	logger.Debug("Starting server on :8080")

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		logger.Error("Failed to start server.")
	}
}
