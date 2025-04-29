package scheduler

import (
	"context"
	"fmt"
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
	ReceivedJobsCount int                   // A counter for amount of total number of jobs received from API
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
	s.mutex.Lock()
	defer s.mutex.Unlock()

	worker := s.roundRobinUnlocked() // Use unlocked version to avoid deadlock
	if worker != nil {
		if task, ok := s.Jobs.Get(); ok {
			// Send the job to the worker using messaging
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Get worker ID - this assumes the Worker struct has a field or method to get ID
			workerId := worker.GetID()

			logger.Debug(fmt.Sprintf("Assinging task %d to worker %s", task.JobID, workerId))

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

				// Track this assignment
				if _, exists := s.WorkerAssignments[workerId]; !exists {
					s.WorkerAssignments[workerId] = []int{}
				}
				s.WorkerAssignments[workerId] = append(s.WorkerAssignments[workerId], task.JobID)

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

		logger.Debug(fmt.Sprintf("Reasigning task %d to worker %s", task.JobID, workerId))

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

// UpdateJob updates an existing job in the scheduler with new status and result
func (s *Scheduler) UpdateJob(jobID int, status sharedModels.JobStatus, result string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	job, exists := s.AllJobs[jobID]
	if !exists {
		return fmt.Errorf("job with ID %d not found", jobID)
	}

	job.Status = status
	job.Result = result
	s.AllJobs[jobID] = job

	// If the job is now complete or failed, remove it from worker assignments
	if status == sharedModels.StatusCompleted || status == sharedModels.StatusFailed {
		// Find which worker had this job
		for workerId, assignments := range s.WorkerAssignments {
			for i, id := range assignments {
				if id == jobID {
					// Remove this job from the assignments
					s.WorkerAssignments[workerId] = append(assignments[:i], assignments[i+1:]...)
					break
				}
			}
		}
	}

	logger.Debug(fmt.Sprintf("Updated job %d, new status: %s", jobID, status))
	return nil
}

// RegisterWorker adds new worker to the system.
func (s *Scheduler) RegisterWorker(worker sharedModels.Worker) {
	s.AvailableWorkers.Add(worker)                  // add to round-robin queue
	s.TotalWorkers = append(s.TotalWorkers, worker) // add to list of all workers
}

// RemoveWorker removes a worker from the system by its ID and reschedules its tasks
func (s *Scheduler) RemoveWorker(workerID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger.Debug(fmt.Sprintf("Removing worker: %s", workerID))

	// Check if there are any tasks assigned to this worker
	jobIDs, workerHasJobs := s.WorkerAssignments[workerID]

	// Reschedule all jobs assigned to this worker
	if workerHasJobs {
		logger.Info(fmt.Sprintf("Rescheduling %d jobs from failed worker %s", len(jobIDs), workerID))
		for _, jobID := range jobIDs {
			if job, exists := s.AllJobs[jobID]; exists && job.Status == sharedModels.StatusRunning {
				logger.Debug(fmt.Sprintf("Rescheduling job %d from failed worker", jobID))

				// Reset job status to pending
				job.Status = sharedModels.StatusPending
				s.AllJobs[jobID] = job

				// Re-add to job queue
				s.Jobs.Add(job)
			}
		}

		// Clear the assignments for this worker
		delete(s.WorkerAssignments, workerID)
	}

	// Create a new queue without the specified worker
	newQueue := NewWorkerQueue()

	// Get all the workers from existing queue
	removedCount := 0
	totalWorkers := s.AvailableWorkers.Size()

	for i := 0; i < totalWorkers; i++ {
		if worker, ok := s.AvailableWorkers.Get(); ok {
			if worker.GetID() != workerID {
				// Keep this worker
				newQueue.Add(worker)
			} else {
				removedCount++
			}
		}
	}

	// Replace the old queue with the filtered one
	s.AvailableWorkers = *newQueue

	// Also remove from TotalWorkers slice
	newTotalWorkers := make([]sharedModels.Worker, 0, len(s.TotalWorkers)-removedCount)
	for _, worker := range s.TotalWorkers {
		if worker.GetID() != workerID {
			newTotalWorkers = append(newTotalWorkers, worker)
		}
	}
	s.TotalWorkers = newTotalWorkers

	logger.Info(fmt.Sprintf("Removed worker %s and rescheduled its tasks", workerID))

	if removedCount == 0 && !workerHasJobs {
		return fmt.Errorf("worker %s not found in the pool", workerID)
	}

	return nil
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
