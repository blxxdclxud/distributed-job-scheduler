package models

import "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"

type Job struct {
	JobID    int
	Priority models.JobPriority
	Script   string
	Status   models.JobStatus
	Result   string // Add result field
}
