package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/globals"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/worker/messaging"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	var rmq_host string

	flag.StringVar(&rmq_host, "rmq", "amqp://guest:guest@localhost:5672/", "rabbitmq host address")

	flag.Parse()
	fmt.Println("rmq host:", rmq_host)

	conn, err := amqp.Dial(rmq_host)
	failOnError(err, "Failed to connect to RabbitMQ")
	log := setupLogger(envLocal)
	id := uuid.New().String()
	if err != nil {
		log.Error("Failed to publish register message")
	}
	if err != nil {
		log.Error("Failed to create RabbitMQ publisher")
	}
	worker := models.NewWorker(conn, log, id)
	worker.Start()
	time.Sleep(1 * time.Second)

	register, err := messaging.NewRabbitMQPublisher(conn, globals.RegisterExchange, "direct")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = register.PublishJSON(ctx, "register", id)
	if err != nil {
		log.Error("Failed to publish register message")
	}
	log.Info("Successfully published register message")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	defer conn.Close()
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
