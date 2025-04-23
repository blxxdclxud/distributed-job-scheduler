package messaging

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/globals"
)

type Rabbit struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbit(conn *amqp.Connection) (*Rabbit, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	//declare exchange to send task
	err = ch.ExchangeDeclare(
		globals.LuaProgramsExchange, // name
		"direct",                    // type
		true,                        // durable
		false,                       // auto-deleted
		false,                       // internal
		false,                       // no-wait
		nil,                         // arguments
	)
	// declare exchange to get worker status
	err = ch.ExchangeDeclare(
		globals.WorkerStatusExchangeName, // name
		"topic",                          // type
		true,                             // durable
		false,                            // auto-deleted
		false,                            // internal
		false,                            // no-wait
		nil,                              // arguments
	)
	// declare exchage for results
	err = ch.ExchangeDeclare(
		globals.ResultExchange, // name
		"topic",                // type
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	//declare exchange for register
	err = ch.ExchangeDeclare(
		globals.RegisterExchange, // name
		"direct",                 // type
		true,                     // durable
		false,                    // auto-deleted
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	)
	return &Rabbit{conn, ch}, nil
}

func (r *Rabbit) SendTaskToWorker(ctx context.Context, luaCode string, workerId string) error {
	err := r.channel.PublishWithContext(ctx,
		globals.LuaProgramsExchange, // exchange
		workerId,                    // routing key
		false,                       // mandatory
		false,                       // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(luaCode),
		})
	if err != nil {
		return err
	}
	return nil
}

func (r *Rabbit) ListenHeartBeat(ctx context.Context) error {
	q, err := r.channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}
	err = r.channel.QueueBind(
		q.Name,                           // queue name
		"heartbeat.*",                    // routing key
		globals.WorkerStatusExchangeName, // exchange
		false,
		nil,
	)
	go func() {

	}()
}
