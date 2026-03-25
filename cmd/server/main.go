package main

import (
	"flag"
	"fmt"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/server"
	"log"

	logger "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/pkg/logger"
)

func main() {
	var rmq_host string

	flag.StringVar(&rmq_host, "rmq", "amqp://guest:guest@localhost:5672/", "rabbitmq host address")

	flag.Parse()
	fmt.Println("rmq host:", rmq_host)

	err := logger.Init("development")
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	server.RunServer()
}
