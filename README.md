# A Golang STAC

This project represents a WIP implementation of [STAC (SpatioTemporal Asset Catalog)](https://github.com/radiantearth/stac-spec). We break from the WIP standard in a few places to spark discussion. It will be helpful to set up a few definitions to facilitate a description of what you'll find in this repo.

## Definitions

### SpatioTemporal Feature:
The fundamental data type in a STAC is the SpatioTemporal Feature. A Feature is a geo-json object which implements the core fields `geometry`, `id` and a temporal extent (which we have termed `observed` -- see "Observed" below). A Feature is not a feature in the traditional geospatial sense (a road, or a building) but metadata describing some particular observation that is available in a Catalog. The group chose the term Feature as a compromise between a term which is sufficiently generic ('item' and 'observation' were some competing proposals) while also aligning with the OGC terminology in a somewhat intuitive way -- a Feature is represented as a geo-json Feature, so this seems to be descriptive.

### Collection:
A collection of Feature objects with a homogenous schema.

### Feature Schema:
A JSON-Schema document, which all Features in a Collection conform to. Our version inlines the core schema, which includes geo-json structure (type, geometry, id) as well as some fields under properties (observed, duration -- see "Observed" below for more on these). We chose to inline to reduce the need for clients to make additional roundtrips to obtain the core schema. Inlining the schema also protects clients from referenced URLs going offline and helps static catalogs stay self-contained.

### Namespace:
We introduced the notion of a namespace in our API, which is an arbitrary grouping of Collections. It is structured like a filesystem path, giving us hierarchy for ACLs and a natural translation to flat catalogs and file systems in our internal storage.

### Properties Schema:
json-schema representation of the properties section of a geo-json object. Represented as yaml to be easier on the eyes. Here's an example:
https://github.com/planetlabs/go-stac/blob/master/pkg/schema/render/testdata/schemas/PSScene4Band.yaml

### Observed:
Rather than have `start` and `end` timestamps, with confusing decisions about what to do when there is a single nomimal timestamp, we went with a field called `observed` and another field called `duration` (= `end - start`) which defaults to zero (naturally leading to a single time point). This would be equivalent to Planet's `acquired` timestamp, using a new term which is hopefully more generic. We also took feedback from users of desktop GIS software that embedding important values outside of the properties section of the geo-json document would make them invisible to readers built into such software (eg ArcGIS), which only display id, geometry and properties. So we moved `observed` and `duration` into the properties section.

## What's inside

- an example Properties Schema (for the PlanetScope 4 Band product):
  - https://github.com/planetlabs/go-stac/blob/master/pkg/schema/render/testdata/schemas/PSScene4Band.yaml
- an example valid SpatioTemporal Feature document
  - https://github.com/planetlabs/go-stac/blob/master/pkg/schema/render/testdata/features/ps4band/valid.json
- invalid Feature documents used by test suite (look for files prefixed with `invalid-*`):
  - https://github.com/planetlabs/go-stac/blob/master/pkg/schema/render/testdata/features/ps4band/
- the stac cli tool


### stac CLI 

Here's what the CLI tool has to say about itself:

```
kasey@PC07ZTYJ:~/go/src/github.com/planetlabs/go-stac$ ./stac schema --help
Commands for working with STAC schemas

Usage:
  stac schema [command]

Available Commands:
  generate    Generate Go code to hold rendered schema docs as compiled values
  get         Retrieve a schema (identified by namespace and collection) from the backend and display
  list        List all the schemas (identified by namespace and collection) that the backend knows about
  render      Render Collection schema docs from Property schema docs
  validate    Perform validation on a feature json object from stdin or a file. Prints results (w/ validation errors) on stdout

Flags:
  -h, --help   help for schema

Global Flags:
      --production   use production logging presets (default is dev)

Use "stac schema [command] --help" for more information about a command.
```

The output of the `generate` subcommand looks like this: tps://github.com/planetlabs/go-stac/blob/master/pkg/schema/static/schema_PSScene4Band_5TO6OOBECLQ2UQBDC5JM.go

Note that this structure also keeps track of the namespace and collection of the generated schema.

Once a build is made using the generated go files, the `list` and `get` subcommands will include the rendered schema in the in-memory registry. It can also validate a feature document and report on validation errors. Here's an example of that:

```
kasey@PC07ZTYJ:~/go/src/github.com/planetlabs/go-stac$ ./go-stac schema validate --namespace /planet/imagery --collection PSScene4Band < /home/kasey/go/src/github.com/planetlabs/go-stac/pkg/schema/render/testdata/features/ps4band/valid.json 
Validation passed!
kasey@PC07ZTYJ:~/go/src/github.com/planetlabs/go-stac$ ./go-stac schema validate --namespace /planet/imagery --collection PSScene4Band < /home/kasey/go/src/github.com/planetlabs/go-stac/pkg/schema/render/testdata/features/ps4band/invalid-
invalid-no-id.json                  invalid-no-observed.json            invalid-no-type.json                invalid-wrong-acquired-format.json
kasey@PC07ZTYJ:~/go/src/github.com/planetlabs/go-stac$ ./go-stac schema validate --namespace /planet/imagery --collection PSScene4Band < /home/kasey/go/src/github.com/planetlabs/go-stac/pkg/schema/render/testdata/features/ps4band/invalid-wrong-acquired-format.json 
Validation failed. Errors:
- field=properties.observed, description=Does not match format 'date-time'
- field=properties.acquired, description=Does not match format 'date-time'
- field=(root), description=Must validate all the schemas (allOf)
```

## building the stac executable

- depends on make and Docker (build is run within a dockerized build container to 
  ensure builds use correct runtime dependencies).

```bash
# building stac
~/go/src/github.com/planetlabs/go-stac$ make
```
