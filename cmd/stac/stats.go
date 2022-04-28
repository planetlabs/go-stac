package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/planetlabs/go-stac/crawler"
	"github.com/schollz/progressbar/v3"
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
			Usage:   "Path to STAC resource (catalog, collection, or item) to crawl",
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
	},
	Action: func(ctx *cli.Context) error {
		entryPath := ctx.String(flagEntry)
		if entryPath == "" {
			return fmt.Errorf("missing --%s", flagEntry)
		}

		mutext := &sync.Mutex{}
		stats := &Stats{Extensions: map[string]uint64{}}

		bar := progressbar.NewOptions64(
			-1,
			progressbar.OptionSetDescription("catalogs: 0; collections: 0; items: 0"),
			progressbar.OptionSetItsString("r"),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionSetWidth(10),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionShowIts(),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
			progressbar.OptionClearOnFinish(),
		)

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

			_ = bar.Add(1)
			bar.Describe(fmt.Sprintf("catalogs: %d; collections: %d; items: %d", stats.Catalogs, stats.Collections, stats.Items))

			for _, extension := range resource.Extensions() {
				count := stats.Extensions[extension]
				stats.Extensions[extension] = count + 1
			}
			return nil
		}

		c := crawler.New(visitor, &crawler.Options{
			Concurrency: ctx.Int(flagConcurrency),
			Recursion:   crawler.RecursionType(ctx.String(flagRecursion)),
		})

		err := c.Crawl(context.Background(), entryPath)
		if err != nil {
			return err
		}

		_ = bar.Finish()
		return json.NewEncoder(os.Stdout).Encode(stats)
	},
}
