package render

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

func TestExecuteCoreTemplateFromYaml(t *testing.T) {
	schemaPath := "schemas/PSScene4Band.yaml"
	absSchemaPath, err := testdataFixturePath(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	schemaBytes, err := ioutil.ReadFile(absSchemaPath)
	if err != nil {
		t.Fatal(err)
	}
	jsonschema, err := UnmarshalYamlSchema(schemaBytes)
	if err != nil {
		t.Fatal(err)
	}
	if jsonschema.Title != "PSScene4Band" {
		t.Fatalf("Expected Unmarshaled schema to have title='PSScene4Band'")
	}
	b := new(bytes.Buffer)
	err = ExecuteCoreTemplate(jsonschema, b)
	if err != nil {
		t.Fatalf("Got error calling ExecuteCoreTemplate() = %s", err)
	}
}

func TestInvalidGeoJson(t *testing.T) {
	docPath := "features/ps4band/invalid-no-type.json"
	schemaPath := "schemas/PSScene4Band.yaml"
	result, err := schemaTester(t, docPath, schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	if result.Valid() {
		t.Fatal("Validation should have failed -- no 'type' field on geo-json object")
	}
}

func TestCoreValidation(t *testing.T) {
	docPath := "features/ps4band/invalid-no-id.json"
	schemaPath := "schemas/PSScene4Band.yaml"
	result, err := schemaTester(t, docPath, schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]string{
		"id": "required",
	}
	errs := assertExpectations(expected, result)
	for _, e := range errs {
		t.Error(e)
	}
}

func TestRequiredPropertyValidation(t *testing.T) {
	docPath := "features/ps4band/invalid-no-observed.json"
	schemaPath := "schemas/PSScene4Band.yaml"
	result, err := schemaTester(t, docPath, schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	errs := assertExpectations(map[string]string{
		"(root)":   "number_all_of",
		"observed": "required",
	}, result)
	for _, e := range errs {
		t.Error(e)
	}
}

func TestDatetimeFormatValidation(t *testing.T) {
	docPath := "features/ps4band/invalid-wrong-acquired-format.json"
	schemaPath := "schemas/PSScene4Band.yaml"
	result, err := schemaTester(t, docPath, schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	errs := assertExpectations(map[string]string{
		"properties.observed": "format",
		"properties.acquired": "format",
		"(root)":              "number_all_of",
	}, result)
	for _, e := range errs {
		t.Error(e)
	}
}

func schemaTester(t *testing.T, docPath, schemaPath string) (*gojsonschema.Result, error) {
	docLoader, err := loaderFromFixture(docPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed loaderFromFixture for document")
	}
	absSchemaPath, err := testdataFixturePath(schemaPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find test fixture path")
	}
	schemaBytes, err := ioutil.ReadFile(absSchemaPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read properties schema file")
	}
	jsonschema, err := UnmarshalYamlSchema(schemaBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal schema yaml")
	}
	b := new(bytes.Buffer)
	err = ExecuteCoreTemplate(jsonschema, b)
	if err != nil {
		return nil, err
	}
	schemaLoader := gojsonschema.NewBytesLoader(b.Bytes())
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, errors.Wrap(err, "failed NewSchema for schema loader")
	}
	return schema.Validate(docLoader)
}

func assertExpectations(expected map[string]string, result *gojsonschema.Result) []error {
	errs := make([]error, 0)
	for _, e := range result.Errors() {
		if expected[e.Field()] != e.Type() {
			errs = append(errs, fmt.Errorf("Unexpected error of type = %s on field = %s", e.Type(), e.Field()))
		}
		delete(expected, e.Field())
	}
	for k, v := range expected {
		errs = append(errs, fmt.Errorf("Did not observe expected error type = %s on field = %s", v, k))
	}
	return errs
}

func loaderFromFixture(path string) (gojsonschema.JSONLoader, error) {
	path, err := testdataFixtureURL(path)
	if err != nil {
		return nil, err
	}
	return gojsonschema.NewReferenceLoader(path), nil
}

func testdataFixturePath(relpath string) (string, error) {
	basePath, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(basePath, "testdata/", relpath), nil
}

func testdataFixtureURL(relpath string) (string, error) {
	tdfp, err := testdataFixturePath(relpath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file://%s", tdfp), nil
}
