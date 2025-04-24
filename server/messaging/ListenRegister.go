package messaging

import (
	"encoding/json"
	"fmt"
	Rabbit2 "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models/Rabbit"
	"go.uber.org/zap"
)

func (r *Rabbit) ListenRegister(c chan Rabbit2.RegistrationWrapper) {
	msgs5, err := r.channel.Consume(
		r.RegisteredQ.Name, // queue
		"",                 // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		fmt.Println(err)
		message := Rabbit2.RegistrationWrapper{WorkerId: "", Err: err}
		c <- message
		return
	}
	for {
		for d := range msgs5 {
			var workerId string
			err = json.Unmarshal(d.Body, &workerId)
			if err != nil {
				fmt.Printf("Failed to unmarshal", zap.Error(err))
			}
			message := Rabbit2.RegistrationWrapper{WorkerId: workerId, Err: err}
			c <- message
		}
	}
}
