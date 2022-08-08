package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/planetlabs/go-stac/crawler"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
)

const (
	statsRepoOwner = "stac-extensions"
	statsRepoName  = "stats"
)

type Stats struct {
	Catalogs    *ResourceStats `json:"stats:catalogs,omitempty"`
	Collections *ResourceStats `json:"stats:collections,omitempty"`
	Items       *ResourceStats `json:"stats:items,omitempty"`
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
			Usage:   "Path or URL to STAC resource (catalog, collection, or item) to crawl",
			EnvVars: []string{toEnvVar(flagEntry)},
		},
		&cli.StringFlag{
			Name:    flagOutput,
			Usage:   "Path to write a version of the entry resource with statistics added (if not provided, stats will be written to stdout)",
			EnvVars: []string{toEnvVar(flagOutput)},
		},
	},
	Action: func(ctx *cli.Context) error {
		rewriteWithStats := false

		outputPath := ctx.String(flagOutput)
		var entryResource crawler.Resource
		var extensionReleaseTag string

		if outputPath != "" {
			client := github.NewClient(nil)
			release, _, releaseErr := client.Repositories.GetLatestRelease(ctx.Context, statsRepoOwner, statsRepoName)
			if releaseErr != nil {
				return fmt.Errorf("failed to get latest release information for https://github.com/%s/%s: %w", statsRepoOwner, statsRepoName, releaseErr)
			}
			extensionReleaseTag = release.GetTagName()
			if extensionReleaseTag == "" {
				return fmt.Errorf("latest release for https://github.com/%s/%s has no version identifier", statsRepoOwner, statsRepoName)
			}
			rewriteWithStats = true
		}

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

		skip := rewriteWithStats

		visitor := func(resource crawler.Resource, info *crawler.ResourceInfo) error {
			if skip {
				entryResource = resource
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

			return nil
		}

		err := crawler.Crawl(entryPath, visitor)
		if err != nil {
			return err
		}

		_ = bar.Finish()

		if !rewriteWithStats {
			return json.NewEncoder(os.Stdout).Encode(stats)
		}

		extensionSchemaRoot := fmt.Sprintf("https://%s.github.io/%s/", statsRepoOwner, statsRepoName)
		extensions := []string{fmt.Sprintf("%s%s/schema.json", extensionSchemaRoot, extensionReleaseTag)}

		for _, extension := range entryResource.Extensions() {
			if strings.HasPrefix(extension, extensionSchemaRoot) {
				continue
			}
			extensions = append(extensions, extension)
		}

		entryResource["stac_extensions"] = extensions
		if stats.Catalogs != nil {
			entryResource["stats:catalogs"] = stats.Catalogs
		}
		if stats.Collections != nil {
			entryResource["stats:collections"] = stats.Collections
		}
		if stats.Items != nil {
			entryResource["stats:items"] = stats.Items
		}

		data, jsonErr := json.MarshalIndent(orderedMap(entryResource), "", "  ")
		if jsonErr != nil {
			return fmt.Errorf("failed to encode resource as JSON: %w", jsonErr)
		}
		return os.WriteFile(outputPath, data, 0644)
	},
}
