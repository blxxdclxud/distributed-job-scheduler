package models

type JobPriority int

const (
	LowPriority JobPriority = iota
	HighPriority
)

type JobStatus string

const (
	StatusRunning   JobStatus = "RUNNING"
	StatusPending   JobStatus = "PENDING"
	StatusCompleted JobStatus = "COMPLETED"
	StatusFailed    JobStatus = "FAILED"
)

type Job struct {
}
