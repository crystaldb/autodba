package api

import (
	"collector-api/internal/config"
	"sync"
)

// SnapshotTask represents a task to be processed, containing information about
// the snapshot's S3 location, collection time, system information, and whether
// the snapshot is compact.
type SnapshotTask struct {
	S3Location  string     // The S3 location of the snapshot
	CollectedAt int64      // The timestamp when the snapshot was collected
	SystemInfo  SystemInfo // Information about the system from which the snapshot was collected
	IsCompact   bool       // Indicates if the snapshot is compact
}

// Queue is a thread-safe queue for managing SnapshotTasks. It uses a mutex
// to ensure safe concurrent access.
type Queue struct {
	tasks    []SnapshotTask // The list of tasks in the queue
	mu       sync.Mutex     // Mutex to protect access to the queue
	isLocked bool           // Indicates if the queue is locked
}

var (
	instance *Queue    // Singleton instance of the Queue
	once     sync.Once // Ensures the Queue is only initialized once
)

// GetQueueInstance returns the singleton instance of the Queue, initializing
// it if necessary.
func GetQueueInstance() *Queue {
	once.Do(func() {
		instance = &Queue{
			tasks: make([]SnapshotTask, 0),
		}
	})
	return instance
}

// Lock locks the queue, preventing other operations from modifying it.
func (q *Queue) Lock() {
	q.mu.Lock()
	q.isLocked = true
	q.mu.Unlock()
}

// Unlock unlocks the queue, allowing other operations to modify it.
func (q *Queue) Unlock() {
	q.mu.Lock()
	q.isLocked = false
	q.mu.Unlock()
}

// IsLocked returns true if the queue is currently locked.
func (q *Queue) IsLocked() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.isLocked
}

// Enqueue adds a new SnapshotTask to the queue.
func (q *Queue) Enqueue(task SnapshotTask) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tasks = append(q.tasks, task)
}

// ProcessQueue processes all tasks in the queue. It handles each task based
// on whether it is compact or full, using the provided configuration.
func (q *Queue) ProcessQueue(cfg *config.Config) error {
	q.mu.Lock()
	tasks := q.tasks
	q.tasks = make([]SnapshotTask, 0)
	q.mu.Unlock()

	for _, task := range tasks {
		if task.IsCompact {
			if err := HandleCompactSnapshot(cfg, task.S3Location, task.CollectedAt, task.SystemInfo); err != nil {
				return err
			}
		} else {
			if err := HandleFullSnapshot(cfg, task.S3Location, task.CollectedAt, task.SystemInfo); err != nil {
				return err
			}
		}
	}
	return nil
}
