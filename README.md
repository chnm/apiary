# Religious Ecologies Data API

This repository contains a microservice written in Go that provides an API to data stored in a PostgreSQL database. It is a component of the [American Religious Ecologies](http://religiousecologies.org) project at the [Roy Rosenzweig Center for History and New Media](https://rrchnm.org).

## Compiling

Using a version of Go that supports Go modules, you should be able to run `go build` in the project root to install dependencies.

There is a Makefile in `cmd/relecapi/` that can be used for compiling and for running the service locally.

## Configuration

Set the following environment variables to configure the server:

- `RELECAPI_DBHOST` (default: `localhost`)
- `RELECAPI_DBPORT` (default: `5432`)
- `RELECAPI_DBNAME` (default: none)
- `RELECAPI_DBUSER` (default: none)
- `RELECAPI_DBPASS` (default: none)
- `RELECAPI_SSL` (default: `require`; see [pq docs](https://godoc.org/github.com/lib/pq))
- `RELECAPI_PORT` (default: `8090`)
- `RELECAPI_LOGGING` (default: `on`)

If logging is on, then access logs will be written to stdout in the Apache Common Log format. Errors and status messages will always be written to stderr.

Obviously this service requires that you be able to access the database.

## Testing

You can run the tests with `go test -v` inside the directory that contains the package for the command: `cmd/relecapi`.
