package messaging

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/globals"
	"go.uber.org/zap"
)

type Rabbit struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	log     *zap.Logger
}

func NewRabbit(conn *amqp.Connection, log *zap.Logger) (*Rabbit, error) {
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
	return &Rabbit{conn, ch, log}, nil
}
