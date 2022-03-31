# STAC Utilities

[![Go Reference](https://pkg.go.dev/badge/github.com/planetlabs/go-stac.svg)](https://pkg.go.dev/github.com/planetlabs/go-stac)

This module ([`github.com/planetlabs/go-stac`](https://github.com/planetlabs/go-stac)) provides a utilities for working with Spatio-Temporal Asset Catalog ([STAC](https://stacspec.org/)) resources.

The `stac` command line utility can be used to crawl and validate STAC metadata.  In addition, the [`github.com/planetlabs/go-stac`](https://github.com/planetlabs/go-stac) module can be used in Go projects.

## CLI

The `stac` program can be installed by downloading one of the archives from [the latest release](https://github.com/planetlabs/go-stac/releases).

Extract the archive and place the `stac` executable somewhere on your path.  See a list of available commands by running `stac` in your terminal.

For Mac users, if you get a message that `stac` can't be opened because Apple cannot check it for malicious software, you can allow access in your system preferences.  Under the Apple menu > Sytem Preferences, click Security & Privacy, then click General.  There you should see an "Allow Anyway" button.

## Library Use

Requires Go >= 1.18

Install the module into your project.
```
go get github.com/planetlabs/go-stac
```

See the [reference documentation](https://pkg.go.dev/github.com/planetlabs/go-stac) for example usage in a Golang project.
