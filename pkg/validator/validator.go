package validator

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/planetlabs/go-stac/pkg/crawler"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"github.com/sirupsen/logrus"
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

type Validator struct {
	concurrency int
	recursion   crawler.RecursionType
	cache       *sync.Map
	group       *singleflight.Group
}

func (v *Validator) loadSchema(schemaUrl string) (*jsonschema.Schema, error) {
	log := logrus.WithField("schema", schemaUrl)
	schema, err, _ := v.group.Do(schemaUrl, func() (interface{}, error) {
		value, ok := v.cache.Load(schemaUrl)
		if ok {
			log.Debug("schema cache hit")
			return value, nil
		}
		schema, err := jsonschema.Compile(schemaUrl)
		if err != nil {
			return nil, err
		}
		log.Debug("schema cache miss")
		v.cache.Store(schemaUrl, schema)
		return schema, nil
	})

	if err != nil {
		return nil, err
	}
	return schema.(*jsonschema.Schema), nil
}

type Options struct {
	Concurrency int
	Recursion   crawler.RecursionType
}

func New(options *Options) *Validator {
	return &Validator{
		concurrency: options.Concurrency,
		recursion:   options.Recursion,
		group:       &singleflight.Group{},
		cache:       &sync.Map{},
	}
}

func schemaUrl(version string, resourceType crawler.ResourceType) string {
	return fmt.Sprintf("https://schemas.stacspec.org/v%s/%s-spec/json-schema/%s.json", version, resourceType, resourceType)
}

func (v *Validator) Validate(ctx context.Context, resourceUrl string) error {
	c := crawler.NewWithOptions(resourceUrl, v.validate, &crawler.Options{
		Concurrency: v.concurrency,
		Recursion:   v.recursion,
	})
	return c.Crawl(ctx)
}

func (v *Validator) validate(resourceUrl string, resource crawler.Resource) error {
	logrus.WithField("resource", resourceUrl).Debug("validating resource")
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
		return coreErr
	}

	for _, extensionUrl := range resource.Extensions() {
		extensionSchema, loadErr := v.loadSchema(extensionUrl)
		if loadErr != nil {
			return loadErr
		}
		extensionErr := extensionSchema.Validate(map[string]interface{}(resource))
		if extensionErr != nil {
			return extensionErr
		}
	}

	return nil
}
