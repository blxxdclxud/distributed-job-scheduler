package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
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

			// Parse job ID from string to int
			jobID := 0
			if result.TaskReply.JobId != "" {
				_, err := fmt.Sscanf(result.TaskReply.JobId, "%d", &jobID)
				if err != nil {
					logger.Error("Invalid job ID in result: " + result.TaskReply.JobId)
					continue
				}
			}

			logger.Debug("Received result for job " + result.TaskReply.JobId)

			// Update job status based on result
			job, err := sched.GetJob(jobID)
			if err != nil {
				logger.Error("Job not found: " + err.Error())
				continue
			}

			// Update job status and result
			if result.TaskReply.Err != nil {
				job.Status = models.StatusFailed
				job.Result = result.TaskReply.Err.Error()
			} else {
				job.Status = models.StatusCompleted
				// Convert result to string
				resultStr := ""
				if result.TaskReply.Results != nil {
					resultStr = fmt.Sprintf("%v", result.TaskReply.Results)
				}
				job.Result = resultStr
			}

			// Update job in scheduler
			sched.UpdateJob(jobID, job)
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
						logger.Warn("Worker may be down: " + workerID)
						// Here you could implement logic to handle dead workers
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
