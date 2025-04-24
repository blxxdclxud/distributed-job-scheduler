package tests

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server/messaging"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models/Rabbit"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func Test_Rabbit(t *testing.T) {
	fmt.Println("sadadsa")
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Println(err)
		return
	}
	r, err := messaging.NewRabbit(conn)
	if err != nil {
		fmt.Println(err)
	}
	ch1 := make(chan Rabbit.HealthReportWrapper, 20)
	ch2 := make(chan Rabbit.RegistrationWrapper, 20)
	ch3 := make(chan Rabbit.TaskReplyWrapper, 20)
	go r.ListenHeartBeat(ch1)
	go r.ListenRegister(ch2)
	go r.ListenTaskResults(ch3)
	time.Sleep(10 * time.Second)
	var workerId string
	var n Rabbit.RegistrationWrapper
	select {
	case n = <-ch2:
		workerId = n.WorkerId
		fmt.Println("Получен workerId:", workerId)
	case <-time.After(10 * time.Second):
		fmt.Println("Failed to get Id")
		return
	}
	luaScript := `local start = os.time(); while os.time() - start < 2 do end; local a, b = 10, 20; return a * b`
	err = r.SendTaskToWorker(context.Background(), luaScript, workerId, "1")
	if err != nil {
		fmt.Println(err)
	}
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
