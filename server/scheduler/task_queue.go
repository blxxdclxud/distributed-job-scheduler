package scheduler

import (
	"github.com/golang-collections/collections/queue"
	models2 "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/models"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"
)

// JobQueues stores all tasks in corresponding queues, grouped by priority level. Acts as a map.
// e.g.: High: high_priority_queue
//
//	Low: low_priority_queue
type JobQueues map[models.JobPriority]*queue.Queue

// NewJobQueues initialize new JobQueues object with empty queues for all existing priority levels
func NewJobQueues() *JobQueues {
	tq := JobQueues{}
	for _, priorityLevel := range models.JobPriorities {
		tq[priorityLevel] = queue.New()
	}

	return &tq
}

// Add appends given task to the corresponding queue according to task's priority
func (t *JobQueues) Add(job models2.Job) {
	(*t)[job.Priority].Enqueue(job)
}

// Get returns the next task to be performed. The queue structure ensures the correct order as tasks appeared in it,
// and the queues with higher priority level will be checked first.
// In other words, the higher the priority level, the sooner the queue will be served.
// Returns nil if no tasks exist.
func (t *JobQueues) Get() (models2.Job, bool) {
	for _, priority := range models.JobPriorities {
		if q := (*t)[priority]; q.Len() > 0 {
			job := q.Dequeue().(models2.Job)
			return job, true
		}
	}
	return models2.Job{}, false
}
