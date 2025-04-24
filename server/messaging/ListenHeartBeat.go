package messaging

import (
	"encoding/json"
	"fmt"
	Rabbit2 "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models/Rabbit"
	"go.uber.org/zap"
)

func (r *Rabbit) ListenHeartBeat(c chan Rabbit2.HealthReport) {
	msgs, err := r.channel.Consume(
		r.HeartBearQ.Name, // queue
		"",                // consumer
		true,              // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	for {
		for d := range msgs {
			var m Rabbit2.HealthReport
			err = json.Unmarshal(d.Body, &m)
			if err != nil {
				fmt.Printf("Failed to unmarshal", zap.Error(err))
				return
			}
			c <- m
			fmt.Println(m.TimeStamp, m.WorkerId)
		}
	}
}
