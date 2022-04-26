[![Go Actions Status](https://github.com/apigee/registry/workflows/Go/badge.svg)](https://github.com/apigee/registry/actions)

# Registry API Core Implementation

This repository contains the core implementation of the Registry API.

## The Registry API

The Registry API allows teams to upload and share machine-readable descriptions
of APIs that are in use and in development. These descriptions include API
specifications in standard formats like [OpenAPI](https://www.openapis.org/),
the
[Google API Discovery Service Format](https://developers.google.com/discovery),
and the
[Protocol Buffers Language](https://developers.google.com/protocol-buffers).
These API specifications can be used by tools like linters, browsers,
documentation generators, test runners, proxies, and API client and server
generators. The Registry API itself can be seen as a machine-readable
enterprise API catalog designed to back online directories, portals, and
workflow managers.

The Registry API is formally described by the Protocol Buffer source files in
[google/cloud/apigeeregistry/v1](google/cloud/apigeeregistry/v1). It closely
follows the Google API Design Guidelines at [aip.dev](https://aip.dev) and
presents a developer experience consistent with production Google APIs. Please
tell us about your experience if you use it.

## This Implementation

This implementation is a [gRPC](https://grpc.io) service written in Go. It can
be run locally or deployed in a container using services including
[Google Cloud Run](https://cloud.google.com/run). It stores data using a
configurable relational interface layer that currently supports
[PostgreSQL](https://www.postgresql.org/) and
[SQLite](https://www.sqlite.org/).

The Registry API service is annotated to support
[gRPC HTTP/JSON transcoding](https://aip.dev/127), which allows it to be
automatically published as a JSON REST API using a proxy. Proxies also enable
[gRPC web](https://github.com/grpc/grpc-web), which allows gRPC calls to be
directly made from browser-based applications. A configuration for the
[Envoy](https://www.envoyproxy.io/) proxy is included
([deployments/envoy/envoy.yaml](deployments/envoy/envoy.yaml)) along with
scripts to build and deploy the Registry API server and a proxy in a single
container on Google Cloud Run.

The Registry API protos also include configuration to support
[generated API clients (GAPICS)](https://googleapis.github.io/gapic-generators/),
which allow idiomatic API usage from a variety of languages. A Go GAPIC library
is generated as part of the build process using
[gapic-generator-go](https://github.com/googleapis/gapic-generator-go). A
sample Go GAPIC-based client is in
[examples/gapic-client](examples/gapic-client).

Two command-line interfaces are included:

- [cmd/apg](cmd/apg), automatically generated from the API description using
  the
  [protoc-gen-go_cli](https://github.com/googleapis/gapic-generator-go/tree/master/cmd/protoc-gen-go_cli)
  tool in
  [gapic-generator-go](https://github.com/googleapis/gapic-generator-go).
- [cmd/registry](cmd/registry), a hand-written command-line tool that uses the
  Go GAPIC library to support additional API management tasks.

The entry point for the Registry API server itself is
[cmd/registry-server](cmd/registry-server).

## Build Instructions

The following tools are needed to build this software:

- Go 1.18 (recommended) or later.
- protoc, the Protocol Buffer Compiler, version 3.19.3.
- make, git, and other elements of common unix build environments.

This repository contains a Makefile that downloads all other dependencies and
builds this software (`make all`). With dependencies downloaded, subsequent
builds can be made with `go install ./...` or `make lite`. The Makefile also
includes targets that build and deploy the API on
[Google Cloud Run](https://cloud.google.com/run) (see below).

## Quickstart

The easiest way to try the Registry API is to run `registry-server` locally. By
default, the server is configured to use a SQLite database.

`registry-server`

Next, in a separate terminal, configure your environment to point to this
server with the following:

`. auth/LOCAL.sh`

Now you can check your server and configuration with the
automatically-generated `apg` client:

`apg admin get-status`

Next run a suite of tests with `make test` and see a corresponding walkthrough
of API features in [tests/demo/walkthrough.sh](tests/demo/walkthrough.sh). For
more demonstrations, see the [demos](demos) directory.

## License

This software is licensed under the Apache License, Version 2.0. See
[LICENSE](LICENSE) for the full license text.

## Disclaimer

This is not an official Google product. Issues filed on Github are not subject
to service level agreements (SLAs) and responses should be assumed to be on an
ad-hoc volunteer basis.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING](CONTRIBUTING.md) for notes
on how to contribute to this project.
