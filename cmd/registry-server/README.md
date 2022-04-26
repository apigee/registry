# registry-server

This directory contains the main entry point for running the Registry API
server. To support running in certain hosted environments, it uses the `PORT`
environment variable to determine on which port to run.

In hosted Google environments, it receives all other configuration from
automatically-provided environment variables. In other enviroments (including
when run locally), `registry-server` requires database configuration as
described in the top-level [README](/README.md) of this repo.

## Running the Registry API server

### Configuration

Configuration for `registry-server` is loaded from a YAML file specified using
the `--configuration` (`-c`) flag.

Configuration files can contain environment variable references. See
[config/registry_server.yaml](/config/registry-server.yaml) for an example. When
that configuration file is specified, the port configuration value can be set
using the `PORT` environment variable. Other useful environment variables are
also defined there.

When no configuration is specified, `registry-server` runs on port 8080 using a
sqlite database stored in a file at `/tmp/registry.db`. For other default
configuration settings, see
[cmd/registry-server/main.go](/cmd/registry-server/main.go).

### Running the Registry API server

Run `. auth/LOCAL.sh` to configure your environment to run the Registry
API server locally and for the included clients to call your local instance.
Start the server by running `registry-server`.

### Optional: Use a PostgreSQL database on the local machine

Ensure you have PostgreSQL [installed](https://www.postgresql.org/download/)
and set up on your machine. After it's ready, update the `database.driver` and
`database.config` values in your configuration.

For example:

```
database:
  driver: postgres
  config: host=localhost port=<dbport> user=<dbuser> dbname=<dbname> password=<dbpassword> sslmode=disable
```

### Optional: Use a PostgreSQL database on Google Cloud SQL

If you don't have an existing PostgreSQL instance, you can follow
[these instructions](https://cloud.google.com/sql/docs/postgres/quickstart).
After your instance is ready, update the `database.driver` and
`database.config` values in your configuration.

For example:

```
database:
  driver: cloudsqlpostgres
  config: host=<project_id>:<region>:<instance_id> user=<dbuser> dbname=<dbname> password=<dbpassword> sslmode=disable
```

### Optional: Proxying a local service with Envoy

`registry-server` provides a gRPC service only. For a transcoded HTTP/JSON
interface, run the [Envoy](https://www.envoyproxy.io) proxy locally using the
configuration in the [deployments/envoy](/deployments/envoy) directory. With a
local installation of Envoy, this can be done by running the following inside
the [deployments/envoy](/deployments/envoy) directory.

`envoy -c envoy.yaml`

## Running the Registry API server in a container

The `containers` directory contains Dockerfiles and other configurations to
allow `registry-server` to be run in containers. Containers can be built that
run `registry-server` standalone (recommended) or in a bundled container that
includes `envoy` and a simple authorization server (mainly for running secured
instances on Cloud Run). x64 and arm64 platforms are currently supported.

To build a container that runs `registry-server` standalone, use the following:

```
docker build -f containers/registry-server/Dockerfile -t registry-server .
```

To run the image with docker, you'll need to expose the default port (8080)
that the server uses in the container. Your `docker run` invocation will look
like this:

```
docker run -p 8080:8080 registry-server:latest
```

Since the default configuration uses a SQLite database, any requests that try
to connect to the database will get an error similar to this:

```
Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub
```

This is because container builds exclude `CGO`, which is required by SQLite. To
resolve this, you can rebuild your container with a modified configuration or,
more simply, override the configuration using environment variables.

Your `docker run` invocation might look like this:

```
docker run \
  -p 8080:8080 \
  -e REGISTRY_DATABASE_DRIVER=postgres \
  -e REGISTRY_DATABASE_CONFIG="host=HOST port=PORT user=USER dbname=DATABASE password=PASSWORD sslmode=disable" \
  registry-server:latest
```

Be sure to replace `HOST` and the other database configuration parameters and
verify that your server is configured to accept remote connections (in
`postgres.conf` and `pg_hba.conf`).

## Running the Registry API server with Google Cloud Run

The Registry API server can be deployed on
[Google Cloud Run](https://cloud.google.com/run). To support this, the
[Makefile](/Makefile) contains targets that build a Docker image and that deploy
it to Google Cloud Run. Both use the `gcloud` command, which should be
authenticated and configured for the project where the services should be run.

Requirements:

- Both targets require the [gcloud](https://cloud.google.com/sdk/gcloud)
  command, which is part of the
  [Google Cloud SDK](https://cloud.google.com/sdk).

- If not already done, `gcloud auth login` gets user credentials for subsequent
  `gcloud` operations and `gcloud config set project PROJECT_ID` can be used to
  set your project ID to the one where you plan to host your servce.

- The Makefile gets your project ID from the `REGISTRY_PROJECT_IDENTIFIER`
  environment variable. This can be set automatically by running
  `. auth/CLOUDRUN.sh`.

`make build` uses [Google Cloud Build](https://cloud.google.com/cloud-build) to
build a container containing the API server. The container is stored in
[Google Container Registry](https://cloud.google.com/container-registry). This
uses the `Dockerfile` at the top level of the repo, which is a link to
[containers/registry-server/Dockerfile](containers/registry-server/Dockerfile).

`make deploy` deploys the built container on
[Google Cloud Run](https://cloud.google.com/run).

When deploying to Cloud Run for the first time, you will be asked a few
questions, including this one:

`Allow unauthenticated invocations to [registry-backend] (y/N)?`

If you answer "y", you will be able to make calls without authentication. This
is the easiest way to test the API, but it's not necessary - running
`. auth/CLOUDRUN.sh` configures your environment so that the Registry CLI
and other tools will authenticate using your user ID.

Important note: If you answer "N" to the above question, Cloud Run will require
an auth token for all requests to the server. `. auth/CLOUDRUN.sh` adds
this token to your environment, but there two possible pitfalls:

1. CORS requests will fail if your backend requires authentication
   ([details](https://groups.google.com/g/gce-discussion/c/WQUxKhZORjo)).
2. Cloud Run removes signatures from accepted JWT tokens, replacing them with
   "SIGNATURE_REMOVED_BY_GOOGLE"
   ([details](https://cloud.google.com/run/docs/troubleshooting#signature-removed)).
   If your deployment includes the Envoy proxy and
   [authz-server](https://github.com/apigee/registry-experimental/tree/main/cmd/authz-server),
   then the authz-server configuration will need to be updated to trust the JWT
   tokens that are passed through, since they've already been verified and
   further checking is impossible. You can do that by setting `trustJWTs: true`
   in
   [authz.yaml](https://github.com/apigee/registry-experimental/tree/main/cmd/authz-server/authz.yaml).

If you initially answer "N" and change your mind, you can enable
unauthenticated calls by going to the Permissions view in the Cloud Run console
and adding the "Cloud Run Invoker" role to the special username "allUsers".
(Changes take a few seconds to propagate.)

Now you can call the API with your generated CLI.

`apg admin get-status`

You can also verify your installation by running `make test`. This will run
tests against the same service that your CLI is configured to use via the
environment variables set by the `auth/*.sh` scripts.

Auth tokens are short-lived. When your token expires, your calls will return a
message like this:
`rpc error: code = Unauthenticated desc = Unauthorized: HTTP status code 401`.
To generate a new token, rerun `. auth/CLOUDRUN.sh`.

## Running the Registry API server on GKE

The [Makefile](/Makefile) contains targets that build a Docker image
(`make build`) and that deploy it to GKE (`make deploy-gke`).

Requirements:

- Ensure you have [gcloud](https://cloud.google.com/sdk/gcloud) and
  [kubectl](https://cloud.google.com/kubernetes-engine/docs/quickstart)
  installed.

- If not already done, `gcloud auth login` gets user credentials for subsequent
  `gcloud` operations and `gcloud config set project PROJECT_ID` can be used to
  set your project ID to the one where you plan to host your servce.

- The Makefile gets your project ID from the `REGISTRY_PROJECT_IDENTIFIER`
  environment variable. This can be set automatically by running
  `. auth/GKE.sh`.

For detailed steps on how to deploy to GKE, please refer to
[deployments/gke/README.md](/deployments/gke/README.md).
