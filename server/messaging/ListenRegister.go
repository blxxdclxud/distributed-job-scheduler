package messaging

import (
	"encoding/json"
	"fmt"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/globals"
	"go.uber.org/zap"
)

func (r *Rabbit) ListenRegister(c chan string) {
	q5, err := r.channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		r.log.Error("Failed to declare a queue", zap.Error(err))
	}
	err = r.channel.QueueBind(
		q5.Name,                  // queue name
		"register",               // routing key
		globals.RegisterExchange, // exchange
		false,
		nil,
	)
	if err != nil {
		r.log.Error("Failed to bind a queue", zap.Error(err))
	}
	msgs5, err := r.channel.Consume(
		q5.Name, // queue
		"",      // consumer
		true,    // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		fmt.Println(err)
	}
	for {
		for d := range msgs5 {
			var workerId string
			err = json.Unmarshal(d.Body, &workerId)
			if err != nil {
				r.log.Error("Failed to unmarshal", zap.Error(err))
			}
			c <- workerId
		}
	}
}
