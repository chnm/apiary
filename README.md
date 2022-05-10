# Apiary: The RRCHNM Data API

This repository provides an API to access data stored in a PostgreSQL database. It is a component of [American Religious Ecologies](http://religiousecologies.org), [America's Public Bible](https://americaspublicbible.org), [Death by Numbers](https://deathbynumbers.org) and other projects at the [Roy Rosenzweig Center for History and New Media](https://rrchnm.org). The API is intended for use by RRCHNM projects and is not general purpose, but we provide the source code in case it is useful.

You can read more about the rationale for this piece of RRCHNM's infrastructure [on our website](https://rrchnm.org/uncategorized/rrchnms-custom-api-for-data-driven-projects/).

## Documentation 

Documentation for the various endpoints and parameters to them can be found on the [Go package website](https://pkg.go.dev/github.com/chnm/apiary). Handlers are all methods on the main server type, so you can find [any endpoint specific documentation](https://pkg.go.dev/github.com/chnm/apiary#Server) on that type.

Note that the root of the API lists out all the endpoints and sample URLs, and may be more useful than the documentation for understanding how to use the API.

## Configuration

Set the following environment variables to configure the server:

- `APIARY_DB` (default: none). A connection string to the database. An example connection string looks like this: `postgres://username:password@host:portnum/databasename`.
- `APIARY_INTERFACE` (default: `0.0.0.0`). The interface to serve the API on.
- `APIARY_PORT` (default: `8090`). The port to serve the API on.
- `APIARY_LOGGING` (default: `on`). If logging is on, then access logs will be written to stdout in the Apache Common Log format. Errors and status messages will always be written to stderr.

## Compiling or running a container

There is a `Makefile` in `cmd/dataapi/` that can be used for compiling and for running the service locally.

- `make build` will build the binary.
- `make install` will build the binary and install it under the name `apiary` to your `$GOPATH`.
- `make serve` will serve the API locally.
- `make docker-build` will create a Docker container for the API.
- `make docker-serve` will create the container and run it locally via Docker.

If you just need to run the Data API locally, it may be most convenient to just run a [Docker container](https://github.com/chnm/dataapi/pkgs/container/dataapi) served from the GitHub Container Registry. There are versions that are tagged with each of the GitHub branches that have been pushed, so that you can try the development version. You still need to set the environment variables, as below. It may be most convenient to run the Docker container with the `Makefile` in the root of this repository.

## Testing

In side the `cmd/dataapi/` directory, you can run `make test` to run integration tests with the database.
