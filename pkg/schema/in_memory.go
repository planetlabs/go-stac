package schema

import (
	"context"
	"errors"
)

// InMemory implements the Getter interface.
var _ Getter = &InMemory{}

// InMemory is a type of schema.Getter that has schemas hard-coded.
type InMemory struct {
	schemas map[Key][]byte
}

// List implements schema.Lister
func (im *InMemory) List(ctx context.Context, filter ...ListFilter) ([]Key, error) {
	response := make([]Key, 0)
SCHEMA:
	for k := range im.schemas {
		for _, filt := range filter {
			if filt(k) == false {
				continue SCHEMA
			}
		}
		response = append(response, k)
	}
	return response, nil
}

// GetSchema implements schema.GetSchema
func (im *InMemory) GetSchema(ctx context.Context, namespace, collection string) ([]byte, error) {
	b, ok := im.schemas[Key{namespace, collection}]
	if !ok {
		return nil, NotFound(errors.New("not found"))
	}
	return b, nil
}

// InMemoryOption is a package interface for options to change how an InMemory object should
// operate when constructed with NewInMemory.
type InMemoryOption func(*InMemory)

// WithSchemas tells the InMemory getter to use the provided schemas as its values.
func WithSchemas(schemas map[Key][]byte) InMemoryOption {
	return func(im *InMemory) {
		im.schemas = schemas
	}
}

// NewInMemory creates an in-memory schema.Getter
func NewInMemory(opts ...InMemoryOption) *InMemory {
	im := &InMemory{}
	for _, opt := range opts {
		opt(im)
	}
	if im.schemas == nil {
		im.schemas = map[Key][]byte{}
	}
	return im
}
