package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/planetlabs/go-stac/crawler"
	"github.com/planetlabs/go-stac/validator"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var recursionValues = []string{string(crawler.All), string(crawler.None), string(crawler.Children)}

var validateCommand = &cli.Command{
	Name:        "validate",
	Usage:       "Validate STAC metadata",
	Description: "Validates that STAC metadata is conforms with the specification.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    flagEntry,
			Usage:   "Path to STAC resource (catalog, collection, or item) to validate",
			EnvVars: []string{toEnvVar(flagEntry)},
		},
		&cli.StringSliceFlag{
			Name:    flagSchema,
			Usage:   "Substitute schema as <original>=<substitute> pairs",
			EnvVars: []string{toEnvVar(flagSchema)},
		},
		&cli.IntFlag{
			Name:    flagConcurrency,
			Usage:   "Concurrency limit",
			Value:   crawler.DefaultOptions.Concurrency,
			EnvVars: []string{toEnvVar(flagConcurrency)},
		},
		&cli.GenericFlag{
			Name:  flagRecursion,
			Usage: fmt.Sprintf("Recursion type (%s)", strings.Join(recursionValues, ", ")),
			Value: &Enum{
				Values:  recursionValues,
				Default: string(crawler.DefaultOptions.Recursion),
			},
			EnvVars: []string{toEnvVar(flagRecursion)},
		},
		&cli.GenericFlag{
			Name:  flagLogLevel,
			Usage: fmt.Sprintf("Log level (%s)", strings.Join(logLevelValues, ", ")),
			Value: &Enum{
				Values:  logLevelValues,
				Default: logrus.InfoLevel.String(),
			},
			EnvVars: []string{toEnvVar(flagLogLevel)},
		},
		&cli.GenericFlag{
			Name:  flagLogFormat,
			Usage: fmt.Sprintf("Log format (%s)", strings.Join(logFormatValues, ", ")),
			Value: &Enum{
				Values:  logFormatValues,
				Default: logFormatText,
			},
			EnvVars: []string{toEnvVar(flagLogFormat)},
		},
	},
	Action: func(ctx *cli.Context) error {
		if err := configureLogger(ctx); err != nil {
			return err
		}

		entryPath := ctx.String(flagEntry)
		if entryPath == "" {
			return fmt.Errorf("missing --%s", flagEntry)
		}

		schemaMap := map[string]string{}
		for _, pair := range ctx.StringSlice(flagSchema) {
			items := strings.Split(pair, "=")
			if len(items) != 2 {
				return fmt.Errorf("invalid --%s value %q", flagSchema, pair)
			}
			schemaMap[items[0]] = items[1]
		}

		v := validator.New(&validator.Options{
			Concurrency: ctx.Int(flagConcurrency),
			Recursion:   crawler.RecursionType(ctx.String(flagRecursion)),
			SchemaMap:   schemaMap,
		})
		err := v.Validate(context.Background(), entryPath)
		if err != nil {
			if validationErr, ok := err.(*validator.ValidationError); ok {
				return fmt.Errorf("%#v\n", validationErr)
			}
			return fmt.Errorf("validation failed: %s\n", err)
		}
		return nil
	},
}
