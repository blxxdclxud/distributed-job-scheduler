package messaging

import (
	"encoding/json"
	"fmt"
	Rabbit2 "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models/Rabbit"
	"go.uber.org/zap"
)

func (r *Rabbit) ListenHeartBeat(c chan Rabbit2.HealthReportWrapper) {
	msgs, err := r.channel.Consume(
		r.HeartBearQ.Name, // queue
		"",                // consumer
		true,              // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		message := Rabbit2.HealthReportWrapper{
			HealthReport: Rabbit2.HealthReport{},
			Err:          err,
		}
		c <- message
		return
	}
	for {
		for d := range msgs {
			var m Rabbit2.HealthReport
			err = json.Unmarshal(d.Body, &m)
			if err != nil {
				fmt.Printf("Failed to unmarshal", zap.Error(err))
				return
			}
			message := Rabbit2.HealthReportWrapper{
				HealthReport: m,
				Err:          err,
			}
			c <- message
			fmt.Println(m.TimeStamp, m.WorkerId)
		}
	}
}
