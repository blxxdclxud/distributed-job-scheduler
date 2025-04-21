package main

import (
	"DistributedJobScheduling/pkg/logger"
	"log"
)

func main() {
	err := logger.Init("development")
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

}
