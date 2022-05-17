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
	"strings"
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

// Crawler crawls STAC resources.
type Crawler struct {
	fileMode    bool
	visitor     Visitor
	recursion   RecursionType
	concurrency int
	filter      func(string) bool
	queue       workgroup.Queue[*Task]
}

// Options for creating a crawler.
type Options struct {
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

	Queue workgroup.Queue[*Task]
}

func (c *Crawler) apply(options *Options) {
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
	Recursion:   Children,
	Concurrency: runtime.GOMAXPROCS(0),
}

// New creates a crawler with the provided options (or DefaultOptions
// if none are provided).
//
// The visitor will be called for each resource resolved.
func New(visitor Visitor, options ...*Options) *Crawler {
	c := &Crawler{
		visitor: visitor,
	}
	c.apply(DefaultOptions)
	for _, opt := range options {
		c.apply(opt)
	}
	return c
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

// Crawl calls the visitor for each resolved resource.
//
// The resource can be a file path or a URL.  Any error returned by visitor
// will stop crawling and be returned by this function.  Context cancellation
// will also stop crawling and the context error will be returned.
func (c *Crawler) Crawl(ctx context.Context, resource string) error {
	wd, wdErr := os.Getwd()
	if wdErr != nil {
		return fmt.Errorf("failed to get working directory: %w", wdErr)
	}
	base, baseErr := normurl.New(fmt.Sprintf("%s%c", wd, os.PathSeparator))
	if baseErr != nil {
		return fmt.Errorf("failed to parse working directory: %w", baseErr)
	}

	loc, locErr := base.Resolve(resource)
	if locErr != nil {
		return locErr
	}
	c.fileMode = loc.IsFilepath()
	worker := workgroup.New(&workgroup.Options[*Task]{
		Context: ctx,
		Limit:   c.concurrency,
		Work:    c.crawl,
		Queue:   c.queue,
	})
	addErr := worker.Add(&Task{Url: loc.String(), Type: resourceTask})
	if addErr != nil {
		return addErr
	}
	return worker.Wait()
}

func (c *Crawler) load(loc *normurl.Locator, value interface{}) error {
	if loc.IsFilepath() {
		if !c.fileMode {
			return fmt.Errorf("cannot crawl file %s in non-file mode", loc)
		}
		return loadFile(loc, value)
	}

	if c.fileMode {
		return fmt.Errorf("cannot crawl URL %s in file mode", loc)
	}
	return loadUrl(loc, value)
}

func (c *Crawler) crawl(worker *workgroup.Worker[*Task], t *Task) error {
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

func (c *Crawler) crawlResource(worker *workgroup.Worker[*Task], resourceUrl string) error {
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

	// check if this looks like a STAC API root catalog that implements OGC API - Features
	if resource.Type() == Catalog && len(resource.ConformsTo()) > 1 {
		for _, link := range resource.Links() {
			if link["rel"] == "data" {
				linkLoc, err := loc.Resolve(link["href"])
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
	}

	if resource.Type() == Collection {
		// shortcut for "items" link
		for _, link := range resource.Links() {
			if link["rel"] == "items" {
				linkLoc, err := loc.Resolve(link["href"])
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
	}

	for _, link := range resource.Links() {
		linkLoc, err := loc.Resolve(link["href"])
		if err != nil {
			return err
		}
		rel := link["rel"]
		if rel == "item" || rel == "child" {
			addErr := worker.Add(&Task{Url: linkLoc.String(), Type: resourceTask})
			if addErr != nil {
				return addErr
			}
		}
	}

	return c.visitor(resourceUrl, resource)
}

func (c *Crawler) crawlCollections(worker *workgroup.Worker[*Task], collectionsUrl string) error {
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
		var selfUrl string
		var itemsUrl string
		for _, link := range resource.Links() {
			rel := link["rel"]
			linkType := link["type"]
			if selfUrl == "" && rel == "self" && strings.HasSuffix(linkType, "json") {
				linkLoc, err := loc.Resolve(link["href"])
				if err != nil {
					return err
				}
				selfUrl = linkLoc.String()
			}
			if itemsUrl == "" && rel == "items" && strings.HasSuffix(linkType, "json") {
				linkLoc, err := loc.Resolve(link["href"])
				if err != nil {
					return err
				}
				itemsUrl = linkLoc.String()
			}
		}
		if selfUrl == "" {
			return fmt.Errorf("missing self link for collection %d in %s", i, collectionsUrl)
		}
		if itemsUrl == "" {
			return fmt.Errorf("missing items link for collection %d in %s", i, collectionsUrl)
		}
		if err := c.visitor(selfUrl, resource); err != nil {
			return err
		}
		addErr := worker.Add(&Task{Url: itemsUrl, Type: featuresTask})
		if addErr != nil {
			return addErr
		}
	}

	for _, link := range response.Links {
		if link["rel"] == "next" && strings.HasSuffix(link["type"], "json") {
			linkLoc, err := loc.Resolve(link["href"])
			if err != nil {
				return err
			}
			addErr := worker.Add(&Task{Url: linkLoc.String(), Type: collectionsTask})
			if addErr != nil {
				return addErr
			}
			break
		}
	}
	return nil
}

func (c *Crawler) crawlFeatures(worker *workgroup.Worker[*Task], featuresUrl string) error {
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

		var itemUrl string
		for _, link := range resource.Links() {
			if link["rel"] == "self" && strings.HasSuffix(link["type"], "json") {
				linkLoc, err := loc.Resolve(link["href"])
				if err != nil {
					return err
				}
				itemUrl = linkLoc.String()
				break
			}
		}
		if itemUrl == "" {
			return fmt.Errorf("missing self link for item %d in %s", i, featuresUrl)
		}

		if err := c.visitor(itemUrl, resource); err != nil {
			return err
		}
	}

	for _, link := range response.Links {
		if link["rel"] == "next" && strings.HasSuffix(link["type"], "json") {
			linkLoc, err := loc.Resolve(link["href"])
			if err != nil {
				return err
			}
			addErr := worker.Add(&Task{Url: linkLoc.String(), Type: featuresTask})
			if addErr != nil {
				return addErr
			}
			break
		}
	}
	return nil
}
