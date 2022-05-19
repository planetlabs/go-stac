// Package crawler implements a STAC resource crawler.
package crawler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/planetlabs/go-stac/internal/normurl"
	"github.com/tschaub/retry"
	"github.com/tschaub/workgroup"
)

var httpClient = retryablehttp.NewClient()

func init() {
	httpClient.Logger = nil
}

// RecursionType informs the crawler how to treat linked resources.
// None will only call the visitor for the first resource.  Children
// will call the visitor for all child catalogs, collections, and items.
// All will call the visitor for parent resources as well as child resources.
type RecursionType string

const (
	None     RecursionType = "none"
	Children RecursionType = "children"
)

func loadFile(loc *normurl.Locator, value any) error {
	data, readErr := ioutil.ReadFile(loc.String())
	if readErr != nil {
		return fmt.Errorf("failed to read file %s: %w", loc, readErr)
	}

	jsonErr := json.Unmarshal(data, value)
	if jsonErr != nil {
		return fmt.Errorf("failed to parse %s: %w", loc, jsonErr)
	}
	return nil
}

func loadUrl(loc *normurl.Locator, value any) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()

	retries := 5

	return retry.Limit(ctx, retries, func(ctx context.Context, attempt int) error {
		err := tryLoadUrl(loc, value)
		if err == nil {
			return nil
		}

		// these come when parsing the response body
		if !errors.Is(err, syscall.ECONNRESET) {
			return retry.Stop(err)
		}

		jitter := time.Duration(rand.Float64()) * time.Second
		time.Sleep(time.Second*time.Duration(math.Pow(2, float64(attempt))) + jitter)
		return err
	})
}

func tryLoadUrl(loc *normurl.Locator, value any) error {
	resp, err := httpClient.Get(loc.String())
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return fmt.Errorf("unexpected response for %s: %d", loc, resp.StatusCode)
	}

	jsonErr := json.NewDecoder(resp.Body).Decode(value)
	if jsonErr != nil {
		return fmt.Errorf("failed to parse %s: %w", loc, jsonErr)
	}
	return nil
}

// Visitor is called for each resource during crawling.
//
// The resource location (URL or file path) is passed as the first argument.
// Any returned error will stop crawling and be returned by Crawl.
type Visitor func(string, Resource) error

// crawler crawls STAC resources.
type crawler struct {
	ctx         context.Context
	entry       *normurl.Locator
	visitor     Visitor
	worker      *workgroup.Worker[*Task]
	recursion   RecursionType
	concurrency int
	filter      func(string) bool
	queue       workgroup.Queue[*Task]
}

// Options for creating a crawler.
type Options struct {
	// Optional context.  If provided, the crawler will stop when the context is done.
	Context context.Context

	// Limit to the number of resources to fetch and visit concurrently.
	Concurrency int

	// Strategy to use when crawling linked resources.  Use None to visit
	// a single resource.  Use Children to only visit linked item/child resources.
	Recursion RecursionType

	// Optional function to limit which resources to crawl.  If provided, the function
	// will be called with the URL or absolute path to a resource before it is crawled.
	// If the function returns false, the resource will not be read and the visitor will
	// not be called.
	Filter func(string) bool

	// Optional queue to use for crawling tasks.  If not provided, an in-memory queue
	// will be used.  When running a crawl across multiple processes, it can be useful
	// to provide a queue that is shared across processes.
	Queue workgroup.Queue[*Task]
}

func (c *crawler) apply(options *Options) {
	if options.Context != nil {
		c.ctx = options.Context
	}
	if options.Concurrency > 0 {
		c.concurrency = options.Concurrency
	}
	if options.Recursion != "" {
		c.recursion = options.Recursion
	}
	if options.Queue != nil {
		c.queue = options.Queue
	}
	if options.Filter != nil {
		c.filter = options.Filter
	}
}

// DefaultOptions used when creating a new crawler.
var DefaultOptions = &Options{
	Context:     context.Background(),
	Recursion:   Children,
	Concurrency: runtime.GOMAXPROCS(0),
}

