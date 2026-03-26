package scheduler

import (
	"github.com/golang-collections/collections/queue"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/shared/models"
)

// WorkerQueue stores all workers and acts as queue
type WorkerQueue struct {
	q *queue.Queue
}

// NewWorkerQueue initialize new WorkerQueue object
func NewWorkerQueue() *WorkerQueue {
	return &WorkerQueue{q: queue.New()}
}

// Add appends given worker to the queue
func (w *WorkerQueue) Add(worker models.Worker) {
	w.q.Enqueue(worker)
}

// Get dequeues and returns the next available worker.
// Returns the worker and true if one exists, otherwise an empty worker and false.
func (w *WorkerQueue) Get() (models.Worker, bool) {
	if w.q.Len() == 0 {
		return models.Worker{}, false
	}

	worker, ok := w.q.Dequeue().(models.Worker)
	if !ok {
		return models.Worker{}, false
	}

	return worker, true
}

// Size returns the number of workers in the queue
func (w *WorkerQueue) Size() int {
	return w.q.Len()
}
