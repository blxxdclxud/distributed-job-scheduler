package scheduler

import (
	"testing"

	sharedModels "gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"
)

// TestPriorityOrdering verifies that High-priority jobs always dequeue before Mid and Low,
// regardless of insertion order (CORR-01).
func TestPriorityOrdering(t *testing.T) {
	s := newTestScheduler()

	// Insert low-priority first, then mid, then high — worst case for a naive FIFO queue.
	for i := 0; i < 5; i++ {
		s.EnqueueJob(sharedModels.LowPriority, "low")
	}
	for i := 0; i < 5; i++ {
		s.EnqueueJob(sharedModels.MidPriority, "mid")
	}
	for i := 0; i < 5; i++ {
		s.EnqueueJob(sharedModels.HighPriority, "high")
	}

	// Dequeue all 15 and verify non-decreasing numeric priority (High=1, Mid=2, Low=3).
	lastPriority := 0
	for i := 0; i < 15; i++ {
		job, ok := s.Jobs.Get()
		if !ok {
			t.Fatalf("expected job at iteration %d but queue was empty", i)
		}
		if int(job.Priority) < lastPriority {
			t.Errorf("priority ordering broken at position %d: got %d after %d",
				i, job.Priority, lastPriority)
		}
		lastPriority = int(job.Priority)
	}
}

// TestRoundRobinDistribution verifies that worker selection spreads load evenly,
// with a max imbalance of 1 job per worker (CORR-02).
func TestRoundRobinDistribution(t *testing.T) {
	const M = 5
	const N = 25

	s := newTestScheduler()
	addWorkers(s, M)

	counts := make(map[string]int, M)

	for i := 0; i < N; i++ {
		w := s.RoundRobin()
		if w == nil {
			t.Fatalf("RoundRobin returned nil at iteration %d — worker queue exhausted", i)
		}
		counts[w.ID]++
		// Re-enqueue so the queue never drains.
		s.AvailableWorkers.Add(*w)
	}

	// Verify all M workers were used.
	if len(counts) != M {
		t.Errorf("expected %d distinct workers, got %d: %v", M, len(counts), counts)
	}

	// Verify imbalance <= 1.
	minCount, maxCount := N, 0
	for _, c := range counts {
		if c < minCount {
			minCount = c
		}
		if c > maxCount {
			maxCount = c
		}
	}
	if maxCount-minCount > 1 {
		t.Errorf("load imbalance too high: max=%d min=%d (allowed <=1), distribution: %v",
			maxCount, minCount, counts)
	}
}
