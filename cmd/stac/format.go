package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/planetlabs/go-stac/crawler"
	"github.com/urfave/cli/v2"
)

var keyOrder = []string{
	"stac_version",
	"stac_extensions",
	"type",
	"id",
	"collection",
	"title",
	"description",
	"keywords",
	"summaries",
	"properties",
	"bbox",
	"extent",
	"geometry",
	"datetime",
	"gsd",
	"platform",
	"instruments",
	"assets",
	"links",
	"href",
	"license",
}

func indexOf(list []string, item string) int {
	for i, candidate := range list {
		if candidate == item {
			return i
		}
	}
	return -1
}

type member struct {
	key   string
	value interface{}
}

type orderedMap map[string]interface{}

func (r orderedMap) members() []*member {
	members := []*member{}

	for key, value := range r {
		members = append(members, &member{key, value})
	}

	sort.Slice(members, func(i int, j int) bool {
		iKey := members[i].key
		jKey := members[j].key
		iIndex := indexOf(keyOrder, iKey)
		jIndex := indexOf(keyOrder, jKey)
		if iIndex > -1 {
			if jIndex > -1 {
				return iIndex < jIndex
			}
			return true
		}
		if jIndex > -1 {
			return false
		}
		return iKey < jKey
	})

	return members
}

func (r orderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{")

	for i, m := range r.members() {
		if i > 0 {
			buf.WriteString(",")
		}

		key, keyErr := json.Marshal(m.key)
		if keyErr != nil {
			return nil, keyErr
		}
		buf.Write(key)

		buf.WriteString(":")

		var value []byte
		var valueErr error
		if obj, ok := m.value.(map[string]interface{}); ok {
			value, valueErr = json.Marshal(orderedMap(obj))
		} else {
			value, valueErr = json.Marshal(m.value)
		}
		if valueErr != nil {
			return nil, valueErr
		}
		buf.Write(value)
	}

	buf.WriteString("}")
	return buf.Bytes(), nil
}

var formatCommand = &cli.Command{
	Name:        "format",
	Usage:       "Format STAC metadata",
	Description: "Crawls STAC resources and write formatted output.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    flagEntry,
			Usage:   "Path to STAC resource (catalog, collection, or item) to crawl",
			EnvVars: []string{toEnvVar(flagEntry)},
		},
		&cli.StringFlag{
			Name:    flagOutput,
			Usage:   "Path to a directory for writing formatted STAC metadata",
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
		baseDir := filepath.Dir(entryPath)

		outputPath := ctx.String(flagOutput)
		if outputPath == "" {
			return fmt.Errorf("missing --%s", flagOutput)
		}

		visitor := func(location string, resource crawler.Resource) error {
			relDir, err := filepath.Rel(baseDir, path.Dir(location))
			if err != nil {
				return fmt.Errorf("failed to make relative path: %w", err)
			}

			outDir := filepath.Join(outputPath, relDir)
			mkdirErr := os.MkdirAll(outDir, 0755)
			if mkdirErr != nil {
				return fmt.Errorf("failed to create output directory: %w", mkdirErr)
			}

			data, err := json.MarshalIndent(orderedMap(resource), "", "  ")
			if err != nil {
				return fmt.Errorf("failed to encode %s: %w", location, err)
			}
			outFile := filepath.Join(outDir, path.Base(location))
			if err := os.WriteFile(outFile, data, 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", outFile, err)
			}
			return nil
		}

		c := crawler.NewWithOptions(visitor, &crawler.Options{
			Concurrency: ctx.Int(flagConcurrency),
			Recursion:   crawler.RecursionType(ctx.String(flagRecursion)),
		})

		return c.Crawl(context.Background(), entryPath)
	},
}
