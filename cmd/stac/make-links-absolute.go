package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

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
		baseDir := path.Dir(entryPath)

		entryUrl := ctx.String(flagUrl)
		if entryUrl == "" {
			return fmt.Errorf("missing --%s", flagUrl)
		}

		baseUrl := path.Dir(entryUrl)

		outputPath := ctx.String(flagOutput)
		if outputPath == "" {
			return fmt.Errorf("missing --%s", flagOutput)
		}

		visitor := func(location string, resource crawler.Resource) error {
			relDir, err := filepath.Rel(baseDir, path.Dir(location))
			if err != nil {
				return fmt.Errorf("failed to make relative path: %w", err)
			}

			links := resource.Links()
			for _, link := range links {
				link["href"] = makeAbsolute(link["href"], filepath.ToSlash(relDir), baseUrl)
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
			return nil
		}

		c := crawler.New(visitor, &crawler.Options{
			Concurrency: ctx.Int(flagConcurrency),
			Recursion:   crawler.RecursionType(ctx.String(flagRecursion)),
		})

		return c.Crawl(context.Background(), entryPath)
	},
}

func makeAbsolute(linkUrl string, resourceDir string, baseUrl string) string {
	if strings.HasPrefix(linkUrl, baseUrl+"/") {
		return linkUrl
	}

	return path.Join(baseUrl, path.Join(resourceDir, linkUrl))
}
