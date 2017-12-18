package schema

import "context"

// Getter is able to get a schema as json-schema from a backend. It may fail in different ways
// that we want callers to be able to handle differently e.g. NotFound error. Instead of returning
// sentinel concrete errors that clients check for like io.EOF, let's instead expose functions
// that can answer questions about the error, like `IsNotFound(err) bool`
type Getter interface {
	GetSchema(ctx context.Context, namespace, collection string) ([]byte, error)
}

// ListFilter is a callback that can exclude an entry by returning false
type ListFilter func(Key) bool

// Lister can list the schema keys in a backend
type Lister interface {
	List(ctx context.Context, filt ListFilter) ([]Key, error)
}

// Key uniquely identifies a Collection, to which a Schema maps 1:1
type Key struct {
	Namespace  string
	Collection string
}

// NotFound wraps the provided error to mark the error as a NotFound error
func NotFound(err error) error {
	if err == nil {
		return nil
	}
	return &notFound{
		cause: err,
	}
}

type notFound struct {
	cause error
}

func (nf *notFound) Cause() error  { return nf.cause }
func (nf *notFound) Error() string { return nf.cause.Error() }
func (nf *notFound) NotFound()     {}
