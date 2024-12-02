package api

import (
	"collector-api/internal/config"
	"sync"
)

type SnapshotTask struct {
	S3Location  string
	CollectedAt int64
	SystemInfo  SystemInfo
	IsCompact   bool
}

type Queue struct {
	tasks    []SnapshotTask
	mu       sync.Mutex
	isLocked bool
}

var (
	instance *Queue
	once     sync.Once
)

func GetQueueInstance() *Queue {
	once.Do(func() {
		instance = &Queue{
			tasks: make([]SnapshotTask, 0),
		}
	})
	return instance
}

func (q *Queue) Lock() {
	q.mu.Lock()
	q.isLocked = true
	q.mu.Unlock()
}

func (q *Queue) Unlock() {
	q.mu.Lock()
	q.isLocked = false
	q.mu.Unlock()
}

func (q *Queue) IsLocked() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.isLocked
}

func (q *Queue) Enqueue(task SnapshotTask) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tasks = append(q.tasks, task)
}

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
