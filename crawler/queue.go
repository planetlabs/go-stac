package crawler

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

const (
	resourceTask    = "resource"
	collectionsTask = "collections"
	childrenTask    = "children"
	featuresTask    = "features"
)

type Task struct {
	Url  string
	Type string
}

type Handler func(task *Task) error

type Queue interface {
	Add(task *Task) error
	Handle(handler Handler)
	Wait() error
}

// NewMemoryQueue is used if a custom queue is not provided for a crawl.
func NewMemoryQueue(ctx context.Context, limit int) Queue {
	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(limit)
	return &memoryQueue{
		ctx:     ctx,
		group:   group,
		mutex:   &sync.Mutex{},
		buffer:  []*Task{},
		handler: nil,
	}
}

type memoryQueue struct {
	ctx     context.Context
	group   *errgroup.Group
	mutex   *sync.Mutex
	buffer  []*Task
	handler Handler
}

func (q *memoryQueue) Add(task *Task) error {
	q.mutex.Lock()
	q.buffer = append(q.buffer, task)
	q.mutex.Unlock()
	q.process()
	return nil
}

func (q *memoryQueue) Handle(handler Handler) {
	q.handler = handler
	q.process()
}

func (q *memoryQueue) Wait() error {
	return q.group.Wait()
}

func (q *memoryQueue) process() {
	if q.handler == nil || len(q.buffer) == 0 {
		return
	}

	q.mutex.Lock()
	count := 0
	for _, task := range q.buffer {
		task := task
		added := q.group.TryGo(func() error {
			defer q.process()
			return q.handler(task)
		})
		if !added {
			break
		}
		count += 1
	}

	if count > 0 {
		q.buffer = q.buffer[count:]
	}
	q.mutex.Unlock()
}

var _ Queue = (*memoryQueue)(nil)
