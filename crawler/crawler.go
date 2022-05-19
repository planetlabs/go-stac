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

func load(entry *normurl.Locator, loc *normurl.Locator, value interface{}) error {
	if loc.IsFilepath() {
		if !entry.IsFilepath() {
			return fmt.Errorf("cannot crawl file %s in non-file mode", loc)
		}
		return loadFile(loc, value)
	}

	if entry.IsFilepath() {
		return fmt.Errorf("cannot crawl URL %s in file mode", loc)
	}
	return loadUrl(loc, value)
}

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

// ErrorHandler is called with any errors during a crawl.  If the function
// returns nil, the crawl will continue.  If the function returns an error,
// the crawl will stop.
type ErrorHandler func(error) error

func wrapErrorHandler(handler ErrorHandler) ErrorHandler {
	return func(err error) error {
		if err == nil {
			return nil
		}
		return handler(err)
	}
}

// crawler crawls STAC resources.
type crawler struct {
	entry        *normurl.Locator
	visitor      Visitor
	worker       *workgroup.Worker[*Task]
	recursion    RecursionType
	filter       func(string) bool
	errorHandler ErrorHandler
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

	// Optional function to handle any errors during the crawl.  By default, any error
	// will stop the crawl.  To continue crawling on error, provide a function that
	// returns nil.
	ErrorHandler ErrorHandler

	// Optional queue to use for crawling tasks.  If not provided, an in-memory queue
	// will be used.  When running a crawl across multiple processes, it can be useful
	// to provide a queue that is shared across processes.
	Queue workgroup.Queue[*Task]
}

func applyOptions(options []*Options) *Options {
	o := &Options{}
	for _, option := range options {
		if option.Context != nil {
			o.Context = option.Context
		}
		if option.Concurrency > 0 {
			o.Concurrency = option.Concurrency
		}
		if option.Recursion != "" {
			o.Recursion = option.Recursion
		}
		if option.Queue != nil {
			o.Queue = option.Queue
		}
		if option.Filter != nil {
			o.Filter = option.Filter
		}
		if option.ErrorHandler != nil {
			o.ErrorHandler = option.ErrorHandler
		}
	}
	return o
}

