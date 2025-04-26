package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/pkg/logger"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/models"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/scheduler"
	sharedModels "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"
)

// Handler stores Scheduler instance as field, that allows to pass new arrived jobs to it
type Handler struct {
	Scheduler *scheduler.Scheduler
}

// SubmitJobHandler is handler that accepts the job submitted by a client.
// It passes the job to the Scheduler in case of successful
func (h *Handler) SubmitJobHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug("got a job submission request...")
	var jobRequest models.JobRequest
	if err := json.NewDecoder(r.Body).Decode(&jobRequest); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid request format")
		return
	}

	jobID, err := h.Scheduler.EnqueueJob(sharedModels.JobPriority(jobRequest.Priority), jobRequest.Script)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "failed to pass the jobRequest to Scheduler")
		return
	}

	ResponseJson(w, http.StatusAccepted, models.JobResponse{
		JobID:     jobID,
		JobStatus: sharedModels.StatusPending,
	})
}

// GetJobStatusHandler is handler that accepts the jos id from a client to check corresponding job's status.
// ID is passed in url.
func (h *Handler) GetJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug("got a job status request...")
	jobID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "job not found")
		return
	}
	status, err := h.Scheduler.GetJob(jobID)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "job not found")
		return
	}

	ResponseJson(w, http.StatusOK, status)
}
