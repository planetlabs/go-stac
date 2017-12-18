package schema

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// Validator is a Validator that uses a schema.Getter for validation purposes.
type Validator struct {
	getter     Getter
	validators map[Key]*gojsonschema.Schema
}

func (v *Validator) getValidator(ctx context.Context, namespace, collection string) (*gojsonschema.Schema, error) {
	key := Key{Namespace: namespace, Collection: collection}
	if schema, exists := v.validators[key]; exists {
		return schema, nil
	}

	schemaBytes, err := v.getter.GetSchema(ctx, namespace, collection)
	if err != nil {
		return nil, err
	}
	loader := gojsonschema.NewBytesLoader(schemaBytes)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return nil, err
	}

	v.validators[key] = schema
	return schema, nil
}

// ValidateFeature retrieves the json-schema document for the type
// identified by namespace and collection and uses gojsonschema to perform validation
// against the supplied feature.
// The []error return value contains simplified gojsonschema error types,
// using the value of the .Description() method of each validation error.
func (v *Validator) ValidateFeature(ctx context.Context, namespace, collection string, feature []byte) ([]error, error) {
	errs := make([]error, 0)
	vdr, err := v.getValidator(ctx, namespace, collection)
	if err != nil {
		return nil, err
	}

	f := gojsonschema.NewBytesLoader(feature)
	result, err := vdr.Validate(f)
	// this means something went wrong with the validation process itself
	if err != nil {
		return nil, err
	}

	// no news is good news
	if result.Valid() {
		return nil, err
	}

	// to avoid exposing the complexity of gojsonschema error types,
	// boil errors down to their display message (e.Description)
	for _, e := range result.Errors() {
		msg := fmt.Sprintf("field=%s, description=%s", e.Field(), e.Description())
		errs = append(errs, errors.New(msg))
	}
	return errs, nil
}

// NewValidator creates a Validator
func NewValidator(g Getter) *Validator {
	return &Validator{
		validators: make(map[Key]*gojsonschema.Schema),
		getter:     g,
	}
}