// Crawl calls the visitor for each resolved resource.
//
// The resource can be a file path or a URL.  Any error returned by visitor
// will stop crawling and be returned by this function.  Context cancellation
// will also stop crawling and the context error will be returned.
func Crawl(resource string, visitor Visitor, options ...*Options) error {
	c, err := newCrawler(resource, visitor, options...)
	if err != nil {
		return err
	}
	addErr := c.worker.Add(&Task{Url: c.entry.String(), Type: resourceTask})
	if addErr != nil {
		return addErr
	}
	return c.worker.Wait()
}

// Join adds a crawler to an existing crawl instead of initiating a new one.
//
// This is useful when two crawlers share the same queue and the first crawler
// has been taken down or the queue is getting too large.  Join will return immediately
// if the queue of tasks is empty.
func Join(resource string, visitor Visitor, options ...*Options) error {
	c, err := newCrawler(resource, visitor, options...)
	if err != nil {
		return err
	}

	return c.worker.Wait()
}

// newCrawler creates a crawler with the provided options (or DefaultOptions
// if none are provided).
//
// The visitor will be called for each resource resolved.
func newCrawler(resource string, visitor Visitor, options ...*Options) (*crawler, error) {
	wd, wdErr := os.Getwd()
	if wdErr != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", wdErr)
	}
	base, baseErr := normurl.New(fmt.Sprintf("%s%c", wd, os.PathSeparator))
	if baseErr != nil {
		return nil, fmt.Errorf("failed to parse working directory: %w", baseErr)
	}

	loc, locErr := base.Resolve(resource)
	if locErr != nil {
		return nil, locErr
	}

	c := &crawler{
		entry:   loc,
		visitor: visitor,
	}

	c.apply(DefaultOptions)
	for _, opt := range options {
		c.apply(opt)
	}

	c.worker = workgroup.New(&workgroup.Options[*Task]{
		Context: c.ctx,
		Limit:   c.concurrency,
		Work:    c.crawl,
		Queue:   c.queue,
	})

	return c, nil
}

const (
	resourceTask    = "resource"
	collectionsTask = "collections"
	featuresTask    = "features"
)

type Task struct {
	Url  string
	Type string
}

func (c *crawler) load(loc *normurl.Locator, value interface{}) error {
	if loc.IsFilepath() {
		if !c.entry.IsFilepath() {
			return fmt.Errorf("cannot crawl file %s in non-file mode", loc)
		}
		return loadFile(loc, value)
	}

	if c.entry.IsFilepath() {
		return fmt.Errorf("cannot crawl URL %s in file mode", loc)
	}
	return loadUrl(loc, value)
}

func (c *crawler) crawl(worker *workgroup.Worker[*Task], t *Task) error {
	if c.filter != nil && !c.filter(t.Url) {
		return nil
	}
	switch t.Type {
	case resourceTask:
		return c.crawlResource(worker, t.Url)
	case collectionsTask:
		return c.crawlCollections(worker, t.Url)
	case featuresTask:
		return c.crawlFeatures(worker, t.Url)
	default:
		return fmt.Errorf("unknown task type: %s", t.Type)
	}
}

func (c *crawler) crawlResource(worker *workgroup.Worker[*Task], resourceUrl string) error {
	loc, locErr := normurl.New(resourceUrl)
	if locErr != nil {
		return locErr
	}

	resource := Resource{}
	loadErr := c.load(loc, &resource)
	if loadErr != nil {
		return loadErr
	}

	if c.recursion == None {
		return c.visitor(resourceUrl, resource)
	}

	links := resource.Links()
	// check if this looks like a STAC API root catalog that implements OGC API - Features
	if resource.Type() == Catalog && len(resource.ConformsTo()) > 1 {
		dataLink := links.Rel("data", LinkTypeApplicationJSON, LinkTypeAnyJSON, LinkTypeNone)
		if dataLink != nil {
			linkLoc, err := loc.Resolve(dataLink["href"])
			if err != nil {
				return err
			}
			addErr := worker.Add(&Task{Url: linkLoc.String(), Type: collectionsTask})
			if addErr != nil {
				return addErr
			}
			return c.visitor(resourceUrl, resource)
		}
	}

	if resource.Type() == Collection {
		// shortcut for "items" link
		itemsLink := links.Rel("items", LinkTypeGeoJSON, LinkTypeApplicationJSON, LinkTypeAnyJSON, LinkTypeNone)
		if itemsLink != nil {
			linkLoc, err := loc.Resolve(itemsLink["href"])
			if err != nil {
				return err
			}
			addErr := worker.Add(&Task{Url: linkLoc.String(), Type: featuresTask})
			if addErr != nil {
				return addErr
			}
			return c.visitor(resourceUrl, resource)
		}
	}

	for _, link := range links {
		rel := link["rel"]
		if rel == "item" || rel == "child" {
			if LinkTypeApplicationJSON(link) || LinkTypeAnyJSON(link) || LinkTypeNone(link) {
				linkLoc, err := loc.Resolve(link["href"])
				if err != nil {
					return err
				}
				addErr := worker.Add(&Task{Url: linkLoc.String(), Type: resourceTask})
				if addErr != nil {
					return addErr
				}
			}
		}
	}

	return c.visitor(resourceUrl, resource)
}

