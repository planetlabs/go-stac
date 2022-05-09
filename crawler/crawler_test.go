package crawler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/planetlabs/go-stac/crawler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrawler(t *testing.T) {
	count := uint64(0)
	visited := &sync.Map{}

	visitor := func(location string, resource crawler.Resource) error {
		atomic.AddUint64(&count, 1)
		_, loaded := visited.LoadOrStore(location, true)
		if loaded {
			return fmt.Errorf("already visited %s", location)
		}
		return nil
	}
	c := crawler.New(visitor)

	err := c.Crawl(context.Background(), "testdata/v1.0.0/catalog-with-collection-of-items.json")
	assert.NoError(t, err)

	assert.Equal(t, uint64(3), count)
}

func TestCrawlerHTTP(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer server.Close()

	count := uint64(0)
	visited := &sync.Map{}

	visitor := func(location string, resource crawler.Resource) error {
		atomic.AddUint64(&count, 1)
		_, loaded := visited.LoadOrStore(location, true)
		if loaded {
			return fmt.Errorf("already visited %s", location)
		}
		return nil
	}
	c := crawler.New(visitor)

	err := c.Crawl(context.Background(), server.URL+"/v1.0.0/catalog-with-collection-of-items.json")
	assert.NoError(t, err)

	assert.Equal(t, uint64(3), count)

	_, visitedCatalog := visited.Load(server.URL + "/v1.0.0/catalog-with-collection-of-items.json")
	assert.True(t, visitedCatalog)

	_, visitedCollection := visited.Load(server.URL + "/v1.0.0/collection-with-items.json")
	assert.True(t, visitedCollection)

	_, visitedItem := visited.Load(server.URL + "/v1.0.0/item-in-collection.json")
	assert.True(t, visitedItem)
}

func TestCrawlerHTTPRetry(t *testing.T) {

	tried := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !tried {
			tried = true
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNotImplemented) // stop trying
	}))
	defer server.Close()

	count := uint64(0)
	visited := &sync.Map{}
	visitor := func(location string, resource crawler.Resource) error {
		atomic.AddUint64(&count, 1)
		_, loaded := visited.LoadOrStore(location, true)
		if loaded {
			return fmt.Errorf("already visited %s", location)
		}
		return nil
	}
	c := crawler.New(visitor)

	err := c.Crawl(context.Background(), server.URL+"/not-found")
	require.Error(t, err)

	assert.True(t, strings.HasPrefix(err.Error(), "unexpected response"))
	assert.Equal(t, uint64(0), count)
	assert.True(t, tried)
}

func TestCrawlerSingle(t *testing.T) {
	count := uint64(0)
	visited := &sync.Map{}

	visitor := func(location string, resource crawler.Resource) error {
		atomic.AddUint64(&count, 1)
		_, loaded := visited.LoadOrStore(location, true)
		if loaded {
			return fmt.Errorf("already visited %s", location)
		}
		return nil
	}
	c := crawler.New(visitor, &crawler.Options{Recursion: crawler.None})

	err := c.Crawl(context.Background(), "testdata/v1.0.0/catalog-with-collection-of-items.json")
	assert.NoError(t, err)

	assert.Equal(t, uint64(1), count)
}

func TestCrawlerChildren(t *testing.T) {
	count := uint64(0)
	visited := &sync.Map{}

	visitor := func(location string, resource crawler.Resource) error {
		atomic.AddUint64(&count, 1)
		_, loaded := visited.LoadOrStore(location, true)
		if loaded {
			return fmt.Errorf("already visited %s", location)
		}
		return nil
	}
	c := crawler.New(visitor, &crawler.Options{Recursion: crawler.Children})

	err := c.Crawl(context.Background(), "testdata/v1.0.0/collection-with-items.json")
	assert.NoError(t, err)

	assert.Equal(t, uint64(2), count)
}

func TestCrawlerAll(t *testing.T) {
	count := uint64(0)
	visited := &sync.Map{}

	visitor := func(location string, resource crawler.Resource) error {
		atomic.AddUint64(&count, 1)
		_, loaded := visited.LoadOrStore(location, true)
		if loaded {
			return fmt.Errorf("already visited %s", location)
		}
		return nil
	}
	c := crawler.New(visitor, &crawler.Options{Recursion: crawler.All})

	err := c.Crawl(context.Background(), "testdata/v1.0.0/item-in-collection.json")
	assert.NoError(t, err)

	assert.Equal(t, uint64(3), count)
}

func TestCrawlerCatalog(t *testing.T) {
	count := uint64(0)
	visited := &sync.Map{}

	visitor := func(location string, resource crawler.Resource) error {
		atomic.AddUint64(&count, 1)
		_, loaded := visited.LoadOrStore(location, true)
		if loaded {
			return fmt.Errorf("already visited %s", location)
		}
		return nil
	}
	c := crawler.New(visitor)

	err := c.Crawl(context.Background(), "testdata/v1.0.0/catalog.json")
	assert.NoError(t, err)

	assert.Equal(t, uint64(1), count)
}

func TestCrawlerAPI(t *testing.T) {
	count := uint64(0)
	visited := &sync.Map{}

	visitor := func(location string, resource crawler.Resource) error {
		atomic.AddUint64(&count, 1)
		_, loaded := visited.LoadOrStore(location, true)
		if loaded {
			return fmt.Errorf("already visited %s", location)
		}
		return nil
	}
	c := crawler.New(visitor)

	err := c.Crawl(context.Background(), "testdata/v1.0.0/api-catalog.json")
	assert.NoError(t, err)

	assert.Equal(t, uint64(3), count)

	_, visitedCatalog := visited.Load("testdata/v1.0.0/api-catalog.json")
	assert.True(t, visitedCatalog)

	_, visitedCollection := visited.Load("testdata/v1.0.0/collection-with-items.json")
	assert.True(t, visitedCollection)

	_, visitedItem := visited.Load("testdata/v1.0.0/item-in-collection.json")
	assert.True(t, visitedItem)
}
