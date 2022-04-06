package validator

import (
	"fmt"

	"github.com/planetlabs/go-stac/crawler"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// ValidationError holds details about a validation error.
type ValidationError struct {
	*jsonschema.ValidationError

	// Location is the file path or URL to the resource that failed validation.
	Location string

	// The resource being crawled.
	Resource crawler.Resource
}

func (err *ValidationError) GoString() string {
	return fmt.Sprintf("invalid %s: %s\n%s", err.Resource.Type(), err.Location, err.ValidationError.GoString())
}

func newValidationError(location string, resource crawler.Resource, err *jsonschema.ValidationError) *ValidationError {
	return &ValidationError{
		Location:        location,
		Resource:        resource,
		ValidationError: err,
	}
}
