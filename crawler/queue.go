package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/planetlabs/go-stac/internal/normurl"
	"golang.org/x/sync/errgroup"
)

type taskType string

const (
	resourceTask    = taskType("resource")
	collectionsTask = taskType("collections")
	childrenTask    = taskType("children")
	featuresTask    = taskType("features")
)

var validTaskTypes = map[taskType]bool{
	resourceTask:    true,
	collectionsTask: true,
	childrenTask:    true,
	featuresTask:    true,
}

type Task struct {
	entry    *normurl.Locator
	resource *normurl.Locator
	taskType taskType
}

func (t *Task) Entry() string {
	if t.entry == nil {
		return ""
	}
	return t.entry.String()
}

func (t *Task) Resource() string {
	if t.resource == nil {
		return ""
	}
	return t.resource.String()
}

func (t *Task) new(resource *normurl.Locator, taskType taskType) *Task {
	return &Task{
		entry:    t.entry,
		resource: resource,
		taskType: taskType,
	}
}

type jsonTask struct {
	Entry    *normurl.Locator
	Resource *normurl.Locator
	Type     string
}

func (t *Task) UnmarshalJSON(data []byte) error {
	var jt jsonTask
	if err := json.Unmarshal(data, &jt); err != nil {
		return err
	}

	t.entry = jt.Entry
	if t.entry == nil {
		return fmt.Errorf("missing entry")
	}

	t.resource = jt.Resource
	if t.resource == nil {
		return fmt.Errorf("missing resource")
	}

	t.taskType = taskType(jt.Type)
	if !validTaskTypes[t.taskType] {
		return fmt.Errorf("invalid task type: %s", t.taskType)
	}

	return nil
}

func (t *Task) MarshalJSON() ([]byte, error) {
	jt := jsonTask{
		Entry:    t.entry,
		Resource: t.resource,
		Type:     string(t.taskType),
	}
	return json.Marshal(jt)
}

type Handler func(task *Task) error

type Queue interface {
	Add(tasks []*Task) error
	Handle(handler Handler)
	Wait() error
}

// NewMemoryQueue is used if a custom queue is not provided for a crawl.
//
// The crawl will stop if the provided context is cancelled.  The limit is used
// to control the number of resources that will be visited concurrently.
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

func (q *memoryQueue) Add(tasks []*Task) error {
	q.mutex.Lock()
	q.buffer = append(q.buffer, tasks...)
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
