package messaging

import (
	"encoding/json"
	"fmt"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/globals"
	Rabbit2 "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models/Rabbit"
	"go.uber.org/zap"
)

func (r *Rabbit) ListenTaskResults(c chan Rabbit2.TaskReply) {
	q, err := r.channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		r.log.Error("Failed to declare a queue", zap.Error(err))
		return
	}
	err = r.channel.QueueBind(
		q.Name,                           // queue name
		"heartbeat.*",                    // routing key
		globals.WorkerStatusExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		r.log.Error("Failed to bind a queue", zap.Error(err))
		return
	}
	msgs, err := r.channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	for {
		for d := range msgs {
			var m Rabbit2.TaskReply
			err = json.Unmarshal(d.Body, &m)
			if err != nil {
				r.log.Error("Failed to unmarshal", zap.Error(err))
				return
			}
			c <- m
			fmt.Println(m.Results, m.Err, m.WorkerId, m.JobId)
		}
	}
}
