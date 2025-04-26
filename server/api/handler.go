package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

	logger.Debug("job submission ok")

	ResponseJson(w, http.StatusAccepted, models.JobResponse{
		JobID:     jobID,
		JobStatus: sharedModels.StatusPending,
	})
}

// GetJobStatusHandler is handler that accepts the jos id from a client to check corresponding job's status.
// ID is passed in url.
func (h *Handler) GetJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger.Debug("Starting job status request...")

	jobID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		logger.Error("Invalid job ID: " + err.Error())
		ErrorResponse(w, http.StatusBadRequest, "invalid job ID")
		return
	}

	logger.Debug(fmt.Sprintf("Getting job %d", jobID))
	job, err := h.Scheduler.GetJob(jobID)
	if err != nil {
		logger.Error("Job not found: " + err.Error())
		ErrorResponse(w, http.StatusNotFound, "job not found")
		return
	}

	// Convert internal Job to JobResponse for proper JSON serialization
	response := models.JobResponse{
		JobID:     job.JobID,
		JobStatus: job.Status,
		JobResult: job.Result,
	}

	logger.Debug(fmt.Sprintf("Found job %d with status %s", jobID, job.Status))
	ResponseJson(w, http.StatusOK, response)

	logger.Debug(fmt.Sprintf("Request completed in %v", time.Since(start)))
}
