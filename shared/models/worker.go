package models

// add here worker, heartbeat structs and other related things
import (
	"log/slog"

	"github.com/rabbitmq/amqp091-go"
	HealthReporter2 "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/worker/HealthReporter"
	Executor2 "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/worker/executor"
)

type Worker struct {
	workerId       string
	RabbitMqConn   *amqp091.Connection
	Logger         *slog.Logger
	Executor       Executor
	HealthReporter HealthReporter
}

type Executor interface {
	ListenTasks(workerId string)
}

type HealthReporter interface {
	SendHealthChecks(workerId string)
}

func NewWorker(connection *amqp091.Connection, Logger *slog.Logger, workerId string) *Worker {
	executor := Executor2.NewExecutor(Logger, connection)
	Health := HealthReporter2.NewHealthReporter(Logger, connection)
	return &Worker{
		RabbitMqConn:   connection,
		Logger:         Logger,
		workerId:       workerId,
		HealthReporter: Health,
		Executor:       executor,
	}
}

func (w *Worker) Start() {
	w.Logger.Info("Starting worker", "worker_id", w.workerId)
	go w.Executor.ListenTasks(w.workerId)
	go w.HealthReporter.SendHealthChecks(w.workerId)
}

// SetWorkerId sets the worker ID
func (w *Worker) SetWorkerId(id string) {
	w.workerId = id
}

// GetID returns the worker ID
func (w *Worker) GetID() string {
	return w.workerId
}
