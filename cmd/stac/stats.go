package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/planetlabs/go-stac/crawler"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
)

type Stats struct {
	Catalogs    *ResourceStats `json:"catalogs,omitempty"`
	Collections *ResourceStats `json:"collections,omitempty"`
	Items       *ResourceStats `json:"items,omitempty"`
}

type ResourceStats struct {
	Count       uint64            `json:"count"`
	Versions    map[string]uint64 `json:"versions,omitempty"`
	Extensions  map[string]uint64 `json:"extensions,omitempty"`
	Conformance map[string]uint64 `json:"conformance,omitempty"`
	Assets      map[string]uint64 `json:"assets,omitempty"`
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
		&cli.BoolFlag{
			Name:    flagExcludeEntry,
			Usage:   "Do not count the entry itself",
			Value:   false,
			EnvVars: []string{toEnvVar(flagExcludeEntry)},
		},
		&cli.IntFlag{
			Name:    flagConcurrency,
			Usage:   "Concurrency limit",
			Value:   crawler.DefaultOptions.Concurrency,
			EnvVars: []string{toEnvVar(flagConcurrency)},
		},
		&cli.BoolFlag{
			Name:    flagNoRecursion,
			Usage:   "Visit a single resource",
			EnvVars: []string{toEnvVar(flagNoRecursion)},
		},
	},
	Action: func(ctx *cli.Context) error {
		entryPath := ctx.String(flagEntry)
		if entryPath == "" {
			return fmt.Errorf("missing --%s", flagEntry)
		}

		mutext := &sync.Mutex{}
		stats := &Stats{}

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

		skip := ctx.Bool(flagExcludeEntry)
		noRecursion := ctx.Bool(flagNoRecursion)

		visitor := func(location string, resource crawler.Resource) error {
			if skip {
				skip = false
				return nil
			}

			mutext.Lock()
			defer mutext.Unlock()

			var resourceStats *ResourceStats

			switch resource.Type() {
			case crawler.Catalog:
				if stats.Catalogs == nil {
					stats.Catalogs = &ResourceStats{}
				}
				for _, conformance := range resource.ConformsTo() {
					if stats.Catalogs.Conformance == nil {
						stats.Catalogs.Conformance = map[string]uint64{}
					}
					stats.Catalogs.Conformance[conformance] += 1
				}
				resourceStats = stats.Catalogs

			case crawler.Collection:
				if stats.Collections == nil {
					stats.Collections = &ResourceStats{}
				}
				resourceStats = stats.Collections

			case crawler.Item:
				if stats.Items == nil {
					stats.Items = &ResourceStats{}
				}
				for _, asset := range resource.Assets() {
					if stats.Items.Assets == nil {
						stats.Items.Assets = map[string]uint64{}
					}
					stats.Items.Assets[asset.Type()] += 1
				}
				resourceStats = stats.Items
			}

			for _, extension := range resource.Extensions() {
				if resourceStats.Extensions == nil {
					resourceStats.Extensions = map[string]uint64{}
				}
				resourceStats.Extensions[extension] += 1
			}

			if resourceStats.Versions == nil {
				resourceStats.Versions = map[string]uint64{}
			}
			resourceStats.Versions[resource.Version()] += 1

			resourceStats.Count += 1

			catalogs := uint64(0)
			if stats.Catalogs != nil {
				catalogs = stats.Catalogs.Count
			}

			collections := uint64(0)
			if stats.Collections != nil {
				collections = stats.Collections.Count
			}

			items := uint64(0)
			if stats.Items != nil {
				items = stats.Items.Count
			}

			_ = bar.Add(1)
			bar.Describe(fmt.Sprintf("catalogs: %d; collections: %d; items: %d", catalogs, collections, items))

			if noRecursion {
				return crawler.ErrStopRecursion
			}

			return nil
		}

		err := crawler.Crawl(entryPath, visitor, &crawler.Options{
			Concurrency: ctx.Int(flagConcurrency),
		})
		if err != nil {
			return err
		}

		_ = bar.Finish()
		return json.NewEncoder(os.Stdout).Encode(stats)
	},
}
