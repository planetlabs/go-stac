package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/planetlabs/go-stac/internal/normurl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskDetail(t *testing.T) {
	entry, entryErr := normurl.New("https://example.com/")
	require.NoError(t, entryErr)

	resource, resourceErr := normurl.New("https://example.com/resource")
	require.NoError(t, resourceErr)

	task := &Task{entry: entry, resource: resource, taskType: resourceTask}

	assert.Equal(t, "https://example.com/", task.Entry())
	assert.Equal(t, "https://example.com/resource", task.Resource())
}

func TestTaskMarshalJSON(t *testing.T) {
	entry, entryErr := normurl.New("https://example.com/")
	require.NoError(t, entryErr)

	resource, resourceErr := normurl.New("https://example.com/resource")
	require.NoError(t, resourceErr)

	task := &Task{entry: entry, resource: resource, taskType: resourceTask}

	data, err := json.Marshal(task)
	require.NoError(t, err)

	expected := `{
		"Entry": {
			"Url": "https://example.com/",
			"File": false
		},
		"Resource": {
			"Url": "https://example.com/resource",
			"File": false
		},
		"Type": "resource"
	}`
	assert.JSONEq(t, expected, string(data))
}

func TestTaskUnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"Entry": {
			"Url": "/path/to/entry",
			"File": true
		},
		"Resource": {
			"Url": "/path/to/resource",
			"File": true
		},
		"Type": "resource"
	}`)

	task := &Task{}
	jsonErr := json.Unmarshal(data, task)
	require.NoError(t, jsonErr)

	entry, entryErr := normurl.New("/path/to/entry")
	require.NoError(t, entryErr)

	resource, resourceErr := normurl.New("/path/to/resource")
	require.NoError(t, resourceErr)

	expected := &Task{entry: entry, resource: resource, taskType: resourceTask}

	assert.Equal(t, expected, task)
}

func TestMemoryQueue(t *testing.T) {
	queue := NewMemoryQueue(context.Background(), 3)

	urls := []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://example.com/3",
		"https://example.com/4",
		"https://example.com/5",
		"https://example.com/6",
		"https://example.com/7",
	}

	visited := sync.Map{}
	queue.Handle(func(task *Task) error {
		_, already := visited.LoadOrStore(task.resource.String(), true)
		if already {
			return fmt.Errorf("already visited %s", task.resource.String())
		}
		return nil
	})

	entry, entryErr := normurl.New("https://example.com/")
	require.NoError(t, entryErr)

	tasks := make([]*Task, len(urls))
	for i, url := range urls {
		resource, resourceErr := normurl.New(url)
		require.NoError(t, resourceErr)
		tasks[i] = &Task{entry: entry, resource: resource, taskType: resourceTask}
	}
	require.NoError(t, queue.Add(tasks))
	require.NoError(t, queue.Wait())

	for _, url := range urls {
		_, ok := visited.Load(url)
		assert.True(t, ok, url)
	}
}

func TestMemoryQueueRecursive(t *testing.T) {
	depth := 4
	width := 3

	queue := NewMemoryQueue(context.Background(), 5)
	visited := sync.Map{}

	count := int64(0)
	queue.Handle(func(task *Task) error {
		_, already := visited.LoadOrStore(task.resource.String(), true)
		if already {
			return fmt.Errorf("already visited %s", task.resource.String())
		}
		atomic.AddInt64(&count, 1)

		parts := strings.Split(task.resource.String(), "-")
		d := len(parts)
		if d >= depth {
			return nil
		}

		tasks := make([]*Task, width)
		for i := 0; i < width; i++ {
			resource, resourceErr := normurl.New(fmt.Sprintf("%s-%d", task.resource.String(), i))
			require.NoError(t, resourceErr)
			tasks[i] = task.new(resource, resourceTask)
		}
		return queue.Add(tasks)
	})

	entry, entryErr := normurl.New("https://example.com/0")
	require.NoError(t, entryErr)

	require.NoError(t, queue.Add([]*Task{{entry: entry, resource: entry, taskType: resourceTask}}))
	require.NoError(t, queue.Wait())

	expected := int64(math.Pow(float64(width), float64(depth))-1) / int64(width-1)
	assert.Equal(t, expected, count)
}

func TestMemoryQueueError(t *testing.T) {
	expectedError := fmt.Errorf("expected error")
	queue := NewMemoryQueue(context.Background(), 3)

	visited := sync.Map{}
	queue.Handle(func(task *Task) error {
		_, already := visited.LoadOrStore(task.resource.String(), true)
		if already {
			return fmt.Errorf("already visited %s", task.resource.String())
		}
		if string(task.taskType) == "error" {
			return expectedError
		}
		return nil
	})

	entry, entryErr := normurl.New("https://example.com/")
	require.NoError(t, entryErr)

	tasks := make([]*Task, 100)
	for i := 0; i < len(tasks); i++ {
		resource, resourceErr := normurl.New(fmt.Sprintf("https://example.com/%d", i))
		require.NoError(t, resourceErr)

		tt := taskType("test")
		if i == 42 {
			tt = taskType("error")
		}
		tasks[i] = &Task{entry: entry, resource: resource, taskType: tt}
	}
	require.NoError(t, queue.Add(tasks))

	err := queue.Wait()
	assert.Equal(t, expectedError, err)
}
