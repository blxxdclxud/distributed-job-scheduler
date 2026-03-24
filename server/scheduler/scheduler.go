package scheduler

import (
	"context"
	"fmt"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/pkg/logger"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/messaging"
	models "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/models"
	sharedModels "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"
	"strconv"
	"sync"
	"time"
)

const SendJobTimeout = 5 * time.Second

// Scheduler is an object that manages tasks and workers. It stores all of them and assigns jobs to available workers.
type Scheduler struct {
	AvailableWorkers  WorkerQueue           // Queue that stores round-robin order of available (free) workers
	TotalWorkers      []sharedModels.Worker // Stores all registered workers: busy and available ones
	Jobs              JobQueues             // Queues that store jobs grouped by priority level
	ReceivedJobsCount int                   // A counter for amount of total number of jobs received from API, used for job ID generating
	mutex             sync.Mutex
	AllJobs           map[int]models.Job // Store all jobs
	rabbitClient      *messaging.Rabbit  // Client for messaging with workers
	WorkerAssignments map[string][]int   // Maps worker IDs to their assigned job IDs
}

// NewScheduler initializes new Scheduler object with empty queues
func NewScheduler() *Scheduler {
	return &Scheduler{
		AvailableWorkers:  *NewWorkerQueue(),
		Jobs:              *NewJobQueues(),
		ReceivedJobsCount: 0,
		mutex:             sync.Mutex{},
		AllJobs:           make(map[int]models.Job),
		WorkerAssignments: make(map[string][]int),
	}
}

// SetRabbitClient sets the RabbitMQ client for worker communication
func (s *Scheduler) SetRabbitClient(client *messaging.Rabbit) {
	s.rabbitClient = client
}

// AssignTask chooses the worker to perform the job and assigns task to it, if there are so.
func (s *Scheduler) assignTaskToWorkerUtil(task models.Job, workerId string) {
	if s.rabbitClient == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), SendJobTimeout)
	defer cancel()

	err := s.rabbitClient.SendTaskToWorker(ctx, task.Script, workerId, strconv.Itoa(task.JobID))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to send task to worker: %v", err))
		// Put the task back in queue
		s.Jobs.Add(task)
	} else {
		// Update job status
		job := s.AllJobs[task.JobID]
		job.Status = sharedModels.StatusRunning
		s.AllJobs[task.JobID] = job

		// Track new job assignment
		if _, ok := s.WorkerAssignments[workerId]; !ok {
			s.WorkerAssignments[workerId] = []int{}
		}
		s.WorkerAssignments[workerId] = append(s.WorkerAssignments[workerId], task.JobID)

		logger.Debug(fmt.Sprintf("Task %d was assigned to worker %s", task.JobID, workerId))
	}

}

// AssignTask chooses the worker to perform the job and assigns task to it, if there are so.
func (s *Scheduler) AssignTask() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	worker := s.RoundRobin() // exactly here it uses Round-robin to select the worker
	if worker != nil {       // enqueue task only if there are an available worker
		if task, ok := s.Jobs.Get(); ok {
			workerId := worker.ID
			logger.Debug(fmt.Sprintf("Assinging task %d to worker %s", task.JobID, workerId))
			s.assignTaskToWorkerUtil(task, workerId)
		}
	}
}

// ReassignTask reassigns exactly selected task to new worker, because its old executor has been failed
func (s *Scheduler) ReassignTask(task models.Job) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	worker := s.RoundRobin()
	if worker != nil { // enqueue task only if there are an available worker
		workerId := worker.ID
		logger.Debug(fmt.Sprintf("Reassinging task %d to worker %s", task.JobID, workerId))
		s.assignTaskToWorkerUtil(task, workerId)
	}
}

// RoundRobin is the algorithm that AssignTask will use to select the worker.
// Workers are stored in queue, so it just dequeues one of them, ensuring the algorithm's logic
func (s *Scheduler) RoundRobin() *sharedModels.Worker {
	// apply mutex to lock the workers queue for other goroutines

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

// GetJob returns the models.Job object by its ID
func (s *Scheduler) GetJob(jobID int) (models.Job, error) {
	// TODO: logic for this method
	return s.AllJobs[jobID], nil
}

// RegisterWorker adds new worker to the system.
func (s *Scheduler) RegisterWorker(worker sharedModels.Worker) {
	s.AvailableWorkers.Add(worker)                  // add to round-robin queue
	s.TotalWorkers = append(s.TotalWorkers, worker) // add to list of all workers
}
