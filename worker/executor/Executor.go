package executor

import (
	"context"
	"fmt"
	"github.com/Shopify/go-lua"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/globals"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models/Rabbit"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/worker/HealthReporter"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/worker/messaging"
	"log/slog"
)

type Executor struct {
	conn              *amqp.Connection
	log               *slog.Logger
	RabbitMQPublisher HealthReporter.Publisher
}

func NewExecutor(log *slog.Logger, RabbitMqConn *amqp.Connection) *Executor {
	p, err := messaging.NewRabbitMQPublisher(RabbitMqConn, globals.ResultExchange, "topic")
	if err != nil {
		log.Error("NewExecutor", "err", err)
		panic(err)
	}
	return &Executor{conn: RabbitMqConn, log: log, RabbitMQPublisher: p}
}

func (e *Executor) ListenTasks(workerId string) {
	ch, err := e.conn.Channel()
	if err != nil {
		e.log.Error("Failed to open a channel", "error", err)
	}
	err = ch.ExchangeDeclare(
		globals.LuaProgramsExchange, // name
		"direct",                    // type
		true,                        // durable
		false,                       // auto-deleted
		false,                       // internal
		false,                       // no-wait
		nil,                         // arguments
	)
	if err != nil {
		e.log.Error("Failed to declare an exchange", "error", err)
	}
	q, err := ch.QueueDeclare(
		"",    // name
		true,  // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		e.log.Error("Failed to declare a queue", "error", err)
	}
	err = ch.QueueBind(
		q.Name,                      // queue name
		workerId,                    // routing key
		globals.LuaProgramsExchange, // exchange
		false,
		nil)
	if err != nil {
		e.log.Error("Failed to bind a queue", "error", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	if err != nil {
		e.log.Error("Failed to register a consumer", "error", err)
	}
	go func() {
		e.log.Info("Listening for tasks...")
		for d := range msgs {
			e.log.Info("Received a message from server")
			err = d.Ack(false)
			if err != nil {
				e.log.Error("Failed to ack message", "error", err)
			}
			res, err := e.Task(d.Body, workerId)
			if err != nil {
				e.log.Error("Failed to process task", globals.ResultExchange, err)
			}
			message := Rabbit.TaskReply{
				Results:  res,
				WorkerId: workerId,
				Err:      err,
			}
			routing_key := "result." + workerId
			err = e.RabbitMQPublisher.PublishJSON(context.Background(), routing_key, message)
			if err != nil {
				e.log.Error("Failed sending results")
			}
		}
	}()

}
func (e *Executor) Task(body []byte, workerId string) (interface{}, error) {
	l := lua.NewState()
	lua.OpenLibraries(l)

	if err := lua.DoString(l, string(body)); err != nil {
		return nil, fmt.Errorf("lua execution error: %w", err)
	}
	if l.Top() == 0 {
		return struct{}{}, nil
	}
	value := l.ToValue(1)
	return value, nil
}
