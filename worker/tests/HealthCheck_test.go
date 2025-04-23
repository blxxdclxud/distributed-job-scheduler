package tests

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/globals"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models/Rabbit"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func Test_Health(t *testing.T) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	err = ch.ExchangeDeclare(
		globals.WorkerStatusExchangeName, // name
		"topic",                          // type
		true,                             // durable
		false,                            // auto-deleted
		false,                            // internal
		false,                            // no-wait
		nil,                              // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,                           // queue name
		"heartbeat.*",                    // routing key
		globals.WorkerStatusExchangeName, // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	err = ch.ExchangeDeclare(
		globals.ResultExchange, // name
		"topic",                // type
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q2, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q2.Name,                // queue name
		"result.*",             // routing key
		globals.ResultExchange, // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")
	go func() {
		for d := range msgs {
			var m Rabbit.HealthReport
			e := json.Unmarshal(d.Body, &m)
			if e != nil {
				fmt.Println(err)
			}
			fmt.Println(m.TimeStamp, m.WorkerId)
		}
	}()
	failOnError(err, "Failed to declare a queue")

	msgs2, err := ch.Consume(
		q2.Name, // queue
		"",      // consumer
		true,    // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	failOnError(err, "Failed to register a consumer")
	go func() {
		for d := range msgs2 {
			var m Rabbit.TaskReply
			e := json.Unmarshal(d.Body, &m)
			if e != nil {
				fmt.Println(err)
			}
			fmt.Println(m.Results, m.Err, m.WorkerId, m.JobId)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.ExchangeDeclare(
		globals.RegisterExchange, // name
		"direct",                 // type
		true,                     // durable
		false,                    // auto-deleted
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q5, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q5.Name,                  // queue name
		"register",               // routing key
		globals.RegisterExchange, // exchange
		false,
		nil,
	)
	msgs5, err := ch.Consume(
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
	var Id string
	done := make(chan bool)
	go func() {
		for d := range msgs5 {
			m := json.Unmarshal(d.Body, &Id)
			if m != nil {
				fmt.Println(m)
			}
			done <- true
			return
		}
	}()
	select {
	case <-done:
		fmt.Println("get id")
		time.Sleep(500 * time.Millisecond)
	case <-time.After(10 * time.Second):
		fmt.Println("Failed to get Id")
		return
	}
	fmt.Println(Id)
	err = ch.ExchangeDeclare(
		globals.LuaProgramsExchange, // name
		"direct",                    // type
		true,                        // durable
		false,                       // auto-deleted
		false,                       // internal
		false,                       // no-wait
		nil,                         // arguments
	)
	if err != nil {
		fmt.Println(err)
	}
	luaScript := `local start = os.time(); while os.time() - start < 2 do end; local a, b = 10, 20; return a * b`
	LuaTask := Rabbit.LuaTask{
		LuaCode: luaScript,
		JobId:   "1",
	}
	body, err := json.Marshal(LuaTask)
	if err != nil {
		log.Fatal(err)
	}
	err = ch.PublishWithContext(ctx,
		globals.LuaProgramsExchange, // exchange
		Id,                          // routing key
		false,                       // mandatory
		false,                       // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	failOnError(err, "Failed to publish a message")

	log.Printf(" [x] Sent code ")

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigChan:
		fmt.Println("Received interrupt signal, shutting down...")
		conn.Close()
	case <-time.After(60 * time.Second):
		fmt.Println("Test timeout reached")
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
