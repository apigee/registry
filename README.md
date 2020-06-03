# API Registry Reference Implementation

This directory contains a reference implementation of the API Registry API.

## The API Registry API

The API Registry API allows teams to upload and share machine-readable
descriptions of APIs that are in use and in development. These descriptions
include API specifications in standard formats like OpenAPI and Protocol
Buffers. These specifications can be used by tools like linters, browsers,
documentation generators, test runners, proxies, and API client and server
generators. The API itself can be seen as a machine-readable enterprise API
catalog that can be used to back online directories, portals, and workflow
managers.

The API is formally described by the files in the [proto](proto) directory.

## This Implementation

This reference implementation is written in Go. It stores data using the
[Google Cloud Datastore API](https://cloud.google.com/datastore) and can be
deployed in a container using [Google Cloud Run](https://cloud.google.com/run).

It is implemented as a [gRPC](https://grpc.io) service and follows the Google
API Design Guidelines at [aip.dev](https://aip.dev). The gRPC service supports
[gRPC HTTP/JSON transcoding](https://aip.dev/127), which allows it to be
automatically published as a JSON REST API using a proxy. Configuration for the
Envoy proxy is included. The service is also configured to support
[generated API clients (GAPICS)](https://googleapis.github.io/gapic-generators/)
and a Go GAPIC library is generated as part of the build process using
[gapic-generator-go](https://github.com/googleapis/gapic-generator-go). A
sample Go GAPIC-based client is in [cmd/client](cmd/client). [cmd/cli](cmd/cli)
contains a command-line interface that is automatically generated from the API
description using the
[protoc-gen-go_cli](https://github.com/googleapis/gapic-generator-go/tree/master/cmd/protoc-gen-go_cli)
tool in [gapic-generator-go](https://github.com/googleapis/gapic-generator-go).
Along with this automatically-generated CLI, the [cmd/registry](cmd/registry)
directory contains a hand-written command-line tool that supports common API
management tasks.

A sample application in [apps/disco-registry](apps/disco-registry) shows a
sample use of the API to build an online catalog of API descriptions obtained
from the
[Google API Discovery Service](https://developers.google.com/discovery).
Another sample, [apps/atlas](apps/atlas) uploads a directory of OpenAPI
specifications from any directory in the same style as
[github.com/APIs-guru/openapi-directory/APIs](https://github.com/APIs-guru/openapi-directory/tree/master/APIs).

## Build Instructions

The following tools are needed to build this software:

- Go 1.13 or later.
- protoc, the Protocol Buffer Compiler, version 3.10 or later.
- make, git, and other elements of common unix build environments.

This repository contains a Makefile that downloads all other dependencies and
builds this software (`make all`). With dependencies downloaded, subsequent
builds can be made with `go install ./...`. The Makefile also includes targets
that build and deploy the API on
[Google Cloud Run](https://cloud.google.com/run) (see below).

## Generated Components

Several directories of generated code are created during the build process (see
the `COMPILE-PROTOS.sh` script for details). These include:

- **`rpc`** containing generated Protocol Buffer support code (in Go).
- **`gapic`** containing the Go GAPIC (generated API client) library.
- **`cmd/cli`** containing a generated command-line interface.

## Enabling the Google Cloud Datastore API

The API service uses the Google Cloud Datastore API. This must be enabled for a
Google Cloud project associated with the API and appropriate credentials must
be available. One way to get credentials is to use
[Application Default Credentials](https://cloud.google.com/docs/authentication/production).
To get set up, just run `gcloud auth application-default login` and sign in.
Then make sure that your project id is set to the project that is enabled to
use the Google Cloud Datastore API. (Note that you only need to do this when
you are running the server locally. When the API server is run with Google
Cloud Run, credentials are automatically provided by the environment.)

Please note: this is equivalent to running Cloud Firestore in Datastore mode.

## Running the API Locally

Running `source AUTH-LOCAL.sh` will configure your environment for the Registry
API server (`registry-server`) and for the clients to call your local instance.
Start the server by running `registry-server`.

## Proxying a Local Service with Envoy

The `registry` server provides a gRPC service only. For a transcoded HTTP/JSON
interface, run the [envoy](https://www.envoyproxy.io) proxy locally using the
configuration in the [envoy](envoy) directory. With a local installation of
`envoy`, this can be done by running the following inside the [envoy](envoy)
directory.

```
sudo envoy -c envoy.yaml
```

Here `sudo` is needed because `envoy` is configured to run on port 80.

## Deploying with Google Cloud Run

This API is designed to be easily deployed on
[Google Cloud Run](https://cloud.google.com/run). To support this, the Makefile
contains targets that build a Docker image and deploy it to Google Cloud Run.
Both use the `gcloud` command, which should be authenticated and configured for
the project where the services should be run.

Requirements:

- Both targets require the [gcloud](https://cloud.google.com/sdk/gcloud)
  command, which is part of the
  [Google Cloud SDK](https://cloud.google.com/sdk).

- If not already done, `gcloud auth login` gets user credentials for subsequent
  `gcloud` operations and `gcloud config set project PROJECT_ID` can be used to
  set your project ID to the one where you plan to host your servce.

- The Makefile gets your project ID from the `REGISTRY_PROJECT_IDENTIFIER`
  environment variable. It can be set automatically by running
  `source AUTH-CLOUDRUN.sh`.

`make build` uses [Google Cloud Build](https://cloud.google.com/cloud-build) to
build a container containing the API server. The container is then stored in
[Google Container Registry](https://cloud.google.com/container-registry).

`make run` deploys that container on
[Google Cloud Run](https://cloud.google.com/run).

When deploying to Cloud Run for the first time, you will be asked a few
questions, including this one:

`Allow unauthenticated invocations to [registry] (y/N)?`

If you answer "y", you will be able to make calls without authentication. This
is the easiest way to test the API, but it's not necessary - running
`source AUTH-CLOUDRUN.sh` configures your environment so that the Registry CLI
and other tools will authenticate with your user ID.

Now you can call the API with your generated CLI.

`cli registry list-products --parent projects/demo --page_size 10`

Note here that `demo` is an arbitrary project ID for use within your Registry
API calls only. It is unrelated to the Google Cloud project ID that you use for
Cloud Run and Cloud Datastore.

Auth tokens are short-lived. When your token expires, your calls will return a
message like this:
`rpc error: code = Unauthenticated desc = Unauthorized: HTTP status code 401`.
To generate a new token, rerun `source AUTH-CLOUDRUN.sh`.

## Proxying a Cloud Run-based Service with Google Cloud Endpoints

For HTTP/JSON transcoding and other API management features, see the
[endpoints](endpoints) directory for instructions and scripts for configuring a
Google Cloud Endpoints frontend for your Cloud Run-based service.
