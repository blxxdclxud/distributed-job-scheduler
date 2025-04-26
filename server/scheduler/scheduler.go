package scheduler

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/pkg/logger"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/messaging"
	models "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/models"
	sharedModels "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"
)

// Scheduler is an object that manages tasks and workers. It stores all of them and assigns jobs to available workers.
type Scheduler struct {
	AvailableWorkers  WorkerQueue           // Queue that stores round-robin order of available (free) workers
	TotalWorkers      []sharedModels.Worker // Stores all registered workers: busy and available ones
	Jobs              JobQueues             // Queues that store jobs grouped by priority level
	ReceivedJobsCount int                   // A counter for amount of total number of jobs received from API, used for job ID generating
	mutex             sync.Mutex
	AllJobs           map[int]models.Job // Store all jobs
	rabbitClient      *messaging.Rabbit  // Client for messaging with workers
}

// NewScheduler initializes new Scheduler object with empty queues
func NewScheduler() *Scheduler {
	return &Scheduler{
		AvailableWorkers:  *NewWorkerQueue(),
		Jobs:              *NewJobQueues(),
		ReceivedJobsCount: 0,
		mutex:             sync.Mutex{},
		AllJobs:           make(map[int]models.Job),
	}
}

// SetRabbitClient sets the RabbitMQ client for worker communication
func (s *Scheduler) SetRabbitClient(client *messaging.Rabbit) {
	s.rabbitClient = client
}

// Create a private method without mutex lock
func (s *Scheduler) roundRobinUnlocked() *sharedModels.Worker {
	if worker, ok := s.AvailableWorkers.Get(); ok {
		return &worker
	}
	return nil
}

// Public method with mutex lock
func (s *Scheduler) RoundRobin() *sharedModels.Worker {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.roundRobinUnlocked()
}

// AssignTask chooses the worker to perform the job and assigns task to it, if there are so.
func (s *Scheduler) AssignTask() {
	log.Printf("Begin assigning the task to a worker")
	s.mutex.Lock()
	defer s.mutex.Unlock()

	worker := s.roundRobinUnlocked() // Use unlocked version to avoid deadlock
	log.Printf("got worker")
	if worker != nil {
		log.Printf("worker is not nil")
		if task, ok := s.Jobs.Get(); ok {
			log.Printf("setting context")

			// Send the job to the worker using messaging
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			log.Printf("getting worker id")

			// Get worker ID - this assumes the Worker struct has a field or method to get ID
			workerId := worker.GetID()

			log.Printf("sending the task to the worker")

			err := s.rabbitClient.SendTaskToWorker(ctx, task.Script, workerId, strconv.Itoa(task.JobID))
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to send task to worker: %v", err))
				// Put the task back in queue
				s.Jobs.Add(task)
			} else {
				// Update job status to running
				job := s.AllJobs[task.JobID]
				job.Status = sharedModels.StatusRunning
				s.AllJobs[task.JobID] = job

				logger.Debug(fmt.Sprintf("Task %d assigned to worker %s", task.JobID, workerId))
			}
		}
	}
}

// ReassignTask reassigns exactly selected task to new worker, because its old executor has been failed
func (s *Scheduler) ReassignTask(task models.Job) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	worker := s.roundRobinUnlocked() // Use unlocked version
	if worker != nil {               // enqueue task only if there are an available worker
		// Send the job to the worker using messaging
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Get worker ID
		workerId := worker.GetID() // You might need to implement this method

		err := s.rabbitClient.SendTaskToWorker(ctx, task.Script, workerId, strconv.Itoa(task.JobID))
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to reassign task to worker: %v", err))
			// Put the task back in queue
			s.Jobs.Add(task)
		} else {
			// Update job status to running
			job := s.AllJobs[task.JobID]
			job.Status = sharedModels.StatusRunning
			s.AllJobs[task.JobID] = job

			logger.Debug(fmt.Sprintf("Task %d reassigned to worker %s", task.JobID, workerId))
		}
	}
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
		Status:   sharedModels.StatusPending, // Set initial status to pending
	}

	// add to queue
	s.Jobs.Add(job)
	s.ReceivedJobsCount++ // update counter

	// Store in AllJobs map for tracking
	s.AllJobs[job.JobID] = job

	return job.JobID, nil
}

// GetJob returns the models.Job object by its ID
func (s *Scheduler) GetJob(jobID int) (models.Job, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock() // This ensures mutex is always unlocked

	job, exists := s.AllJobs[jobID]
	if !exists {
		return models.Job{}, fmt.Errorf("job with ID %d not found", jobID)
	}

	return job, nil
}

// UpdateJob updates an existing job in the scheduler
func (s *Scheduler) UpdateJob(jobID int, job models.Job) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.AllJobs[jobID]
	if !exists {
		return fmt.Errorf("job with ID %d not found", jobID)
	}

	s.AllJobs[jobID] = job
	logger.Debug(fmt.Sprintf("Updated job %d, new status: %s", jobID, job.Status))
	return nil
}

// RegisterWorker adds new worker to the system.
func (s *Scheduler) RegisterWorker(worker sharedModels.Worker) {
	s.AvailableWorkers.Add(worker)                  // add to round-robin queue
	s.TotalWorkers = append(s.TotalWorkers, worker) // add to list of all workers
}

// StartTaskProcessing begins a goroutine to continuously process tasks
func (s *Scheduler) StartTaskProcessing() {
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			s.AssignTask() // Try to assign tasks periodically
		}
	}()
	logger.Debug("Task processing started")
}
