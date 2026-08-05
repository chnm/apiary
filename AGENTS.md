# AGENTS.md

This file is the repository guide for coding agents. Human contributors should
also read `CONTRIBUTING.md`.

## Mission and constraints

Apiary is a Go HTTP API shared by multiple RRCHNM projects. Treat existing
routes, query parameters, JSON field names, null handling, and defaults as
public contracts. Prefer a focused, backward-compatible patch over broad
cleanup.

Never commit credentials, connection strings, database content, `.env` files,
or built binaries. Do not invent database schema details: derive table and
column names from the existing SQL and tests, and flag assumptions.

## Start here

Before editing:

1. Read `README.md` and `CONTRIBUTING.md`.
2. Check `git status --short --branch` and active worktrees. Other agents may be
   working in this repository; preserve their branches and uncommitted changes.
3. Inspect the relevant dataset's `handler.go`, `endpoints.go`,
   implementation files, and tests. Inspect the root assemblers when adding a
   dataset or changing service-level behavior.
4. Use the Go version/toolchain declared in `go.mod`.

The architecture is deliberately small:

- `cmd/apiary/main.go` owns process signals and graceful shutdown.
- `server.go` loads environment configuration, connects to PostgreSQL, creates
  the in-memory response cache, and assembles the server.
- `internal/datasets/<dataset>/` owns that dataset's handlers, SQL, response
  types, route registration, catalog metadata, and focused tests.
- `routes.go` assembles the dataset route registrations and service-level
  routes.
- `endpoints.go` assembles dataset catalogs into the root endpoint response.
- `middleware.go` defines middleware order and cache/error behavior.
- `db/` owns connection setup.
- `internal/httpx/` and `internal/params/` contain narrowly shared HTTP and
  parameter helpers.
- `internal/testsupport/` contains reusable test helpers.

## Implementation rules

- Format Go changes with `gofmt`.
- Use `r.Context()` for handler database calls and propagate it into helper
  functions. Avoid `context.TODO()` in new request paths.
- Parameterize every request-derived SQL value. A finite, validated allowlist
  may select a known SQL fragment such as a sort direction; never interpolate
  arbitrary input.
- Validate query and path parameters before database work. Return intentional
  4xx responses for bad client input.
- Check query, scan, row-iteration, encoding, and shutdown errors. Close rows
  promptly.
- Keep middleware ordering deliberate. Successful responses currently receive
  a one-week client cache header; errors are `no-store`; the server-side cache
  has a one-hour TTL and uses `nocache` as its refresh key.
- Preserve `GET` and `HEAD` registration unless the endpoint contract changes.
- Keep each dataset's route registration and endpoint catalog synchronized.
  Add new dataset packages to both root assemblers.
- Add focused regression tests. Prefer table-driven tests when covering
  validation boundaries or several related cases.

Do not opportunistically rewrite old handlers when a narrow change is enough.
If you notice a separate issue, report it or leave a scoped TODO only when that
is useful to maintainers.

## Validation ladder

Run checks in proportion to the change, escalating from fast to external:

```console
gofmt -w path/to/changed_file.go
go test . ./internal/...
go build ./...
go vet ./...
go test -race . ./internal/...
go mod verify
go mod tidy -diff
go test -run '^$' ./db
```

External test requirements:

- `go test ./db` needs Docker because Gnomock starts PostgreSQL.
- `go test ./cmd/apiary` needs `APIARY_DB` to reference a compatible database.
- `go test ./...` includes both external-service suites.
- `make vuln` may require network access to obtain `govulncheck`.

Do not claim an external test passed unless it was actually run. Report the
exact commands run and any environment-based omissions.

For documentation-only changes, inspect links and Markdown rendering and run
`git diff --check`; Go tests are normally unnecessary unless documentation
changes executable examples or accompanies code.

## Route-change checklist

When adding or modifying a route, verify all of the following:

- handler and response types
- registration in the dataset's `handler.go`
- discoverability and examples in the dataset's `endpoints.go`
- root route and catalog assembly when introducing a dataset package
- path/query validation and parameterized SQL
- request-context propagation
- content type, status codes, and error cache behavior
- tests for happy path, invalid input, and cancellation where applicable
- README or package comments when the behavior is user-facing

## Handoff

Keep the final handoff concise: summarize behavior changed, list validation
commands and outcomes, identify tests not run, and call out any schema or
compatibility assumptions.
