package crawler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

	wd, wdErr := os.Getwd()
	require.NoError(t, wdErr)

	_, visitedCatalog := visited.Load(filepath.Join(wd, "testdata/v1.0.0/catalog-with-collection-of-items.json"))
	assert.True(t, visitedCatalog)

	_, visitedCollection := visited.Load(filepath.Join(wd, "testdata/v1.0.0/collection-with-items.json"))
	assert.True(t, visitedCollection)

	_, visitedItem := visited.Load(filepath.Join(wd, "testdata/v1.0.0/item-in-collection.json"))
	assert.True(t, visitedItem)
}

func TestCrawlerFilterItem(t *testing.T) {
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
	c := crawler.New(visitor, &crawler.Options{
		Filter: func(location string) bool {
			return !strings.HasSuffix(location, "/item-in-collection.json")
		},
	})

	err := c.Crawl(context.Background(), "testdata/v1.0.0/catalog-with-collection-of-items.json")
	assert.NoError(t, err)

	assert.Equal(t, uint64(2), count)

	wd, wdErr := os.Getwd()
	require.NoError(t, wdErr)

	_, visitedCatalog := visited.Load(filepath.Join(wd, "testdata/v1.0.0/catalog-with-collection-of-items.json"))
	assert.True(t, visitedCatalog)

	_, visitedCollection := visited.Load(filepath.Join(wd, "testdata/v1.0.0/collection-with-items.json"))
	assert.True(t, visitedCollection)
}

func TestCrawlerFilterCollection(t *testing.T) {
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
	c := crawler.New(visitor, &crawler.Options{
		Filter: func(location string) bool {
			return !strings.HasSuffix(location, "/collection-with-items.json")
		},
	})

	err := c.Crawl(context.Background(), "testdata/v1.0.0/catalog-with-collection-of-items.json")
	assert.NoError(t, err)

	assert.Equal(t, uint64(1), count)

	wd, wdErr := os.Getwd()
	require.NoError(t, wdErr)

	_, visitedCatalog := visited.Load(filepath.Join(wd, "testdata/v1.0.0/catalog-with-collection-of-items.json"))
	assert.True(t, visitedCatalog)
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

func TestCrawlerCollection081(t *testing.T) {
	count := uint64(0)
	visited := &sync.Map{}

	visitor := func(location string, resource crawler.Resource) error {
		atomic.AddUint64(&count, 1)
		_, loaded := visited.LoadOrStore(location, resource)
		if loaded {
			return fmt.Errorf("already visited %s", location)
		}
		return nil
	}
	c := crawler.New(visitor, &crawler.Options{Recursion: crawler.None})

	err := c.Crawl(context.Background(), "testdata/v0.8.1/5633320870809797824_root_collection.json")
	assert.NoError(t, err)

	assert.Equal(t, uint64(1), count)

	wd, wdErr := os.Getwd()
	require.NoError(t, wdErr)

	value, ok := visited.Load(filepath.Join(wd, "testdata/v0.8.1/5633320870809797824_root_collection.json"))
	require.True(t, ok)

	resource, ok := value.(crawler.Resource)
	require.True(t, ok)

	assert.Equal(t, resource.Type(), crawler.Collection)
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

	wd, wdErr := os.Getwd()
	require.NoError(t, wdErr)

	_, visitedCatalog := visited.Load(filepath.Join(wd, "testdata/v1.0.0/api-catalog.json"))
	assert.True(t, visitedCatalog)

	_, visitedCollection := visited.Load(filepath.Join(wd, "testdata/v1.0.0/collection-with-items.json"))
	assert.True(t, visitedCollection)

	_, visitedItem := visited.Load(filepath.Join(wd, "testdata/v1.0.0/item-in-collection.json"))
	assert.True(t, visitedItem)
}

func TestCrawlerAPICollection(t *testing.T) {
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

	err := c.Crawl(context.Background(), "testdata/v1.0.0/api-collection.json")
	assert.NoError(t, err)

	assert.Equal(t, uint64(2), count)

	wd, wdErr := os.Getwd()
	require.NoError(t, wdErr)

	_, visitedCollection := visited.Load(filepath.Join(wd, "testdata/v1.0.0/api-collection.json"))
	assert.True(t, visitedCollection)

	_, visitedItem := visited.Load(filepath.Join(wd, "testdata/v1.0.0/item-in-collection.json"))
	assert.True(t, visitedItem)
}
