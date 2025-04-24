package messaging

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/globals"
	Rabbit2 "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models/Rabbit"
)

func (r *Rabbit) SendTaskToWorker(ctx context.Context, luaCode string, workerId string, JobId string) error {
	LuaTask := Rabbit2.LuaTask{
		LuaCode: luaCode,
		JobId:   JobId,
	}
	body, err := json.Marshal(LuaTask)
	if err != nil {
		return err
	}
	err = r.channel.PublishWithContext(ctx,
		globals.LuaProgramsExchange, // exchange
		workerId,                    // routing key
		false,                       // mandatory
		false,                       // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return err
	}
	return nil
}
