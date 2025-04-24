package messaging

import (
	"encoding/json"
	"fmt"
	Rabbit2 "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models/Rabbit"
	"go.uber.org/zap"
)

func (r *Rabbit) ListenTaskResults(c chan Rabbit2.TaskReply) {
	msgs, err := r.channel.Consume(
		r.TaskREsultQ.Name, // queue
		"",                 // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		fmt.Println("Failed to register consumer")
		return
	}
	for {
		for d := range msgs {
			var m Rabbit2.TaskReply
			err = json.Unmarshal(d.Body, &m)
			if err != nil {
				fmt.Printf("Failed to unmarshal", zap.Error(err))
				return
			}
			c <- m
			fmt.Println(m.Results, m.Err, m.WorkerId, m.JobId)
		}
	}
}
