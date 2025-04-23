package scheduler

import (
	models "DistributedJobScheduling/server/models"
	sharedModels "DistributedJobScheduling/shared/models"
	"sync"
)

// Scheduler is an object that manages tasks and workers. It stores all of them and assigns jobs to available workers.
type Scheduler struct {
	AvailableWorkers  WorkerQueue           // Queue that stores round-robin order of available (free) workers
	TotalWorkers      []sharedModels.Worker // Stores all registered workers: busy and available ones
	Jobs              JobQueues             // Queues that store jobs grouped by priority level
	ReceivedJobsCount int                   // A counter for amount of total number of jobs received from API, used for job ID generating
	mutex             sync.Mutex
}

// NewScheduler initializes new Scheduler object with empty queues
func NewScheduler() *Scheduler {
	return &Scheduler{
		AvailableWorkers:  *NewWorkerQueue(),
		Jobs:              *NewJobQueues(),
		ReceivedJobsCount: 0,
		mutex:             sync.Mutex{},
	}
}

// AssignTask chooses the worker to perform the job and assigns task to it, if there are so.
func (s *Scheduler) AssignTask() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	worker := s.RoundRobin() // exactly here it uses Round-robin to select the worker
	if worker != nil {       // enqueue task only if there are an available worker
		if task, ok := s.Jobs.Get(); ok {
			// TODO: transfer job to worker (functionality of `messaging` package)
			task.Priority = sharedModels.LowPriority // placeholder
		}
	}

}

// ReassignTask reassigns exactly selected task to new worker, because its old executor has been failed
func (s *Scheduler) ReassignTask(task models.Job) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	worker := s.RoundRobin()
	if worker != nil { // enqueue task only if there are an available worker
		// TODO: transfer job to worker (functionality of `messaging` package)
		task.Priority = sharedModels.LowPriority // placeholder
	}

}

// RoundRobin is the algorithm that AssignTask will use to select the worker.
// Workers are stored in queue, so it just dequeues one of them, ensuring the algorithm's logic
func (s *Scheduler) RoundRobin() *sharedModels.Worker {
	// apply mutex to lock the workers queue for other goroutines
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if worker, ok := s.AvailableWorkers.Get(); ok {
		return &worker
	}
	return nil
}

// EnqueueJob adds new job to jobs queue. Job is formed from passed priority level and script.
// Job ID generates as *total existing jobs amount* + 1.
func (s *Scheduler) EnqueueJob(priority sharedModels.JobPriority, script string) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	job := models.Job{
		JobID:    s.ReceivedJobsCount + 1,
		Priority: priority,
		Script:   script,
	}

	// add to queue
	s.Jobs.Add(job)
	s.ReceivedJobsCount++ // update counter

	return job.JobID, nil
}

// RegisterWorker adds new worker to the system.
func (s *Scheduler) RegisterWorker(worker sharedModels.Worker) {
	s.AvailableWorkers.Add(worker)                  // add to round-robin queue
	s.TotalWorkers = append(s.TotalWorkers, worker) // add to list of all workers
}
