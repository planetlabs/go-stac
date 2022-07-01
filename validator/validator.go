// Package validator implements a STAC resource validation.
package validator

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"sync"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	"github.com/planetlabs/go-stac/crawler"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"golang.org/x/sync/singleflight"
)

func init() {
	jsonschema.SetRegexpProvider(func() jsonschema.Regexp {
		return &ecmaRegexp{}
	})
}

type ecmaRegexp struct {
	re *regexp2.Regexp
}

var _ jsonschema.Regexp = (*ecmaRegexp)(nil)

func (r *ecmaRegexp) Compile(expr string) error {
	re, err := regexp2.Compile(expr, regexp2.ECMAScript)
	if err != nil {
		return err
	}
	re.MatchTimeout = time.Second
	r.re = re
	return nil
}

func (r *ecmaRegexp) MustCompile(expr string) {
	re := regexp2.MustCompile(expr, regexp2.ECMAScript)
	re.MatchTimeout = time.Second
	r.re = re
}

func (r *ecmaRegexp) MatchString(s string) bool {
	match, _ := r.re.MatchString(s)
	return match
}

func (r *ecmaRegexp) String() string {
	return r.re.String()
}

// Validator allows validation of STAC resources.
type Validator struct {
	concurrency int
	noRecursion bool
	cache       *sync.Map
	group       *singleflight.Group
	compiler    *jsonschema.Compiler
	schemaMap   map[string]string
	logger      logr.Logger
}

// Options for the Validator.
type Options struct {
	// Limit to the number of resources to fetch and validate concurrently.
	Concurrency int

	// Set to true to validate a single resource and avoid validating all linked resources.
	NoRecursion bool

	// A lookup of substitute schema locations.  The key is the original schema location
	// and the value is the substitute location.
	SchemaMap map[string]string

	// Logger to use for logging.
	Logger *logr.Logger
}

func (v *Validator) apply(options *Options) {
	if options.Concurrency != 0 {
		v.concurrency = options.Concurrency
	}
	if options.NoRecursion {
		v.noRecursion = options.NoRecursion
	}
	if options.SchemaMap != nil {
		v.schemaMap = options.SchemaMap
	}
	if options.Logger != nil {
		v.logger = *options.Logger
	}
}

// New creates a new Validator.
func New(options ...*Options) *Validator {
	v := &Validator{
		concurrency: runtime.GOMAXPROCS(0),
		group:       &singleflight.Group{},
		cache:       &sync.Map{},
		compiler:    jsonschema.NewCompiler(),
		logger:      funcr.New(func(prefix string, args string) {}, funcr.Options{}),
	}
	for _, opt := range options {
		v.apply(opt)
	}
	return v
}

func (v *Validator) loadSchema(schemaUrl string) (*jsonschema.Schema, error) {
	log := v.logger.WithValues("schema", schemaUrl)
	if substituteUrl, ok := v.schemaMap[schemaUrl]; ok {
		log = log.WithValues("substitute", substituteUrl)
		schemaUrl = substituteUrl
	}
	if schema, ok := v.cache.Load(schemaUrl); ok {
		log.V(1).Info("schema cache hit")
		return schema.(*jsonschema.Schema), nil
	}
	schema, err, _ := v.group.Do(schemaUrl, func() (interface{}, error) {
		value, ok := v.cache.Load(schemaUrl)
		if ok {
			log.V(1).Info("schema cache hit")
			return value, nil
		}
		schema, err := v.compiler.Compile(schemaUrl)
		if err != nil {
			return nil, err
		}
		log.V(1).Info("schema cache miss")
		v.cache.Store(schemaUrl, schema)
		return schema, nil
	})

	if err != nil {
		return nil, err
	}
	return schema.(*jsonschema.Schema), nil
}

func schemaUrl(version string, resourceType crawler.ResourceType) string {
	return fmt.Sprintf("https://schemas.stacspec.org/v%s/%s-spec/json-schema/%s.json", version, resourceType, resourceType)
}

// Validate validates a STAC resource.
//
// The resource can be a path to a local file or a URL.  Validation will stop
// with the first invalid resource and the resulting ValidationError will be
// returned.  Context cancellation will also stop validation and the context
// error will be returned.
func (v *Validator) Validate(ctx context.Context, resource string) error {
	return crawler.Crawl(resource, v.validate, &crawler.Options{
		Queue: crawler.NewMemoryQueue(ctx, v.concurrency),
	})
}

func (v *Validator) validate(resourceUrl string, resource crawler.Resource) error {
	v.logger.Info("validating resource", "resource", resourceUrl)
	version := resource.Version()
	if version == "" {
		return errors.New("unexpected or missing 'stac_version' member")
	}
	resourceType := resource.Type()
	if resourceType == "" {
		return errors.New("unexpected or missing 'type' member")
	}

	coreSchema, loadErr := v.loadSchema(schemaUrl(version, resourceType))
	if loadErr != nil {
		return loadErr
	}
	coreErr := coreSchema.Validate(map[string]interface{}(resource))
	if coreErr != nil {
		if err, ok := coreErr.(*jsonschema.ValidationError); ok {
			return newValidationError(resourceUrl, resource, err)
		}
		return coreErr
	}

	for _, extension := range resource.Extensions() {
		extensionUrl, urlErr := url.Parse(extension)
		if urlErr != nil || !extensionUrl.IsAbs() {
			// this is expected for stac < 1.0.0
			v.logger.V(1).Info("invalid extension URL", "extension", extension)
			continue
		}

		extensionSchema, loadErr := v.loadSchema(extensionUrl.String())
		if loadErr != nil {
			return loadErr
		}
		extensionErr := extensionSchema.Validate(map[string]interface{}(resource))
		if extensionErr != nil {
			return extensionErr
		}
	}

	if v.noRecursion {
		return crawler.ErrStopRecursion
	}

	return nil
}
