package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/planetlabs/go-stac/crawler"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type Stats struct {
	Catalogs    uint64            `json:"catalogs"`
	Collections uint64            `json:"collections"`
	Items       uint64            `json:"items"`
	Extensions  map[string]uint64 `json:"extensions"`
}

var statsCommand = &cli.Command{
	Name:        "stats",
	Usage:       "Generate STAC statistics",
	Description: "Crawls STAC resources and reports on statistics.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    flagEntry,
			Usage:   "Path to STAC resource (catalog, collection, or item) to stats",
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

		mutext := &sync.Mutex{}
		stats := &Stats{Extensions: map[string]uint64{}}

		visitor := func(location string, resource crawler.Resource) error {
			mutext.Lock()
			defer mutext.Unlock()

			switch resource.Type() {
			case crawler.Catalog:
				stats.Catalogs += 1
			case crawler.Collection:
				stats.Collections += 1
			case crawler.Item:
				stats.Items += 1
			}

			for _, extension := range resource.Extensions() {
				count := stats.Extensions[extension]
				stats.Extensions[extension] = count + 1
			}
			return nil
		}

		c := crawler.NewWithOptions(visitor, &crawler.Options{
			Concurrency: ctx.Int(flagConcurrency),
			Recursion:   crawler.RecursionType(ctx.String(flagRecursion)),
		})

		err := c.Crawl(context.Background(), entryPath)
		if err != nil {
			return err
		}

		return json.NewEncoder(os.Stdout).Encode(stats)
	},
}
