package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/pkg/logger"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/api"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/messaging"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/scheduler"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models/Rabbit"
)

// RunServer initializes all components of the server: API, scheduler, etc...
// Now takes the RabbitMQ host address as a parameter
func RunServer(rmqHost string) {
	logger.Debug("Connecting to RabbitMQ at " + rmqHost)

	// Connect to RabbitMQ
	conn, err := amqp.Dial(rmqHost)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ: " + err.Error())
	}
	defer conn.Close()

	// Initialize scheduler
	sched := scheduler.NewScheduler()

	// Create RabbitMQ client
	rabbitClient, err := messaging.NewRabbit(conn)
	if err != nil {
		logger.Fatal("Failed to create RabbitMQ client: " + err.Error())
	}

	// Set RabbitMQ client in scheduler
	sched.SetRabbitClient(rabbitClient)

	// Create channels for RabbitMQ communication
	healthReportCh := make(chan Rabbit.HealthReportWrapper, 100)
	registrationCh := make(chan Rabbit.RegistrationWrapper, 10)
	taskResultCh := make(chan Rabbit.TaskReplyWrapper, 100)

	// Start listeners
	go rabbitClient.ListenHeartBeat(healthReportCh)
	go rabbitClient.ListenRegister(registrationCh)
	go rabbitClient.ListenTaskResults(taskResultCh)

	// Process worker registrations
	go func() {
		for reg := range registrationCh {
			if reg.Err != nil {
				logger.Error("Error in worker registration: " + reg.Err.Error())
				continue
			}

			logger.Debug("Registering worker: " + reg.WorkerId)
			worker := models.Worker{} // Create a worker struct
			worker.SetWorkerId(reg.WorkerId)
			sched.RegisterWorker(worker)
		}
	}()

	// Process task results
	go func() {
		for result := range taskResultCh {
			if result.Err != nil {
				logger.Error("Error in task result: " + result.Err.Error())
				continue
			}

			jobID, err := strconv.Atoi(result.TaskReply.JobId)
			if err != nil {
				logger.Error("Invalid job ID in result: " + result.TaskReply.JobId)
				continue
			}

			logger.Debug(fmt.Sprintf("Received result for job %d from worker %s",
				jobID, result.TaskReply.WorkerId))

			var status models.JobStatus
			var resultStr string

			if result.TaskReply.Err != "" {
				status = models.StatusFailed
				resultStr = result.TaskReply.Err
				logger.Warn(fmt.Sprintf("Job %d failed: %s", jobID, resultStr))
			} else {
				status = models.StatusCompleted
				resultStr = result.TaskReply.Results
				logger.Info(fmt.Sprintf("Job %d completed successfully with result: %s",
					jobID, resultStr))
			}

			// Update job status with the result
			err = sched.UpdateJob(jobID, status, resultStr)
			if err != nil {
				logger.Error("Failed to update job: " + err.Error())
			}
		}
	}()

	// Process heartbeats from workers
	go func() {
		workerLastSeen := make(map[string]time.Time)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case heartbeat := <-healthReportCh:
				if heartbeat.Err != nil {
					continue
				}
				workerLastSeen[heartbeat.HealthReport.WorkerId] = time.Now()

			case <-ticker.C:
				// Check for workers that haven't sent a heartbeat in a while
				threshold := time.Now().Add(-30 * time.Second)
				for workerID, lastSeen := range workerLastSeen {
					if lastSeen.Before(threshold) {
						logger.Warn(fmt.Sprintf("Worker %s hasn't sent heartbeat in >30s, removing...", workerID))

						// Remove worker and reschedule its tasks
						err := sched.RemoveWorker(workerID)
						if err != nil {
							logger.Error(fmt.Sprintf("Failed to remove worker: %v", err))
						} else {
							// Delete from map to prevent repeated removal attempts
							delete(workerLastSeen, workerID)
							logger.Info(fmt.Sprintf("Worker %s removed and its tasks rescheduled", workerID))
						}
					}
				}
			}
		}
	}()

	// Start task processing in scheduler
	sched.StartTaskProcessing()

	// Set up API
	apiHandler := api.Handler{Scheduler: sched}
	router := api.RegisterRoutes(apiHandler)

	// Start HTTP server
	logger.Debug("Starting server on :8080")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error: " + err.Error())
		}
	}()

	// Create a channel to handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-stop
	logger.Debug("Shutting down server...")
}
