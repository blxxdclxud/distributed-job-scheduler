package InitRabbit

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/globals"
)

func InitHeartBeat(ch *amqp.Channel) (amqp.Queue, error) {
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return amqp.Queue{}, err
	}
	err = ch.QueueBind(
		q.Name,                           // queue name
		"heartbeat.*",                    // routing key
		globals.WorkerStatusExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return amqp.Queue{}, err
	}
	return q, nil
}
