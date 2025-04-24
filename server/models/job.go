package models

import "DistributedJobScheduling/shared/models"

type Job struct {
	JobID    int
	Priority models.JobPriority
	Script   string
}
