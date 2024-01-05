# Development

See the [`go.mod`](./go.mod) file for the version of Go used for development.

Commits pushed to the default branch will be linted and tested.  You can run the linter locally before pushing a commit by installing [`golangci-lint`](https://golangci-lint.run/) (see the [`test.yml`](./.github/workflows/test.yml) workflow file for the exact version):

```shell
golangci-lint run -verbose
```

To run the tests:

```shell
go test ./...
```

## Releasing

Releases are created by pushing a tag named like `v{major}.{minor}.{patch}`.  After determining the appropriate release number, create and push a release tag from the default branch:

```shell
git tag v1.2.3
git push origin v1.2.3
```

After the [release workflow](./.github/workflows/release.yml) runs, this will create a [draft release](https://github.com/planetlabs/go-stac/releases) on GitHub.  After making any edits to the release notes, publish the draft release.
