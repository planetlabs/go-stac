# STAC Utilities

Utilities for working with Spatio-Temporal Asset Catalog ([STAC](https://stacspec.org/)) resources.

[![Go Reference](https://pkg.go.dev/badge/github.com/planetlabs/go-stac.svg)](https://pkg.go.dev/github.com/planetlabs/go-stac)
![Tests](https://github.com/planetlabs/go-stac/actions/workflows/test.yml/badge.svg)

The `stac` command line utility can be used to crawl and validate STAC metadata.  In addition, the [`github.com/planetlabs/go-stac`](https://pkg.go.dev/github.com/planetlabs/go-stac) module can be used in Go projects.

## Command Line Interface

The `stac` program can be installed by downloading one of the archives from [the latest release](https://github.com/planetlabs/go-stac/releases).

Extract the archive and place the `stac` executable somewhere on your path.  See a list of available commands by running `stac` in your terminal.

For Mac users, if you get a message that the program can't be opened because Apple cannot check it for malicious software, you can allow access in your system preferences.  Under the **Apple** menu > **Sytem Preferences**, click **Security & Privacy**, then click **General**.  There you should see an **Allow Anyway** button.

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

The structure of the output conforms with the schema of the [STAC Stats extension](https://github.com/stac-extensions/stats), so the results can be added to a STAC entrypoint to provide stats on child catalogs, collections, and items.  When generating output to be added to a catalog or collection, you don't want to include counts for the entrypoint itself in the reported statistics.  The `--exclude-entry` flag is used to report statistics on resources linked from the entry but not on the entry itself.

To generate statistics for the STAC Stats extension, run the following:

    stac stats --entry path/to/catalog.json --exclude-entry

Paste the resulting top-level `stats:*` prefixed properties into your `catalog.json` and add the extension identifier to your catalog's `stac_extensions` property as described by the [STAC Stats extension](https://github.com/stac-extensions/stats).

## Library Use

Requires Go >= 1.18

Install the module into your project.
```
go get github.com/planetlabs/go-stac
```

See the [reference documentation](https://pkg.go.dev/github.com/planetlabs/go-stac) for example usage in a Golang project.
