# FLAME Reference Implementation

This directory contains a reference implementation of the FLAME (Full Lifecycle
API Management) API.

## The FLAME API

The FLAME API allows teams to upload and share machine-readable descriptions of
APIs that are in use and in development. These descriptions include API
specifications in standard formats for use by tools like linters, browsers,
documentation generators, test runners, API proxies, and client and service
generators. The API itself can be seen as a machine-readable enterprise API
catalog that can be used to build online directories, portals, and workflow
managers.

The API is formally described by the files in the [proto](proto) directory.

## This Implementation

This reference implementation is written in Go. It stores data using the
[Google Cloud Datastore API](https://cloud.google.com/datastore) and can be
deployed in a container using [Google Cloud Run](https://cloud.google.com/run).

It is implemented as a [gRPC](https://grpc.io) service and follows the Google
API Design Guidelines at [aip.dev](https://aip.dev).

The gRPC service supports [gRPC HTTP/JSON transcoding](https://aip.dev/127),
which allows gRPC service to be automatically published as a JSON REST API
using a proxy. Configuration for Envoy is included.

The gRPC service is configured to support
[generated API clients (GAPICS)](https://googleapis.github.io/gapic-generators/)
and a Go GAPIC library is generated as part of the build process using
[gapic-generator-go](https://github.com/googleapis/gapic-generator-go).

The build process also creates a command-line interface that is automatically
generated from the API description using the
[protoc-gen-go_cli](https://github.com/googleapis/gapic-generator-go/tree/master/cmd/protoc-gen-go_cli)
tool in [gapic-generator-go](https://github.com/googleapis/gapic-generator-go).

A sample application in the `apps/disco-flame` directory shows a sample use of
the API to build an online catalog of API descriptions obtained from the
[Google API Discovery Service](https://developers.google.com/discovery).

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
- **`cmd/cli`** containing the generated command-line interface.

## Enabling the Google Cloud Datastore API

The API service uses the Google Cloud Datastore API. This must be enabled for a
Google Cloud project associated with the API and appropriate credentials must
be available. One way to run the API locally is to create and download
[Service Account](https://cloud.google.com/compute/docs/access/service-accounts)
credentials, save them to a local file, and then point to this file with the
`GOOGLE_APPLICATION_CREDENTIALS` enviroment variable. (Note that when the API
server is run with Google Cloud Run, credentials are automatically provided by
the environment).

Please note: this is equivalent to running Cloud Firestore in Datastore mode.
Create your service account with the "Datastore Owner" or "Datastore User"
role.

## Running the API Locally

The INIT-LOCAL.sh script will setup your environment for the Flame API server
(`flamed`) and for the clients to call your local instance. Start the server by
running `flamed`.

The `flamed` server provides a gRPC service only. For a transcoded HTTP/JSON
interface, run the [envoy](https://www.envoyproxy.io) proxy locally using the
configuration in the [envoy](envoy) directory. With a local installation of
`envoy`, this can be done by running the following inside the [envoy](envoy)
directory.

```
sudo envoy -c envoy.yaml
```

Here `sudo` is needed because `envoy` is configured to run on port 80.

## Deploying on Google Cloud Run

This API is designed to be easily deployed on
[Google Cloud Run](https://cloud.google.com/run). To support this, the Makefile
contains targets that build a Docker image and deploy it to Google Cloud Run.
Both use the `gcloud` command, which should be authenticated and configured for
the project where the services should be run.

Requirements:

- Both targets require the [gcloud](https://cloud.google.com/sdk/gcloud)
  command, which is part of the
  [Google Cloud SDK](https://cloud.google.com/sdk).

- The `FLAME_PROJECT_IDENTIFIER` environment variable must be set to the Google
  Cloud project id where the API service is to be hosted.

If not already done, `gcloud auth login` gets user credentials for subsequent
`gcloud` operations.

If not already done, `gcloud config set project PROJECT_ID` can be used to set
your project ID, which should match \$FLAME_PROJECT_IDENTIFIER.

`make build` uses [Google Cloud Build](https://cloud.google.com/cloud-build) to
build a container containing the API server. The container is then stored in
[Google Container Registry](https://cloud.google.com/container-registry).

`make run` deploys that container on
[Google Cloud Run](https://cloud.google.com/run).

## Securing the API

When deploying to Cloud Run for the first time, you will be asked a few
questions, including this one:

`Allow unauthenticated invocations to [flame] (y/N)?`

If you answer "y", you will be able to make calls without authentication. This
is the easiest way to test the API, but it's also not too difficult to secure
the API and authenticate using a service account with a [Google Cloud IAM]()
role. To do this, choose "N" when you deploy, then go the the APIs Credentials
page in the Google Cloud Console and create a service account. When you reach
the "Service account permisions" screen, add the `Cloud Run Invoker` role to
your new service account. Then use the "Create Key" button to create and
download a private key in the JSON format.

You can use this private key with the `gcloud` command to obtain authorization
tokens to send with your API calls. To do that, you need to "activate" your
account credentials with `gcloud`. You can do that with a command like the
following, in which you substitute the email address associated with your
service account and the path to your downloaded private key:

```
gcloud auth activate-service-account flame-client@your-project-identifier.iam.gserviceaccount.com --key-file ~/Downloads/your-project-identifier-e48bd9f1c60a.json
```

Then you can use `gcloud auth print-identity-token` to get an auth token for
your service account. This requires an `--audiences` parameter that is the
address of your Cloud Run-hosted service. Here's how you can get that:

```
export AUDIENCES=$(gcloud beta run services describe flame --platform managed --format="value(status.address.url)")
```

Next, use `gcloud auth print-identity-token` to get your auth token.

```
export CLI_FLAME_TOKEN=`gcloud auth print-identity-token flame-client@your-project-identifier.iam.gserviceaccount.com --audiences="$AUDIENCES"
```

To use the CLI, it's also helpful to set the `CLI_FLAME_ADDRESS` environment
variable. Do that with the hostname associated with your service and be sure to
add port 443 since Cloud Run hosted services use SSL. Here's an example (your
service address will be different):

```
export CLI_FLAME_ADDRESS=flame-ozfrf5cr5b-uw.a.run.app:443
```

Now you can call the API with your generated CLI.

`cli flame list-products --parent projects/my-project-id --page_size 10`

Note here that `my-project-id` is arbitrary and for use within your FLAME API
calls only. It is unrelated to the Google Cloud project ID that you use for
Cloud Run and Cloud Datastore.

Auth tokens are short-lived. When your token expires, your calls will return a
message like this:
`rpc error: code = Unauthenticated desc = Unauthorized: HTTP status code 401`.
To generate a new token, rerun `gcloud auth print-identity-token` as shown
above.
