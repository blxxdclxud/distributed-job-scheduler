package messaging

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/utils/InitRabbit"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/globals"
)

type Rabbit struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	HeartBearQ  amqp.Queue
	RegisteredQ amqp.Queue
	TaskREsultQ amqp.Queue
}

func NewRabbit(conn *amqp.Connection) (*Rabbit, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	//declare exchange to send task
	err = ch.ExchangeDeclare(
		globals.LuaProgramsExchange, // name
		"direct",                    // type
		true,                        // durable
		false,                       // auto-deleted
		false,                       // internal
		false,                       // no-wait
		nil,                         // arguments
	)
	// declare exchange to get worker status
	err = ch.ExchangeDeclare(
		globals.WorkerStatusExchangeName, // name
		"topic",                          // type
		true,                             // durable
		false,                            // auto-deleted
		false,                            // internal
		false,                            // no-wait
		nil,                              // arguments
	)
	// declare exchage for results
	err = ch.ExchangeDeclare(
		globals.ResultExchange, // name
		"topic",                // type
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	//declare exchange for register
	err = ch.ExchangeDeclare(
		globals.RegisterExchange, // name
		"direct",                 // type
		true,                     // durable
		false,                    // auto-deleted
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	)
	HeartBeatQ, err := InitRabbit.InitHeartBeat(ch)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	TaskResultQ, err := InitRabbit.InitTaskResult(ch)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	RegisterQ, err := InitRabbit.InitRegister(ch)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &Rabbit{
		conn:        conn,
		channel:     ch,
		HeartBearQ:  HeartBeatQ,
		RegisteredQ: RegisterQ,
		TaskREsultQ: TaskResultQ,
	}, nil
}
