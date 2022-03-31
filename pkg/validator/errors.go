package validator

import "github.com/santhosh-tekuri/jsonschema/v5"

// ValidationError holds details about a validation error.
type ValidationError struct {
	*jsonschema.ValidationError

	// Resource is the file path or URL to the resource that failed validation.
	Resource string
}

func newValidationError(resource string, err *jsonschema.ValidationError) *ValidationError {
	return &ValidationError{
		Resource:        resource,
		ValidationError: err,
	}
}
