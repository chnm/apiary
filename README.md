# Apiary 🐝

This repository is the space for the [Roy Rosenzweig Center for History and New Media](https://rrchnm.org) data API --- ***Api***ary. The databases are our honeycombs and the data is our honey.

It is a component of [American Religious Ecologies](http://religiousecologies.org), [America's Public Bible](https://americaspublicbible.org), [Death by Numbers](https://deathbynumbers.org) and other data-driven projects at the [Roy Rosenzweig Center for History and New Media](https://rrchnm.org).

The complete documentation of the handlers and endpoints are available on the [Go package documentation](https://pkg.go.dev/github.com/chnm/dataapi).

## Compiling or running a container

Using a version of Go that supports Go modules, you should be able to run `go build` in the project root to install dependencies.

There is a `Makefile` in `cmd/dataapi/` that can be used for compiling and for running the service locally.

If you just need to run the Data API locally, it may be most convenient to just run a [Docker container](https://github.com/chnm/dataapi/pkgs/container/dataapi) served from the GitHub Container Registry. There are versions that are tagged with each of the GitHub branches that have been pushed, so that you can try the development version. You still need to set the environment variables, as below. It may be most convenient to run the Docker container with the `Makefile` in the root of this repository.

## Configuration

Set the following environment variables to configure the server:

- `APIARY_DB` (default: none). A connection string to the database.
- `APIARY_INTERFACE` (default: `0.0.0.0`)
- `APIARY_PORT` (default: `8090`). The port to serve the API on.
- `APIARY_LOGGING` (default: `on`)

If logging is on, then access logs will be written to stdout in the Apache Common Log format. Errors and status messages will always be written to stderr.

Obviously this service requires that you be able to access the database.

## Testing

You can run the tests with `go test -v` inside the directory that contains the package for the command: `cmd/dataapi`.
