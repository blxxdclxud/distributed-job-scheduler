package InitRabbit

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/globals"
)

func InitTaskResult(ch *amqp.Channel) (amqp.Queue, error) {
	q2, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return q2, err
	}
	err = ch.QueueBind(
		q2.Name,                // queue name
		"result.*",             // routing key
		globals.ResultExchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return q2, err
	}
	return q2, nil
}
