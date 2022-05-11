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
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/go-retryablehttp"
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
	All      RecursionType = "all"
	None     RecursionType = "none"
	Children RecursionType = "children"
)

func normalizeUrl(resourceUrl string) (string, bool, error) {
	u, err := url.Parse(resourceUrl)
	if err != nil {
		return "", false, err
	}
	if u.Scheme == "" {
		return resourceUrl, true, nil
	}
	if u.Scheme == "file" {
		resourceUrl := strings.TrimPrefix(resourceUrl, "file://")
		if runtime.GOOS == "windows" {
			resourceUrl = filepath.FromSlash(strings.TrimPrefix(resourceUrl, "/"))
			return resourceUrl, true, nil
		}
		return resourceUrl, true, nil
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", false, fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
	return resourceUrl, false, nil
}

func loadFile(resourcePath string, value any) error {
	data, readErr := ioutil.ReadFile(resourcePath)
	if readErr != nil {
		return fmt.Errorf("failed to read file %s: %w", resourcePath, readErr)
	}

	jsonErr := json.Unmarshal(data, value)
	if jsonErr != nil {
		return fmt.Errorf("failed to parse %s: %w", resourcePath, jsonErr)
	}
	return nil
}

func loadUrl(resourceUrl string, value any) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()

	retries := 5

	return retry.Limit(ctx, retries, func(ctx context.Context, attempt int) error {
		err := tryLoadUrl(resourceUrl, value)
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

func tryLoadUrl(resourceUrl string, value any) error {
	resp, err := httpClient.Get(resourceUrl)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return fmt.Errorf("unexpected response for %s: %d", resourceUrl, resp.StatusCode)
	}

	jsonErr := json.NewDecoder(resp.Body).Decode(value)
	if jsonErr != nil {
		return fmt.Errorf("failed to parse %s: %w", resourceUrl, jsonErr)
	}
	return nil
}

func resolveURL(baseUrl string, resourceUrl string) (string, error) {
	res, resIsRelOrFile, err := normalizeUrl(resourceUrl)
	if err != nil {
		return "", err
	}

	if !resIsRelOrFile {
		return res, nil
	}

	base, baseIsFilePath, err := normalizeUrl(baseUrl)
	if err != nil {
		return "", err
	}

	if baseIsFilePath {
		baseDir := filepath.Dir(base)
		return filepath.Join(baseDir, res), nil
	}

	b, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}

	r, err := url.Parse(res)
	if err != nil {
		return "", err
	}

	resolved := b.ResolveReference(r)
	return resolved.String(), nil
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
	visited     *sync.Map
	recursion   RecursionType
	concurrency int
}

// Options for creating a crawler.
type Options struct {
	// Limit to the number of resources to fetch and visit concurrently.
	Concurrency int

	// Strategy to use when crawling linked resources.  Use None to visit
	// a single resource.  Use Children to only visit linked item/child resources.
	// Use All to visit parent and child resources.
	Recursion RecursionType
}

func (c *Crawler) apply(options *Options) {
	if options.Concurrency > 0 {
		c.concurrency = options.Concurrency
	}
	if options.Recursion != "" {
		c.recursion = options.Recursion
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
		visited: &sync.Map{},
	}
	c.apply(DefaultOptions)
	for _, opt := range options {
		c.apply(opt)
	}
	return c
}

type taskType string

const (
	resourceTask    taskType = "resource"
	collectionsTask taskType = "collections"
	featuresTask    taskType = "features"
)

type task struct {
	url  string
	kind taskType
}

// Crawl calls the visitor for each resolved resource.
//
// The resource can be a file path or a URL.  Any error returned by visitor
// will stop crawling and be returned by this function.  Context cancellation
// will also stop crawling and the context error will be returned.
func (c *Crawler) Crawl(ctx context.Context, resource string) error {
	resourceUrl, isFilepath, err := normalizeUrl(resource)
	if err != nil {
		return err
	}
	c.fileMode = isFilepath
	worker := &workgroup.Worker[*task]{
		Context: ctx,
		Limit:   c.concurrency,
		Work:    c.crawl,
	}
	worker.Add(&task{url: resourceUrl, kind: resourceTask})
	return worker.Wait()
}

func (c *Crawler) shouldVisit(resourceUrl string) bool {
	if c.recursion != All {
		return true
	}
	_, visited := c.visited.LoadOrStore(resourceUrl, true)
	return !visited
}

func (c *Crawler) normalizeAndLoad(url string, value interface{}) (string, error) {
	url, isFilepath, err := normalizeUrl(url)
	if err != nil {
		return url, err
	}

	if isFilepath {
		if !c.fileMode {
			return url, fmt.Errorf("cannot crawl file %s in non-file mode", url)
		}
		return url, loadFile(url, value)
	}
	if c.fileMode {
		return url, fmt.Errorf("cannot crawl URL %s in file mode", url)
	}
	return url, loadUrl(url, value)
}

func (c *Crawler) crawl(worker *workgroup.Worker[*task], t *task) error {
	switch t.kind {
	case resourceTask:
		return c.crawlResource(worker, t.url)
	case collectionsTask:
		return c.crawlCollections(worker, t.url)
	case featuresTask:
		return c.crawlFeatures(worker, t.url)
	default:
		return fmt.Errorf("unknown task type: %s", t.kind)
	}
}

func (c *Crawler) crawlResource(worker *workgroup.Worker[*task], resourceUrl string) error {
	if !c.shouldVisit(resourceUrl) {
		return nil
	}

	resource := Resource{}
	resourceUrl, loadErr := c.normalizeAndLoad(resourceUrl, &resource)
	if loadErr != nil {
		return loadErr
	}

	if c.recursion != None {
		// check if this looks like a STAC API root catalog that implements OGC API - Features
		if resource.Type() == Catalog && len(resource.ConformsTo()) > 1 {
			for _, link := range resource.Links() {
				if link["rel"] == "data" {
					linkURL, err := resolveURL(resourceUrl, link["href"])
					if err != nil {
						return err
					}
					worker.Add(&task{url: linkURL, kind: collectionsTask})
					return c.visitor(resourceUrl, resource)
				}
			}
		}

		for _, link := range resource.Links() {
			linkURL, err := resolveURL(resourceUrl, link["href"])
			if err != nil {
				return err
			}
			switch link["rel"] {
			case "root", "parent":
				if c.recursion != All {
					continue
				}
			case "item", "child":
				break
			default:
				continue
			}
			worker.Add(&task{url: linkURL, kind: resourceTask})
		}
	}

	return c.visitor(resourceUrl, resource)
}

func (c *Crawler) crawlCollections(worker *workgroup.Worker[*task], collectionsUrl string) error {
	response := &featureCollectionsResponse{}
	collectionsUrl, loadErr := c.normalizeAndLoad(collectionsUrl, response)
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
			if selfUrl == "" && link["rel"] == "self" && link["type"] == "application/json" {
				resolvedUrl, err := resolveURL(collectionsUrl, link["href"])
				if err != nil {
					return err
				}
				selfUrl = resolvedUrl
			}
			if itemsUrl == "" && link["rel"] == "items" && (link["type"] == "application/geo+json" || link["type"] == "application/json") {
				resolvedUrl, err := resolveURL(collectionsUrl, link["href"])
				if err != nil {
					return err
				}
				itemsUrl = resolvedUrl
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
		worker.Add(&task{url: itemsUrl, kind: featuresTask})
	}

	for _, link := range response.Links {
		if link["rel"] == "next" && link["type"] == "application/json" {
			nextUrl, err := resolveURL(collectionsUrl, link["href"])
			if err != nil {
				return err
			}
			worker.Add(&task{url: nextUrl, kind: collectionsTask})
			break
		}
	}
	return nil
}

func (c *Crawler) crawlFeatures(worker *workgroup.Worker[*task], featuresUrl string) error {
	response := &featureCollectionResponse{}
	featuresUrl, loadErr := c.normalizeAndLoad(featuresUrl, response)
	if loadErr != nil {
		return loadErr
	}
	for i, resource := range response.Features {
		if resource.Type() != Item {
			return fmt.Errorf("expected item at index %d, got %s", i, resource.Type())
		}

		var itemUrl string
		for _, link := range resource.Links() {
			if link["rel"] == "self" && link["type"] == "application/geo+json" {
				resolvedUrl, err := resolveURL(featuresUrl, link["href"])
				if err != nil {
					return err
				}
				itemUrl = resolvedUrl
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
		if link["rel"] == "next" && link["type"] == "application/json" {
			nextUrl, err := resolveURL(featuresUrl, link["href"])
			if err != nil {
				return err
			}
			worker.Add(&task{url: nextUrl, kind: featuresTask})
			break
		}
	}
	return nil
}
