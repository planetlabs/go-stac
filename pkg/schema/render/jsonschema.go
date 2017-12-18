package render

// DefaultSchemaVersion is the JSON Schema version, which should be used as the default value for '$schema'.
// If extending JSON Schema with custom values use a custom URI.
// RFC draft-wright-json-schema-00, section 6
var DefaultSchemaVersion = "http://json-schema.org/draft-04/schema#"

// RootJSONSchema is the root schema, which differs from any other node only by the addition of the "$schema" attribute.
// RFC draft-wright-json-schema-00, section 4.5
type RootJSONSchema struct {
	JSONSchema
	// RFC draft-wright-json-schema-00
	Schema string `json:"$schema"` //  Section 6.1
}

// JSONSchema represents a JSON Schema object type.
type JSONSchema struct {
	// RFC draft-wright-json-schema-00
	Version string `json:"$schema,omitempty"` // section 6.1
	Ref     string `json:"$ref,omitempty"`    // section 7
	// RFC draft-wright-json-schema-validation-00, section 5
	MultipleOf           int                    `json:"multipleOf,omitempty"`           // section 5.1
	Maximum              *float64               `json:"maximum,omitempty"`              // section 5.2
	ExclusiveMaximum     bool                   `json:"exclusiveMaximum,omitempty"`     // section 5.3
	Minimum              *float64               `json:"minimum,omitempty"`              // section 5.4
	ExclusiveMinimum     bool                   `json:"exclusiveMinimum,omitempty"`     // section 5.5
	MaxLength            int                    `json:"maxLength,omitempty"`            // section 5.6
	MinLength            int                    `json:"minLength,omitempty"`            // section 5.7
	Pattern              string                 `json:"pattern,omitempty"`              // section 5.8
	AdditionalItems      *JSONSchema            `json:"additionalItems,omitempty"`      // section 5.9
	Items                *JSONSchema            `json:"items,omitempty"`                // section 5.9
	MaxItems             int                    `json:"maxItems,omitempty"`             // section 5.10
	MinItems             int                    `json:"minItems,omitempty"`             // section 5.11
	UniqueItems          bool                   `json:"uniqueItems,omitempty"`          // section 5.12
	MaxProperties        int                    `json:"maxProperties,omitempty"`        // section 5.13
	MinProperties        int                    `json:"minProperties,omitempty"`        // section 5.14
	Required             []string               `json:"required,omitempty"`             // section 5.15
	Properties           map[string]*JSONSchema `json:"properties,omitempty"`           // section 5.16
	PatternProperties    map[string]*JSONSchema `json:"patternProperties,omitempty"`    // section 5.17
	AdditionalProperties *bool                  `json:"additionalProperties,omitempty"` // section 5.18
	Dependencies         map[string]*JSONSchema `json:"dependencies,omitempty"`         // section 5.19
	Enum                 []interface{}          `json:"enum,omitempty"`                 // section 5.20
	Type                 string                 `json:"type,omitempty"`                 // section 5.21
	AllOf                []*JSONSchema          `json:"allOf,omitempty"`                // section 5.22
	AnyOf                []*JSONSchema          `json:"anyOf,omitempty"`                // section 5.23
	OneOf                []*JSONSchema          `json:"oneOf,omitempty"`                // section 5.24
	Not                  *JSONSchema            `json:"not,omitempty"`                  // section 5.25
	Definitions          Definitions            `json:"definitions,omitempty"`          // section 5.26
	// RFC draft-wright-json-schema-validation-00, section 6, 7
	Title       string      `json:"title,omitempty"`       // section 6.1
	Description string      `json:"description,omitempty"` // section 6.1
	Default     interface{} `json:"default,omitempty"`     // section 6.2
	Format      string      `json:"format,omitempty"`      // section 7
	// RFC draft-wright-json-schema-hyperschema-00, section 4
	Media          *JSONSchema `json:"media,omitempty"`          // section 4.3
	BinaryEncoding string      `json:"binaryEncoding,omitempty"` // section 4.3
	// RFC draft-wright-json-schema-01, section 9.2
	ID string `json:"$id,omitempty"` // section 9.2
}

// Definitions hold schema definitions.
// http://json-schema.org/latest/json-schema-validation.html#rfc.section.5.26
// RFC draft-wright-json-schema-validation-00, section 5.26
type Definitions map[string]*JSONSchema
