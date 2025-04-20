package api

import (
	"DistributedJobScheduling/server/models"
	sharedModels "DistributedJobScheduling/shared/models"
	"encoding/json"
	"net/http"
)

// SubmitJobHandler is handler that accepts the job submitted by a client.
// It passes the job to the scheduler in case of successful
func SubmitJobHandler(w http.ResponseWriter, r *http.Request) {
	var jobRequest models.JobRequest
	if err := json.NewDecoder(r.Body).Decode(&jobRequest); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid request format")
		return
	}

	var jobID int
	//id, err := pass_to_scheduler()
	//if err != nil {
	//	ErrorResponse(w, http.StatusInternalServerError, "failed to pass the jobRequest to scheduler")
	// return
	//}

	ResponseJson(w, http.StatusAccepted, models.JobResponse{
		JobID:     jobID,
		JobStatus: sharedModels.StatusPending,
	})
}

// GetJobStatusHandler is handler that accepts the jos id from a client to check corresponding job's status.
// ID is passed in url.
func GetJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	//jobID := mux.Vars(r)["id"]
	//status, err := get_status()
	//if err != nil {
	//	ErrorResponse(w, http.StatusBadRequest, "job not found")
	//	return
	//}

	//ResponseJson(w, http.StatusOK, status)
}