func (c *crawler) crawlCollections(worker *workgroup.Worker[*Task], collectionsUrl string) error {
	loc, locErr := normurl.New(collectionsUrl)
	if locErr != nil {
		return locErr
	}
	response := &featureCollectionsResponse{}
	loadErr := c.load(loc, response)
	if loadErr != nil {
		return loadErr
	}

	for i, resource := range response.Collections {
		if resource.Type() != Collection {
			return fmt.Errorf("expected collection at index %d, got %s", i, resource.Type())
		}
		links := resource.Links()

		selfLink := links.Rel("self", LinkTypeApplicationJSON, LinkTypeAnyJSON, LinkTypeNone)
		if selfLink == nil {
			return fmt.Errorf("missing self link for collection %d in %s", i, collectionsUrl)
		}
		selfLinkLoc, selfLinkErr := loc.Resolve(selfLink["href"])
		if selfLinkErr != nil {
			return selfLinkErr
		}
		if err := c.visitor(selfLinkLoc.String(), resource); err != nil {
			return err
		}

		itemsLink := links.Rel("items", LinkTypeGeoJSON, LinkTypeApplicationJSON, LinkTypeAnyJSON, LinkTypeNone)
		if itemsLink == nil {
			return fmt.Errorf("missing items link for collection %d in %s", i, collectionsUrl)
		}
		itemsLinkLoc, itemsLinkErr := loc.Resolve(itemsLink["href"])
		if itemsLinkErr != nil {
			return itemsLinkErr
		}
		addErr := worker.Add(&Task{Url: itemsLinkLoc.String(), Type: featuresTask})
		if addErr != nil {
			return addErr
		}
	}

	nextLink := response.Links.Rel("next", LinkTypeApplicationJSON, LinkTypeAnyJSON, LinkTypeNone)
	if nextLink != nil {
		linkLoc, err := loc.Resolve(nextLink["href"])
		if err != nil {
			return err
		}
		addErr := worker.Add(&Task{Url: linkLoc.String(), Type: collectionsTask})
		if addErr != nil {
			return addErr
		}
	}

	return nil
}

func (c *crawler) crawlFeatures(worker *workgroup.Worker[*Task], featuresUrl string) error {
	loc, locErr := normurl.New(featuresUrl)
	if locErr != nil {
		return locErr
	}

	response := &featureCollectionResponse{}
	loadErr := c.load(loc, response)
	if loadErr != nil {
		return loadErr
	}
	for i, resource := range response.Features {
		if resource.Type() != Item {
			return fmt.Errorf("expected item at index %d, got %s", i, resource.Type())
		}

		links := resource.Links()
		selfLink := links.Rel("self", LinkTypeGeoJSON, LinkTypeApplicationJSON, LinkTypeAnyJSON, LinkTypeNone)
		if selfLink == nil {
			return fmt.Errorf("missing self link for item %d in %s", i, featuresUrl)
		}

		selfLinkLoc, selfLinkErr := loc.Resolve(selfLink["href"])
		if selfLinkErr != nil {
			return selfLinkErr
		}

		if err := c.visitor(selfLinkLoc.String(), resource); err != nil {
			return err
		}
	}

	nextLink := response.Links.Rel("next", LinkTypeApplicationJSON, LinkTypeAnyJSON, LinkTypeNone)
	if nextLink != nil {
		linkLoc, err := loc.Resolve(nextLink["href"])
		if err != nil {
			return err
		}
		addErr := worker.Add(&Task{Url: linkLoc.String(), Type: featuresTask})
		if addErr != nil {
			return addErr
		}
	}

	return nil
}
