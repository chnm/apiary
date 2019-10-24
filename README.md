# Religious Ecologies Data API

This repository contains a microservice written in Go that provides an API to data stored in a PostgreSQL database. It is a component of the [American Religious Ecologies](http://religiousecologies.org) project at the [Roy Rosenzweig Center for History and New Media](https://rrchnm.org).

## Compiling

Run this command to compile and install the binary in your `$GOPATH`:

```
go get github.com/religious-ecologies/relecapi/cmd/relecapi
```

## Configuration

Set the following environment variables to configure the server:

- `RELECAPI_DBHOST`
- `RELECAPI_DBPORT`
- `RELECAPI_DBNAME`
- `RELECAPI_DBUSER`
- `RELECAPI_DBPASS`
- `RELECAPI_SSL`
- `RELECAPI_PORT`

Obviously this service requires that you be able to access the database.

## Testing

You can run the tests with `go test -v` inside the directory that contains the package for the command: `cmd/relecapi`.
