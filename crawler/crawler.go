// Package crawler implements a STAC resource crawler.
package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/tschaub/workgroup"
)

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

func loadFile(resourcePath string) (Resource, error) {
	data, readErr := ioutil.ReadFile(resourcePath)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", resourcePath, readErr)
	}

	resource := Resource{}
	jsonErr := json.Unmarshal(data, &resource)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", resourcePath, jsonErr)
	}

	return resource, nil
}

func loadUrl(resourceUrl string) (Resource, error) {
	resp, err := http.DefaultClient.Get(resourceUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("unexpected response for %s: %d", resourceUrl, resp.StatusCode)
	}

	resource := Resource{}
	jsonErr := json.NewDecoder(resp.Body).Decode(&resource)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", resourceUrl, jsonErr)
	}
	return resource, nil
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

// Crawl calls the visitor for each resolved resource.
//
// The resource can be a file path or a URL.
func (c *Crawler) Crawl(ctx context.Context, resource string) error {
	resourceUrl, isFilepath, err := normalizeUrl(resource)
	if err != nil {
		return err
	}
	c.fileMode = isFilepath
	worker := &workgroup.Worker[string]{
		Context: ctx,
		Limit:   c.concurrency,
		Work:    c.crawl,
	}
	worker.Add(resourceUrl)
	return worker.Wait()
}

func (c *Crawler) shouldVisit(resourceUrl string) bool {
	if c.recursion != All {
		return true
	}
	_, visited := c.visited.LoadOrStore(resourceUrl, true)
	return !visited
}

func (c *Crawler) crawl(worker *workgroup.Worker[string], resourceUrl string) error {
	if !c.shouldVisit(resourceUrl) {
		return nil
	}

	resourceUrl, isFilepath, err := normalizeUrl(resourceUrl)
	if err != nil {
		return err
	}

	var resource Resource
	var loadErr error
	if isFilepath {
		if !c.fileMode {
			return fmt.Errorf("cannot crawl file %s in non-file mode", resourceUrl)
		}
		resource, loadErr = loadFile(resourceUrl)
	} else {
		if c.fileMode {
			return fmt.Errorf("cannot crawl URL %s in file mode", resourceUrl)
		}
		resource, loadErr = loadUrl(resourceUrl)
	}

	if loadErr != nil {
		return loadErr
	}

	if c.recursion != None {
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
			worker.Add(linkURL)
		}
	}

	return c.visitor(resourceUrl, resource)
}
