package scheduler

import (
	"fmt"
	"testing"

	sharedModels "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"
)

// newTestScheduler returns a Scheduler with no RabbitMQ client (safe for unit tests).
func newTestScheduler() *Scheduler {
	return NewScheduler()
}

// addWorkers registers n workers with sequential IDs into s.
func addWorkers(s *Scheduler, n int) {
	for i := 1; i <= n; i++ {
		s.RegisterWorker(sharedModels.Worker{ID: fmt.Sprintf("worker-%d", i)})
	}
}

// BenchmarkEnqueueJob placeholder — full implementation in plan 02.
func BenchmarkEnqueueJob(b *testing.B) {
	b.Skip("implemented in plan 02")
}
