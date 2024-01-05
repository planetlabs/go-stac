# STAC Utilities

Utilities for working with Spatio-Temporal Asset Catalog ([STAC](https://stacspec.org/)) resources.

[![Go Reference](https://pkg.go.dev/badge/github.com/planetlabs/go-stac.svg)](https://pkg.go.dev/github.com/planetlabs/go-stac)
![Tests](https://github.com/planetlabs/go-stac/actions/workflows/test.yml/badge.svg)

The `stac` command line utility can be used to crawl and validate STAC metadata.  In addition, the [`github.com/planetlabs/go-stac`](https://pkg.go.dev/github.com/planetlabs/go-stac) module can be used in Go projects.

## Command Line Interface

The `stac` program can be installed by downloading one of the archives from [the latest release](https://github.com/planetlabs/go-stac/releases).

Extract the archive and place the `stac` executable somewhere on your path.  See a list of available commands by running `stac` in your terminal.

Mac users can install the `stac` program with [`brew`](https://brew.sh/):

    brew install planetlabs/tap/go-stac

### CLI Usage

Run `stac help` to see a full list of commands and their arguments.  The primary `stac` commands are documented below.

#### stac validate

The `stac validate` command crawls STAC resources and validates them against the appropriate schema.

Example use:

    stac validate --entry path/to/catalog.json

The `--entry` can be a file path or URL pointing to a catalog, collection, or item.  By default, all catalogs, collections, and items linked from the entry point will be validated.  Use the `--no-recursion` option to validate a single resource without crawling to linked resources.  See `stac validate --help` for a full list of supported options.

#### stac stats

The `stac stats` command crawls STAC resources and prints out counts of resource type, versions, extensions, asset types, and conformance classes (for API endpoints).

Example use:

    stac stats --entry path/to/catalog.json

The `--entry` can be a file path or URL pointing to a catalog, collection, or item.  The stats output is a JSON object with top-level properties for catalog, collection, and item stats.

The structure of the output conforms with the schema of the [STAC Stats extension](https://github.com/stac-extensions/stats), so the results can be added to a STAC entrypoint to provide stats on child catalogs, collections, and items.  The `stac stats` command can write out a copy of the provided entrypoint with statistics added.

To write out a version of a catalog or collection that includes metadata for the STAC Stats extension, run the following:

    stac stats --entry path/to/catalog.json --output path/to/catalog-with-stats.json

## Library Use

Requires Go >= 1.18

Install the module into your project.
```
go get github.com/planetlabs/go-stac
```

See the [reference documentation](https://pkg.go.dev/github.com/planetlabs/go-stac) for example usage in a Golang project.

## Development

See the [development doc](./development.md) for information on developing and releasing the `go-stac` library.
