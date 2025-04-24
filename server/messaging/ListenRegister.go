package messaging

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
)

func (r *Rabbit) ListenRegister(c chan string) {
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
	}
	for {
		for d := range msgs5 {
			var workerId string
			err = json.Unmarshal(d.Body, &workerId)
			if err != nil {
				fmt.Printf("Failed to unmarshal", zap.Error(err))
			}
			c <- workerId
		}
	}
}
