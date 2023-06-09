[![Go Actions Status](https://github.com/apigee/registry/workflows/Go/badge.svg)](https://github.com/apigee/registry/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/apigee/registry)](https://goreportcard.com/report/github.com/apigee/registry)
[![codecov](https://codecov.io/gh/apigee/registry/branch/main/graph/badge.svg?token=YX7LGTSYGD)](https://codecov.io/gh/apigee/registry)
[![In Solidarity](https://github.com/jpoehnelt/in-solidarity-bot/raw/main/static/badge-flat.png)](https://github.com/apps/in-solidarity)

# Registry API Core Implementation

This repository contains the core implementation of the Registry API. Please see
the [wiki](https://github.com/apigee/registry/wiki) for more information.

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
generators. The Registry API itself can be seen as a machine-readable enterprise
API catalog designed to back online directories, portals, and workflow managers.

The Registry API is formally described by the Protocol Buffer source files in
[google/cloud/apigeeregistry/v1](google/cloud/apigeeregistry/v1). It closely
follows the Google API Design Guidelines at [aip.dev](https://aip.dev) and
presents a developer experience consistent with production Google APIs. Please
tell us about your experience if you use it.

## The Registry Tool

The Registry Tool (`registry`) is a command-line tool that simplifies setup and
operation of a registry. See [cmd/registry](cmd/registry) and
[the Registry wiki](https://github.com/apigee/registry/wiki/registry) for more
information. The `registry` tool can be built from sources here or installed
with this script on Linux or Darwin:

```
curl -L https://raw.githubusercontent.com/apigee/registry/main/downloadLatest.sh | sh -
```

## This Implementation

This implementation is a [gRPC](https://grpc.io) service written in Go. It can
be run locally or deployed in a container using services including
[Google Cloud Run](https://cloud.google.com/run). It stores data using a
configurable relational interface layer that currently supports
[PostgreSQL](https://www.postgresql.org/) and [SQLite](https://www.sqlite.org/).

The Registry API service is annotated to support
[gRPC HTTP/JSON transcoding](https://aip.dev/127), which allows it to be
automatically published as a JSON REST API using a proxy. Proxies also enable
[gRPC web](https://github.com/grpc/grpc-web), which allows gRPC calls to be
directly made from browser-based applications. A configuration for the
[Envoy](https://www.envoyproxy.io/) proxy is included
([deployments/envoy/envoy.yaml](deployments/envoy/envoy.yaml)).

The Registry API protos also include configuration to support
[generated API clients (GAPICS)](https://googleapis.github.io/gapic-generators/),
which allow idiomatic API usage from a variety of languages. A Go GAPIC library
is generated as part of the build process using
[gapic-generator-go](https://github.com/googleapis/gapic-generator-go).

A command-line interface is in [cmd/registry](cmd/registry) and provides a
mixture of hand-written high-level features and automatically generated
subcommands that call individual RPC methods of the Registry API.

The entry point for the Registry API server itself is
[cmd/registry-server](cmd/registry-server). For more on running the server, see
[cmd/registry-server/README.md](cmd/registry-server/README.md).

## Build Instructions

The following tools are needed to build this software:

- Go 1.20 (recommended) or later.
- protoc, the Protocol Buffer Compiler (see
  [tools/PROTOC-VERSION.sh](/tools/PROTOC-VERSION.sh) for the currently-used
  version).
- make, git, and other elements of common unix build environments.

This repository contains a [Makefile](/Makefile) that downloads all other
dependencies and builds this software (`make all`). With dependencies
downloaded, subsequent builds can be made with `go install ./...` or
`make lite`.

## Quickstart

The easiest way to try the Registry API is to run `registry-server` locally. By
default, the server is configured to use a SQLite database.

`registry-server`

Next, in a separate terminal, configure your environment to point to this server
with the following:

`. auth/LOCAL.sh`

Now you can check your server and configuration with the `registry` tool:

`registry rpc admin get-status`

Next run a suite of tests with `make test` and see a corresponding walkthrough
of API features in [tests/demo/walkthrough.sh](tests/demo/walkthrough.sh). For
more demonstrations, see the [demos](demos) directory.

## Tests

This repository includes tests that verify `registry-server`. These server tests
focus on correctness at the API level and compliance with the API design
guidelines described at [aip.dev](https://aip.dev). Server tests are included in
runs of `make test` and `go test ./...`, and the server tests can be run by
themselves with `go test ./server/registry`. By default, server tests verify the
local code in `./server/registry`, but to allow **API conformance testing**, the
tests can be run to verify remote servers using the following options:

- With the `-remote` flag, tests are run against a remote server according to
  the configuration used by the `registry` tool. This runs the entire suite of
  tests. **WARNING**: These tests are destructive and will overwrite everything
  in the remote server.
- With the `-hosted PROJECT_ID` flag, tests are run against a remote server in a
  hosted environment within a single project that is expected to already exist.
  The server is identified and authenticated with the configuration used by the
  `registry` tool. Only the methods of the Registry service are tested (Admin
  service methods are excluded). **WARNING**: These tests are destructive and
  will overwrite everything in the specified project.

A small set of **performance benchmarks** is in
[tests/benchmark](/tests/benchmark). These tests run against remote servers
specified by the `registry` tool configuration and test a single project that is
expected to already exist. **WARNING**: These tests are destructive and will
overwrite everything in the specified project. Benchmarks can be run with the
following invocation:

```
go test ./tests/benchmark --bench=. --project_id=$PROJECTID --benchtime=${ITERATIONS}x --timeout=0
```

All of the test configurations described above are verified in this repository's
[CI tests](.github/workflows/go.yml).

## License

This software is licensed under the Apache License, Version 2.0. See
[LICENSE](LICENSE) for the full license text.

## Disclaimer

This is not an official Google product. Issues filed on GitHub are not subject
to service level agreements (SLAs) and responses should be assumed to be on an
ad-hoc volunteer basis.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING](CONTRIBUTING.md) for notes
on how to contribute to this project.
