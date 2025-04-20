package models

import "DistributedJobScheduling/shared/models"

type JobRequest struct {
	Script   string `json:"script"`             // Lua script
	Priority int    `json:"priority,omitempty"` // 0 = Low, 1 = High
}

type JobResponse struct {
	JobID     int              `json:"job_id"`
	JobStatus models.JobStatus `json:"status"`
	JobResult string           `json:"result,omitempty"`
}
