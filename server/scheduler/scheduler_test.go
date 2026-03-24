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

// BenchmarkEnqueueJob measures job submission throughput (BENCH-01).
func BenchmarkEnqueueJob(b *testing.B) {
	s := newTestScheduler()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.EnqueueJob(sharedModels.HighPriority, "return 1")
	}
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "jobs/sec")
}

// BenchmarkPrioritySelection measures dequeue latency from mixed-priority queues (BENCH-03).
func BenchmarkPrioritySelection(b *testing.B) {
	s := newTestScheduler()
	priorities := []sharedModels.JobPriority{
		sharedModels.HighPriority,
		sharedModels.MidPriority,
		sharedModels.LowPriority,
	}
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		s.EnqueueJob(priorities[i%3], "return 1")
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Jobs.Get()
	}
}

// BenchmarkAssignTask measures worker-selection + job-dequeue latency (BENCH-02).
func BenchmarkAssignTask(b *testing.B) {
	s := newTestScheduler()
	b.StopTimer()
	// Pre-fill jobs and workers
	for i := 0; i < b.N; i++ {
		s.EnqueueJob(sharedModels.HighPriority, "return 1")
	}
	// Use a single worker, re-enqueue after each AssignTask call
	w := sharedModels.Worker{ID: "bench-worker"}
	s.RegisterWorker(w)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.AssignTask()
		// Re-enqueue worker so next iteration has a worker available
		s.AvailableWorkers.Add(w)
	}
}
