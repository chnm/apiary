# Apiary: the RRCHNM Data API

Apiary is the shared PostgreSQL-backed data API for projects at the
[Roy Rosenzweig Center for History and New Media](https://rrchnm.org),
including:

- [American Religious Ecologies](https://religiousecologies.org)
- [America's Public Bible](https://americaspublicbible.org)
- [Death by Numbers](https://deathbynumbers.org)

The service is intended primarily for RRCHNM projects rather than as a
general-purpose public API. Its source is available as a reference and for
reuse. For background on why RRCHNM maintains a shared API, see
[RRCHNM's custom API for data-driven projects](https://rrchnm.org/uncategorized/rrchnms-custom-api-for-data-driven-projects/).

## Using the API

The root URL returns a JSON catalog of every registered endpoint together with
sample requests. It is the best starting point for exploring a running
instance:

```console
curl https://data.chnm.org/
```

Apiary is deployed as a service rather than consumed as an importable Go
library. Handler implementations, response types, route registration, and
catalog metadata are grouped by dataset under `internal/datasets`; the running
service's root catalog is the authoritative endpoint reference.

All routes accept `GET` and `HEAD`. Responses are CORS-enabled and compressed
when the client supports it. Successful responses may be cached; append the
`nocache` query parameter when you need the server to refresh a cached result:

```console
curl "http://localhost:8090/bom/parishes?nocache"
```

## Requirements

- Go 1.25 or newer; `go.mod` selects Go 1.26.5 as the preferred toolchain
- PostgreSQL and a connection string for running the service
- Docker for container builds and the database integration test in `db/`

## Configuration

Apiary reads configuration from environment variables:

| Variable | Default | Description |
| --- | --- | --- |
| `APIARY_DB` | none | PostgreSQL connection string, such as `postgres://user:password@host:5432/database` |
| `APIARY_INTERFACE` | `0.0.0.0` | Interface on which the HTTP server listens |
| `APIARY_PORT` | `8090` | HTTP port |
| `APIARY_LOGGING` | `on` | Set to `off` to disable access logs; errors and status messages still go to stderr |

Keep local credentials in an ignored `.env` file or your shell environment.
The application does not load `.env` files by itself.

## Run locally

Export a database connection and start the server:

```console
export APIARY_DB="postgres://user:password@localhost:5432/apiary"
make serve
```

Then inspect the endpoint catalog:

```console
curl http://localhost:8090/
```

Useful Make targets include:

| Command | Purpose |
| --- | --- |
| `make build` | Build the local binary at `cmd/apiary/apiary` |
| `make serve` | Build and serve on `localhost` |
| `make install` | Install `apiary` into `$GOBIN`, or `$GOPATH/bin` when `$GOBIN` is unset |
| `make test` | Run all Go tests; some require database or Docker access |
| `make bench` | Run benchmarks with access logging disabled |
| `make vuln` | Run `govulncheck` |
| `make docker-build` | Build a local `apiary:test` image |
| `make docker-serve` | Build and run that image with local environment variables |
| `make serve-ghcr` | Run the GHCR image tagged for the current Git branch; the branch name must be a valid container tag |

To use Docker directly:

```console
docker build --tag apiary:test .
docker run --rm --publish 8090:8090 \
  --env APIARY_DB \
  --name apiary-test \
  apiary:test
```

## Test and validate

The root and internal package tests are hermetic and form the fast local
feedback loop:

```console
go test . ./internal/...
go vet ./...
go build ./...
```

The CI check additionally verifies module files, runs the root tests with the
race detector, compiles the database tests without starting Docker, and runs
security and workflow linters:

```console
go mod verify
go mod tidy -diff
go test -race . ./internal/...
go test -run '^$' ./db
gosec -exclude-dir=.claude ./...
actionlint -color
```

The test suites have different runtime requirements:

- `go test -race . ./internal/...` runs the deterministic service and dataset
  unit tests.
- `go test ./cmd/apiary` runs integration tests against the PostgreSQL
  database configured by `APIARY_DB`. These tests validate response contracts
  and stable invariants rather than snapshotting mutable database records.
- `go test ./db` runs the Gnomock connection test and requires a Docker daemon.
- `make test` or `go test ./...` runs all suites and therefore requires both
  PostgreSQL and Docker.

See [CONTRIBUTING.md](CONTRIBUTING.md) for repository layout, change guidance,
and the pull request checklist.

## License

See [LICENSE.txt](LICENSE.txt).
