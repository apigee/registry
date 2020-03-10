# FLAME Reference Implementation

This directory contains a reference implementation of the FLAME
(Full Lifecycle API Management) API. 

## The FLAME API

The FLAME API allows teams to upload and share machine-readable descriptions of
APIs that are in use and in development. These descriptions include API specifications
in standard formats for use by tools like linters, browsers, documentation generators,
test runners, API proxies, and client and service generators. The API itself
can be seen as a machine-readable enterprise API catalog that can be used to build
online directories, portals, and workflow managers.

## This Implementation

This reference implementation is written in Go. 
It stores data using the [Google Cloud Datastore API](https://cloud.google.com/datastore) 
and can be deployed in a container using [Google Cloud Run](https://cloud.google.com/run).

It is implemented as a [gRPC](https://grpc.io) service and follows the
Google API Design Guidelines at [aip.dev](https://aip.dev).

The gRPC service supports [gRPC HTTP/JSON transcoding](https://aip.dev/127), which allows
gRPC service to be automatically published as a JSON REST API using a proxy.
Configuration for Envoy is included.

The gRPC service is configured to support
[generated API clients (GAPICS)](https://googleapis.github.io/gapic-generators/)
and a Go GAPIC library is generated as part of the build process using
[gapic-generator-go](https://github.com/googleapis/gapic-generator-go). 

The build process also creates a command-line interface that is 
automatically generated from the API description using the
[protoc-gen-go_cli](https://github.com/googleapis/gapic-generator-go/tree/master/cmd/protoc-gen-go_cli)
tool in [gapic-generator-go](https://github.com/googleapis/gapic-generator-go).

A sample application in the `apps/disco-flame` directory shows a sample use of the API
to build an online catalog of API descriptions obtained from the
[Google API Discovery Service](https://developers.google.com/discovery).

## Build Instructions

The following tools are needed to build this software:

- Go 1.13 or later.
- protoc, the Protocol Buffer Compiler, version 3.10 or later.
- make, git, and other elements of common unix build environments.

This repository contains a Makefile that downloads all other dependencies
and builds this software (`make all`). With dependencies downloaded, subsequent
builds can be made with `go install ./...`. The Makefile also includes
targets that build and deploy the API on
[Google Cloud Run](https://cloud.google.com/run) (see below).

## Generated Components

Several directories of generated code are created during the build
process (see the `COMPILE-PROTOS.sh` script for details). These include:

- **`rpc`** containing generated Protocol Buffer support code (in Go).
- **`gapic`** containing the Go GAPIC (generated API client) library.
- **`cmd/cli`** containing the generated command-line interface.

## Enabling the Google Cloud Datastore API

The API service uses the Google Cloud Datastore API. This must be enabled for a Google Cloud
project associated with the API and appropriate credentials must be available. One way
to run the API locally is to create and download
[Service Account](https://cloud.google.com/compute/docs/access/service-accounts) credentials,
save them to a local file, and then point to this file with the 
`GOOGLE_APPLICATION_CREDENTIALS` enviroment variable. (Note that when the API server
is run with Google Cloud Run, credentials are automatically provided by the environment).

## Deploying on Google Cloud Run

This API is intended to be easily deployed on [Google Cloud Run](https://cloud.google.com/run).
To support this, the Makefile contains targets that build a Docker image and deploy
it to Google Cloud Run.

Requirements:

- Both targets require the [gcloud](https://cloud.google.com/sdk/gcloud) command, 
which is part of the [Google Cloud SDK](https://cloud.google.com/sdk).
- The `FLAME_PROJECT_IDENTIFIER` environment variable must be set to the Google Cloud
project id where the API service is to be hosted.

`make build` uses [Google Cloud Build](https://cloud.google.com/cloud-build)
to build a container containing the API server. The container is then stored in
[Google Container Registry](https://cloud.google.com/container-registry).

`make run` deploys that container on [Google Cloud Run](https://cloud.google.com/run).

## Securing the API

(TODO) 

export AUDIENCES=`gcloud beta run services describe flame --platform managed --format="value(status.address.url)"`

`gcloud auth activate-service-account flame-client@your-project-identifier.iam.gserviceaccount.com --key-file ~/Downloads/your-project-identifier-e48bd9f1c60a.json`

export CLI_FLAME_TOKEN=`gcloud auth print-identity-token flame-client@your-project-identifier.iam.gserviceaccount.com --audiences="$AUDIENCES"`

export CLI_FLAME_ADDRESS=flame-ozfrf5bp4a-uw.a.run.app:443