// DefaultOptions used when creating a new crawler.
var DefaultOptions = &Options{
	Context:      context.Background(),
	Recursion:    Children,
	Concurrency:  runtime.GOMAXPROCS(0),
	ErrorHandler: func(err error) error { return err },
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

	opt := applyOptions(append([]*Options{DefaultOptions}, options...))

	c := &crawler{
		entry:        loc,
		visitor:      visitor,
		recursion:    opt.Recursion,
		filter:       opt.Filter,
		errorHandler: wrapErrorHandler(opt.ErrorHandler),
	}

	c.worker = workgroup.New(&workgroup.Options[*Task]{
		Work:    c.crawl,
		Context: opt.Context,
		Limit:   opt.Concurrency,
		Queue:   opt.Queue,
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
	loadErr := load(c.entry, loc, &resource)
	if loadErr != nil {
		return c.errorHandler(loadErr)
	}

	if err := c.errorHandler(c.visitor(resourceUrl, resource)); err != nil {
		return err
	}

	if c.recursion == None {
		return nil
	}

	links := resource.Links()
	// check if this looks like a STAC API root catalog that implements OGC API - Features
	if resource.Type() == Catalog && len(resource.ConformsTo()) > 1 {
		dataLink := links.Rel("data", LinkTypeApplicationJSON, LinkTypeAnyJSON, LinkTypeNone)
		if dataLink != nil {
			linkLoc, err := loc.Resolve(dataLink["href"])
			if err != nil {
				return c.errorHandler(err)
			}
			return worker.Add(&Task{Url: linkLoc.String(), Type: collectionsTask})
		}
	}

	if resource.Type() == Collection {
		// shortcut for "items" link
		itemsLink := links.Rel("items", LinkTypeGeoJSON, LinkTypeApplicationJSON, LinkTypeAnyJSON, LinkTypeNone)
		if itemsLink != nil {
			linkLoc, err := loc.Resolve(itemsLink["href"])
			if err != nil {
				return c.errorHandler(err)
			}
			return worker.Add(&Task{Url: linkLoc.String(), Type: featuresTask})
		}
	}

	for _, link := range links {
		rel := link["rel"]
		if rel == "item" || rel == "child" {
			if LinkTypeApplicationJSON(link) || LinkTypeAnyJSON(link) || LinkTypeNone(link) {
				linkLoc, err := loc.Resolve(link["href"])
				if err != nil {
					return c.errorHandler(err)
				}
				addErr := worker.Add(&Task{Url: linkLoc.String(), Type: resourceTask})
				if addErr != nil {
					return addErr
				}
			}
		}
	}

	return nil
}

func (c *crawler) crawlCollections(worker *workgroup.Worker[*Task], collectionsUrl string) error {
	loc, locErr := normurl.New(collectionsUrl)
	if locErr != nil {
		return locErr
	}
	response := &featureCollectionsResponse{}
	loadErr := load(c.entry, loc, response)
	if loadErr != nil {
		return c.errorHandler(loadErr)
	}

	for i, resource := range response.Collections {
		if resource.Type() != Collection {
			return c.errorHandler(fmt.Errorf("expected collection at index %d, got %s", i, resource.Type()))
		}
		links := resource.Links()

		selfLink := links.Rel("self", LinkTypeApplicationJSON, LinkTypeAnyJSON, LinkTypeNone)
		if selfLink == nil {
			return c.errorHandler(fmt.Errorf("missing self link for collection %d in %s", i, collectionsUrl))
		}
		selfLinkLoc, selfLinkErr := loc.Resolve(selfLink["href"])
		if selfLinkErr != nil {
			return c.errorHandler(selfLinkErr)
		}
		if err := c.errorHandler(c.visitor(selfLinkLoc.String(), resource)); err != nil {
			return err
		}

		itemsLink := links.Rel("items", LinkTypeGeoJSON, LinkTypeApplicationJSON, LinkTypeAnyJSON, LinkTypeNone)
		if itemsLink == nil {
			return c.errorHandler(fmt.Errorf("missing items link for collection %d in %s", i, collectionsUrl))
		}
		itemsLinkLoc, itemsLinkErr := loc.Resolve(itemsLink["href"])
		if itemsLinkErr != nil {
			return c.errorHandler(itemsLinkErr)
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
			return c.errorHandler(err)
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
		return c.errorHandler(locErr)
	}

	response := &featureCollectionResponse{}
	loadErr := load(c.entry, loc, response)
	if loadErr != nil {
		return c.errorHandler(loadErr)
	}
	for i, resource := range response.Features {
		if resource.Type() != Item {
			return c.errorHandler(fmt.Errorf("expected item at index %d, got %s", i, resource.Type()))
		}

		links := resource.Links()
		selfLink := links.Rel("self", LinkTypeGeoJSON, LinkTypeApplicationJSON, LinkTypeAnyJSON, LinkTypeNone)
		if selfLink == nil {
			return c.errorHandler(fmt.Errorf("missing self link for item %d in %s", i, featuresUrl))
		}

		selfLinkLoc, selfLinkErr := loc.Resolve(selfLink["href"])
		if selfLinkErr != nil {
			return c.errorHandler(selfLinkErr)
		}

		if err := c.errorHandler(c.visitor(selfLinkLoc.String(), resource)); err != nil {
			return err
		}
	}

	nextLink := response.Links.Rel("next", LinkTypeApplicationJSON, LinkTypeAnyJSON, LinkTypeNone)
	if nextLink != nil {
		linkLoc, err := loc.Resolve(nextLink["href"])
		if err != nil {
			return c.errorHandler(err)
		}
		addErr := worker.Add(&Task{Url: linkLoc.String(), Type: featuresTask})
		if addErr != nil {
			return addErr
		}
	}

	return nil
}
