package messaging

import (
	"encoding/json"
	"fmt"

	Rabbit2 "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models/Rabbit"
)

func (r *Rabbit) ListenTaskResults(c chan Rabbit2.TaskReplyWrapper) {
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
		message := Rabbit2.TaskReplyWrapper{
			TaskReply: Rabbit2.TaskReply{},
			Err:       err,
		}
		c <- message
		return
	}
	for {
		for d := range msgs {
			var m Rabbit2.TaskReply
			err = json.Unmarshal(d.Body, &m)
			if err != nil {
				fmt.Printf("Failed to unmarshal", "error", err)
			}
			message := Rabbit2.TaskReplyWrapper{
				TaskReply: m,
				Err:       err,
			}
			c <- message
			fmt.Println(m.Results, m.Err, m.WorkerId, m.JobId)
		}
	}
}
