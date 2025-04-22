package main

import (
	"DNO/pkg/logger"
	"log"
)

func main() {
	err := logger.Init("development")
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

}
