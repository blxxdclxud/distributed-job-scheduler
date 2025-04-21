package server

import (
	"DistributedJobScheduling/pkg/logger"
	"fmt"
)

func main() {
	err := logger.Init("development")
	if err != nil {
		fmt.Errorf("failed to initialize logger: %w", err)
	}

}
