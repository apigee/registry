# FLAME Reference Implementation

This directory contains a reference implementation of the FLAME
(Full Lifecycle API Management) API. 

The reference implementation is written in Go. 
It stores data using the [Google Cloud Datastore API]() 
and can be deployed in a container using [Google Cloud Run]().

It is implemented as a [gRPC]() service and follows the
Google API Design Guidelines that are published at [aip.dev](https://aip.dev).

The gRPC service supports gRPC JSON transcoding, which allows a JSON REST API
to be automatically published using a proxy. Configuration files for Envoy are
included.

The gRPC service is configured to support GAPIC generated API clients
and a Go GAPIC library is generated as part of the build process. 

The build process also creates a command-line interface that is 
automatically generated from the API description.

## Cloud Run

make build

make run

export AUDIENCES=`gcloud beta run services describe flame --platform managed --format="value(status.address.url)"`

`gcloud auth activate-service-account flame-client@your-project-identifier.iam.gserviceaccount.com --key-file ~/Downloads/your-project-identifier-e48bd9f1c60a.json`

export CLI_FLAME_TOKEN=`gcloud auth print-identity-token flame-client@your-project-identifier.iam.gserviceaccount.com --audiences="$AUDIENCES"`

export CLI_FLAME_ADDRESS=flame-ozfrf5bp4a-uw.a.run.app:443
