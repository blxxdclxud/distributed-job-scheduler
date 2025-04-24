package server

import (
	"DistributedJobScheduling/server/api"
	"DistributedJobScheduling/server/scheduler"
	"net/http"
)

// RunServer initializes all components of the server: API, scheduler, etc...
// ---- just placeholder now !!! ----
func RunServer() {
	sched := scheduler.NewScheduler()

	apiHandler := api.Handler{Scheduler: sched}
	router := api.RegisterRoutes(apiHandler)
	http.Handle("/", router)
}
