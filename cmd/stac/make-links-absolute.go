package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/planetlabs/go-stac/crawler"
	"github.com/urfave/cli/v2"
)

var absoluteLinksCommand = &cli.Command{
	Name:        "make-links-absolute",
	Usage:       "Rewrite links in STAC metadata",
	Description: "Crawls STAC resources and makes all links absolute.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    flagEntry,
			Usage:   "Path to STAC resource (catalog, collection, or item) to crawl",
			EnvVars: []string{toEnvVar(flagEntry)},
		},
		&cli.StringFlag{
			Name:    flagUrl,
			Usage:   "URL for the STAC entry resource",
			EnvVars: []string{toEnvVar(flagUrl)},
		},
		&cli.StringFlag{
			Name:    flagOutput,
			Usage:   "Path to a directory for writing updated STAC metadata",
			EnvVars: []string{toEnvVar(flagOutput)},
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
		baseDir := path.Dir(entryPath)

		entryUrl := ctx.String(flagUrl)
		if entryUrl == "" {
			return fmt.Errorf("missing --%s", flagUrl)
		}
		baseUrl, urlErr := url.Parse(entryUrl)
		if urlErr != nil {
			return fmt.Errorf("trouble parsing url %q: %w", entryUrl, urlErr)
		}
		baseUrl.Path = path.Dir(baseUrl.Path)

		outputPath := ctx.String(flagOutput)
		if outputPath == "" {
			return fmt.Errorf("missing --%s", flagOutput)
		}

		noRecursion := ctx.Bool(flagNoRecursion)

		visitor := func(location string, resource crawler.Resource) error {
			relDir, err := filepath.Rel(baseDir, path.Dir(location))
			if err != nil {
				return fmt.Errorf("failed to make relative path: %w", err)
			}

			links := resource.Links()
			for _, link := range links {
				relUrl, err := url.Parse(link["href"])
				if err != nil {
					return fmt.Errorf("failed to parse link %q: %w", link["href"], err)
				}
				absUrl := makeAbsolute(relUrl, filepath.ToSlash(relDir), baseUrl)
				link["href"] = absUrl.String()
			}
			resource["links"] = links

			outDir := filepath.Join(outputPath, relDir)
			mkdirErr := os.MkdirAll(outDir, 0755)
			if mkdirErr != nil {
				return fmt.Errorf("failed to create output directory: %w", mkdirErr)
			}

			data, err := json.MarshalIndent(resource, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to encode %s: %w", location, err)
			}
			outFile := filepath.Join(outDir, path.Base(location))
			if err := os.WriteFile(outFile, data, 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", outFile, err)
			}

			if noRecursion {
				return crawler.ErrStopRecursion
			}

			return nil
		}

		return crawler.Crawl(entryPath, visitor, &crawler.Options{
			Concurrency: ctx.Int(flagConcurrency),
		})
	},
}

func cloneUrl(u *url.URL) *url.URL {
	newUrl := *u
	return &newUrl
}

func makeAbsolute(linkUrl *url.URL, resourceDir string, baseUrl *url.URL) *url.URL {
	if linkUrl.IsAbs() {
		return linkUrl
	}

	newUrl := cloneUrl(baseUrl)
	newUrl.Path = path.Join(baseUrl.Path, path.Join(resourceDir, linkUrl.Path))

	return newUrl
}
