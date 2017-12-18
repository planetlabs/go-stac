package static

import (
	"bytes"
	"fmt"
	"io"
	"text/template"

	"github.com/planetlabs/go-stac/pkg/schema"
	"github.com/planetlabs/go-stac/pkg/schema/render"
)

// FeatureSchema holds a json-schema document describing the
// schema for all Features in a given Collection. It is designed to
// represent a schema in generated code to be stored in the Registry at runtime
type FeatureSchema struct {
	Namespace    string
	Collection   string
	CollectionID string
	Schema       string
}

// SchemaRegistry holds the map of namespace+collection -> schemas
// it keeps an unexported member to safely hold the map of schemas.
type SchemaRegistry struct {
	schemas map[schema.Key][]byte
}

// GetSchemas returns a copy of the internally held map to be used
// by the InMemory schema.Getter (via WithSchemas())
func (ssr *SchemaRegistry) GetSchemas() map[schema.Key][]byte {
	exmap := make(map[schema.Key][]byte)
	// maps are reference types, so we need to make a copy for external consumption,
	// lest we risk accidental mutation.
	for key, value := range ssr.schemas {
		exmap[key] = value
	}
	return exmap
}

// Registry is the public face of the generated schemas.
// It provides public accessors and internal registration methods via the
// SchemaRegistry type.
// Note that Registry is not thread safe. I made an assumption that we only
// use registerSchema upon program initialization, which executes in serial.
// thereafter there should be no writes, so concurrent reads are bueno.
var Registry = &SchemaRegistry{schemas: make(map[schema.Key][]byte)}

func registerSchema(s FeatureSchema) error {
	c := schema.Key{Namespace: s.Namespace, Collection: s.Collection}
	_, exists := Registry.schemas[c]
	if exists {
		return fmt.Errorf("Duplicate schema registration for %v", c)
	}
	// I don't think we need a mutex here, assuming that this is only called
	// via init() functions in generated packages when the program initializes.
	Registry.schemas[c] = []byte(s.Schema)

	return nil
}

// GenerateGoSchema generates a golang source file with the core template
// inlined as go strings.
func GenerateGoSchema(s *render.JSONSchema, namespace, collection, collectionID string, w io.Writer) error {
	b := &bytes.Buffer{}
	err := render.ExecuteCoreTemplate(s, b)
	if err != nil {
		return err
	}

	return goConstantsTemplate.Execute(w, struct {
		Namespace    string
		Collection   string
		CollectionID string
		Schema       string
	}{
		Namespace:    namespace,
		Collection:   collection,
		CollectionID: collectionID,
		Schema:       b.String(),
	})
}

var goConstantsTemplateString = `package static 

var {{ .CollectionID }} = FeatureSchema{
  Namespace: "{{ .Namespace }}",
  Collection: "{{ .Collection }}",
  CollectionID: "{{ .CollectionID }}",
  Schema: ` + "`{{ .Schema }}`" + `,
}

func init() {
  err := registerSchema({{ .CollectionID }})
  if err != nil {
    panic(err)
  }
}
`

var goConstantsTemplate *template.Template

func init() {
	var err error

	goConstantsTemplate, err = template.New("core").Parse(goConstantsTemplateString)
	if err != nil {
		panic(err)
	}
}
