package models

import "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"

// JobRequest represents the structure of JSON body of the request from the client
type JobRequest struct {
	Script   string `json:"script"`             // Lua script
	Priority int    `json:"priority,omitempty"` // 0 = Low, 1 = Mid, 2 = High
}

// JobResponse represents the structure of JSON body of the response of the server
type JobResponse struct {
	JobID     int              `json:"job_id"`
	JobStatus models.JobStatus `json:"status"`
	JobResult string           `json:"result,omitempty"`
}
