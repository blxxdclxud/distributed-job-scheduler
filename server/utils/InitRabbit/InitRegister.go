package InitRabbit

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/globals"
)

func InitRegister(ch *amqp.Channel) (amqp.Queue, error) {
	q5, err := ch.QueueDeclare(
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
		q5.Name,                  // queue name
		"register",               // routing key
		globals.RegisterExchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return amqp.Queue{}, err
	}
	return q5, nil
}
