package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/planetlabs/go-stac/pkg/crawler"
	"github.com/planetlabs/go-stac/pkg/validator"
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

		v := validator.New(&validator.Options{
			Concurrency: ctx.Int(flagConcurrency),
			Recursion:   crawler.RecursionType(ctx.String(flagRecursion)),
		})
		err := v.Validate(context.Background(), entryPath)
		if err != nil {
			if validationErr, ok := err.(*validator.ValidationError); ok {
				return fmt.Errorf("validation failed: %s\n%#v\n", validationErr.Resource, validationErr)
			}
			return fmt.Errorf("validation failed:\n%#v\n", err)
		}
		return nil
	},
}
