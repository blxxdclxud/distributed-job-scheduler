package scheduler

import (
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"
	"github.com/golang-collections/collections/queue"
)

// WorkerQueue stores all workers and acts as queue
type WorkerQueue struct {
	q *queue.Queue
}

// NewWorkerQueue initialize new WorkerQueue object
func NewWorkerQueue() *WorkerQueue {
	return &WorkerQueue{q: queue.New()}
}

// Add appends given task to the corresponding queue according to task's priority
func (w *WorkerQueue) Add(task models.Worker) {
	w.q.Enqueue(task)
}

// Get returns the next available worker that will perform the task.
// Returns nil if no workers exist.
func (w *WorkerQueue) Get() (models.Worker, bool) {

	if worker, ok := w.q.Dequeue().(models.Worker); ok {
		return worker, true
	}
	return models.Worker{}, false

}
