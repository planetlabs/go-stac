package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/planetlabs/go-stac/pkg/schema"
	"github.com/planetlabs/go-stac/pkg/schema/render"
	"github.com/planetlabs/go-stac/pkg/schema/static"
)

// Schema is a top-level command for operations on schemas.
func Schema(logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Commands for working with STAC schemas",
	}
	return cmd
}

// List displays a list of schemas
func List(logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all the schemas (identified by namespace and collection) that the backend knows about",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			backend := schema.NewInMemory(schema.WithSchemas(static.Registry.GetSchemas()))
			keys, err := backend.List(ctx)
			if err != nil {
				return err
			}
			namespaceCollections := make(map[string][]string)
			for _, k := range keys {
				if _, exists := namespaceCollections[k.Namespace]; !exists {
					namespaceCollections[k.Namespace] = make([]string, 0)
				}
				namespaceCollections[k.Namespace] = append(namespaceCollections[k.Namespace], k.Collection)
			}
			for namespace, collections := range namespaceCollections {
				fmt.Printf("- %s\n", namespace)
				for _, c := range collections {
					fmt.Printf("\t%s\n", c)
				}
			}
			return nil
		},
	}
	return cmd
}

// Get displays a single schemas
func Get(logger *zap.Logger) *cobra.Command {
	const (
		nsArg  = "namespace"
		colArg = "collection"
	)
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve a schema (identified by namespace and collection) from the backend and display",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			backend := schema.NewInMemory(schema.WithSchemas(static.Registry.GetSchemas()))
			namespace := viper.GetString(nsArg)
			collection := viper.GetString(colArg)
			schemaBytes, err := backend.GetSchema(ctx, namespace, collection)
			if err != nil {
				return err
			}
			fmt.Println(string(schemaBytes))
			return nil
		},
	}
	cmd.Flags().String(nsArg, "", "parent Namespace of Collection")
	cmd.Flags().String(colArg, "", "STAC Collection name")
	return cmd
}

// Validate performs validation on a feature json object
func Validate(logger *zap.Logger) *cobra.Command {
	const (
		nsArg       = "namespace"
		colArg      = "collection"
		featPathArg = "feature-path"
	)
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Perform validation on a feature json object from stdin or a file. Prints results (w/ validation errors) on stdout",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var reader io.Reader
			if viper.GetString(featPathArg) != "" {
				file, err := os.Open(viper.GetString(featPathArg))
				if err != nil {
					return err
				}
				reader = file
			} else {
				reader = os.Stdin
			}
			feature, err := ioutil.ReadAll(reader)
			if err != nil {
				return err
			}

			namespace := viper.GetString(nsArg)
			collection := viper.GetString(colArg)
			backend := schema.NewInMemory(schema.WithSchemas(static.Registry.GetSchemas()))
			validator := schema.NewValidator(backend)
			resultErrs, err := validator.ValidateFeature(ctx, namespace, collection, feature)
			if err != nil {
				return err
			}

			if len(resultErrs) == 0 {
				fmt.Println("Validation passed!")
				os.Exit(0)
			}
			fmt.Println("Validation failed. Errors:")
			for _, e := range resultErrs {
				fmt.Printf("- %s\n", e.Error())
			}
			return nil
		},
	}
	cmd.Flags().String(nsArg, "", "parent Namespace of Collection")
	cmd.Flags().String(colArg, "", "STAC Collection name")
	cmd.Flags().String(featPathArg, "", "(optional) Path of feature json file to validate. Will use stdin if omitted")
	return cmd
}

// Render renders a Feature Collection Schema document, using the core schema
// and templating in the supplied Property schema document.
func Render(logger *zap.Logger) *cobra.Command {
	const (
		yamlPath   = "yaml"
		outputPath = "out"
	)
	cmd := &cobra.Command{
		Use:   "render",
		Short: "Render Collection schema docs from Property schema docs",
		RunE: func(cmd *cobra.Command, args []string) error {
			outpath := viper.GetString(outputPath)
			var w io.Writer
			if outpath != "" {
				var err error
				w, err = os.OpenFile(viper.GetString(outputPath), os.O_RDWR|os.O_CREATE, 0664)
				if err != nil {
					return err
				}
			} else {
				w = os.Stdout
			}

			schemaBytes, err := ioutil.ReadFile(viper.GetString(yamlPath))
			if err != nil {
				return err
			}

			jsonschema, err := render.UnmarshalYamlSchema(schemaBytes)
			if err != nil {
				return err
			}
			err = render.ExecuteCoreTemplate(jsonschema, w)
			if err != nil {
				return err
			}

			return nil
		},
	}
	cmd.Flags().String(yamlPath, "", "Path to yaml file to use as input")
	cmd.Flags().String(outputPath, "", "Path to output .json file")
	return cmd
}

// Generate is like Render, but instead of the output being a json schema file, it
// runs the json-schema document through a go template to be added to the static schema registry.
func Generate(logger *zap.Logger) *cobra.Command {
	const (
		yamlPath      = "yaml"
		nsArg         = "namespace"
		colArg        = "collection"
		staticPathArg = "static-path"
	)
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Go code to hold rendered schema docs as compiled values",
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace := viper.GetString(nsArg)
			collection := viper.GetString(colArg)
			staticPath := viper.GetString(staticPathArg)
			nsHash := base32.HexEncoding.EncodeToString(
				sha256.New().Sum([]byte(namespace)))
			collectionID := fmt.Sprintf("%s_%s", collection, nsHash[:20])
			fname := fmt.Sprintf("schema_%s.go", collectionID)
			outFH, err := os.OpenFile(path.Join(staticPath, fname), os.O_RDWR|os.O_CREATE, 0664)
			if err != nil {
				return err
			}

			schemaBytes, err := ioutil.ReadFile(viper.GetString(yamlPath))
			if err != nil {
				return err
			}

			jsonschema, err := render.UnmarshalYamlSchema(schemaBytes)
			if err != nil {
				return err
			}
			err = static.GenerateGoSchema(jsonschema, namespace, collection, collectionID, outFH)
			if err != nil {
				return err
			}

			return nil
		},
	}
	cmd.Flags().String(yamlPath, "", "Path to yaml file to use as input")
	cmd.Flags().String(nsArg, "", "STAC namespace where the Collection will live")
	cmd.Flags().String(colArg, "", "STAC Collection name")
	cmd.Flags().String(staticPathArg, "", "Absolute path to pkg/schema/static")
	return cmd
}
