package HealthReporter

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/globals"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models/Rabbit"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/worker/messaging"
	"log/slog"
	"time"
)

type HealthReporter struct {
	log               *slog.Logger
	conn              *amqp.Connection
	RabbitMQPublisher Publisher
}

type Publisher interface {
	PublishJSON(ctx context.Context, routingKey string, message interface{}) error
	Close() error
}

func NewHealthReporter(log *slog.Logger, RabbitMqConn *amqp.Connection) *HealthReporter {
	p, err := messaging.NewRabbitMQPublisher(RabbitMqConn, globals.WorkerStatusExchangeName, "topic")
	if err != nil {
		log.Error("NewHealthReporter", "err", err)
		panic(err)
	}

	return &HealthReporter{log: log, conn: RabbitMqConn, RabbitMQPublisher: p}
}

func (h *HealthReporter) SendHealthChecks(workerId string) {
	h.log.Info("HealthReporter SendHealthChecks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			message := Rabbit.HealthReport{WorkerId: workerId, TimeStamp: time.Now().Unix()}
			routing_key := "heartbeat." + workerId
			err := h.RabbitMQPublisher.PublishJSON(ctx, routing_key, message)
			if err != nil {
				h.log.Error("HealthReporter SendHealthChecks", "err", err)
			}
			h.log.Info("HealthReport sent")
		}
	}
}
