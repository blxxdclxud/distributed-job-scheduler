package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Shopify/go-lua"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/globals"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models/Rabbit"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/worker/HealthReporter"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/worker/messaging"
)

type Executor struct {
	conn               *amqp.Connection
	log                *slog.Logger
	RabbitMQPublisher  HealthReporter.Publisher
	RabbitAckPublisher HealthReporter.Publisher
}

func NewExecutor(log *slog.Logger, RabbitMqConn *amqp.Connection) *Executor {
	p, err := messaging.NewRabbitMQPublisher(RabbitMqConn, globals.ResultExchange, "topic")
	ack, err := messaging.NewRabbitMQPublisher(RabbitMqConn, globals.WorkerStatusExchangeName, "topic")
	if err != nil {
		log.Error("NewExecutor", "err", err)
		panic(err)
	}
	return &Executor{conn: RabbitMqConn, log: log, RabbitMQPublisher: p, RabbitAckPublisher: ack}
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
			AckMessage := Rabbit.HealthReport{WorkerId: workerId, TimeStamp: time.Now().Unix()}
			routing_key_ack := "heartbeat." + workerId
			err = e.RabbitAckPublisher.PublishJSON(context.Background(), routing_key_ack, AckMessage)
			if err != nil {
				e.log.Error("Failed to publish ack message", "error", err)
			}
			var task Rabbit.LuaTask
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				e.log.Error("Failed to unmarshal task", "error", err)
			}
			res, err := e.Task(task.LuaCode, workerId)
			if err != nil {
				e.log.Error("Failed to process task", globals.ResultExchange, err)
			}
			var resultStr string
			if res != nil {
				resultStr = fmt.Sprintf("%v", res)
			} else {
				resultStr = ""
			}
			var errStr string
			if err != nil {
				errStr = err.Error()
			} else {
				errStr = ""
			}
			message := Rabbit.TaskReply{
				Results:  resultStr,
				WorkerId: workerId,
				Err:      errStr,
				JobId:    task.JobId,
			}
			routing_key := "result." + workerId
			err = e.RabbitMQPublisher.PublishJSON(context.Background(), routing_key, message)
			if err != nil {
				e.log.Error("Failed sending results", "error", err)
			}
		}
	}()

}
func (e *Executor) Task(body string, workerId string) (interface{}, error) {
	l := lua.NewState()
	lua.OpenLibraries(l)

	if err := lua.DoString(l, body); err != nil {
		return nil, fmt.Errorf("lua execution error: %w", err)
	}
	if l.Top() == 0 {
		return struct{}{}, nil
	}
	value := l.ToValue(1)
	return value, nil
}
