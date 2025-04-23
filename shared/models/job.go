package models

type JobPriority int

const (
	HighPriority JobPriority = iota + 1
	MidPriority
	LowPriority
)

// JobPriorities contains all priorities sorted in ascending order
var JobPriorities = []JobPriority{HighPriority, MidPriority, LowPriority}

type JobStatus string

const (
	StatusRunning   JobStatus = "RUNNING"
	StatusPending   JobStatus = "PENDING"
	StatusCompleted JobStatus = "COMPLETED"
	StatusFailed    JobStatus = "FAILED"
)

//type Job struct {
//}
