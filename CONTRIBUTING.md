# Contributing to Apiary

Apiary serves several RRCHNM projects from a shared PostgreSQL database.
Changes to routes, queries, and response shapes can affect multiple
applications, so favor small changes with focused tests and clear compatibility
notes.

## Repository layout

| Path | Purpose |
| --- | --- |
| `cmd/apiary/` | Executable entry point and database-backed endpoint tests |
| `db/` | PostgreSQL connection pool and container-backed integration test |
| `routes.go` | Canonical route registration |
| `endpoints.go` | Root endpoint catalog and example requests |
| `server.go` | Configuration, database connection, cache, and HTTP server lifecycle |
| `middleware.go` | Logging, CORS, client caching, compression, response cache, and recovery |
| top-level feature files | Handlers, SQL, and response types grouped by dataset |
| `.github/workflows/` | Build, test, vulnerability, image, and deployment automation |

## Development workflow

1. Branch from an up-to-date `main`.
2. Keep unrelated formatting and refactors out of the change.
3. Add or update focused tests with the implementation.
4. Run the checks appropriate to the files you changed.
5. Document new routes, parameters, configuration, or operational behavior.

Format Go files with `gofmt`. Do not hand-edit `go.sum`; use Go module commands
and verify that `go mod tidy -diff` is clean.

## Adding or changing an endpoint

An endpoint change usually requires updates in several places:

1. Implement the handler and its response types in the relevant feature file.
2. Register the route and allowed methods in `routes.go`.
3. Add the route and useful examples to the catalog in `endpoints.go`.
4. Validate query and path parameters before executing SQL.
5. Pass `r.Context()` into database calls so canceled requests stop work.
6. Use PostgreSQL parameters for values. Never build SQL by concatenating
   untrusted request data.
7. Set the correct response `Content-Type` and return a deliberate HTTP status
   for invalid input and internal errors.
8. Add focused tests for valid input, validation failures, and cancellation
   where relevant.

Preserve existing URL and JSON contracts unless a coordinated breaking change
has been agreed upon. Be especially cautious about renaming JSON fields,
changing `null` behavior, or altering default pagination and filters.

## Testing

Use the smallest useful test while iterating:

```console
go test .
go test ./path/to/package -run TestName
```

Before opening a pull request, run the hermetic CI-equivalent checks:

```console
go mod verify
go mod tidy -diff
go build ./...
go vet ./...
go test -race .
go test -run '^$' ./db
```

Additional suites have external requirements:

- `go test ./db` starts PostgreSQL through Gnomock and needs Docker.
- `go test ./cmd/apiary` needs `APIARY_DB` for the shared server created by its
  `TestMain`.
- `go test ./...` runs both of those suites.
- `make vuln` downloads and runs the pinned `govulncheck` release.

If you cannot run an external-service test, state that clearly in the pull
request rather than implying the full suite passed.

## Pull request checklist

- [ ] Route, query parameter, JSON, and cache behavior remain compatible or the
      compatibility impact is documented.
- [ ] SQL values derived from requests are parameterized.
- [ ] Database calls use the request context where possible.
- [ ] `routes.go` and `endpoints.go` agree.
- [ ] Focused tests cover the change.
- [ ] Hermetic build, vet, and test checks pass.
- [ ] User-facing or operator-facing documentation is updated.
- [ ] No credentials, database dumps, generated binaries, or local `.env` files
      are included.
