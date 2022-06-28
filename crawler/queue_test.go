package crawler_test

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/planetlabs/go-stac/crawler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryQueue(t *testing.T) {
	queue := crawler.NewMemoryQueue(context.Background(), 3)

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
	queue.Handle(func(task *crawler.Task) error {
		_, already := visited.LoadOrStore(task.Url, true)
		if already {
			return fmt.Errorf("already visited %s", task.Url)
		}
		return nil
	})

	for _, url := range urls {
		err := queue.Add(&crawler.Task{Url: url, Type: "test"})
		require.NoError(t, err)
	}

	require.NoError(t, queue.Wait())

	for _, url := range urls {
		_, ok := visited.Load(url)
		assert.True(t, ok, url)
	}
}

func TestMemoryQueueRecursive(t *testing.T) {
	depth := 4
	width := 3

	queue := crawler.NewMemoryQueue(context.Background(), 5)
	visited := sync.Map{}

	count := int64(0)
	queue.Handle(func(task *crawler.Task) error {
		_, already := visited.LoadOrStore(task.Url, true)
		if already {
			return fmt.Errorf("already visited %s", task.Url)
		}
		atomic.AddInt64(&count, 1)

		parts := strings.Split(task.Url, "/")
		d := len(parts)
		if d >= depth {
			return nil
		}

		for i := 0; i < width; i++ {
			url := fmt.Sprintf("%s/%d", task.Url, i)
			err := queue.Add(&crawler.Task{Url: url})
			if err != nil {
				return err
			}
		}

		return nil
	})

	require.NoError(t, queue.Add(&crawler.Task{Url: "0"}))
	require.NoError(t, queue.Wait())

	expected := int64(math.Pow(float64(width), float64(depth))-1) / int64(width-1)
	assert.Equal(t, expected, count)
}

func TestMemoryQueueError(t *testing.T) {
	expectedError := fmt.Errorf("expected error")
	queue := crawler.NewMemoryQueue(context.Background(), 3)

	visited := sync.Map{}
	queue.Handle(func(task *crawler.Task) error {
		_, already := visited.LoadOrStore(task.Url, true)
		if already {
			return fmt.Errorf("already visited %s", task.Url)
		}
		if task.Type == "error" {
			return expectedError
		}
		return nil
	})

	for i := 0; i < 100; i++ {
		url := fmt.Sprintf("https://example.com/%d", i)
		taskType := "test"
		if i == 42 {
			taskType = "error"
		}
		require.NoError(t, queue.Add(&crawler.Task{Url: url, Type: taskType}))
	}

	err := queue.Wait()
	assert.Equal(t, expectedError, err)
}
