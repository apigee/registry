# registry-server

`registry-server` is the Registry API server. It supports all of the methods of
the [Registry service](/google/cloud/apigeeregistry/v1/registry_service.proto)
and a management interface defined in the
[Admin service](/google/cloud/apigeeregistry/v1/admin_service.proto).

## Usage

`registry-server` takes a single option, `-c` (or `--configuration`), which
specifies an optional configuration file, described below. When run without
options, `registry-server` runs in an evaluation mode, taking requests on port
8080 and storing data in a SQLite database at `/tmp/registry.db`. For other
default configuration settings, see
[cmd/registry-server/main.go](/cmd/registry-server/main.go).

### Configuration

Configuration for `registry-server` is loaded from a YAML file specified using
the `--configuration` (`-c`) flag. See
[config/registry_server.yaml](/config/registry-server.yaml) for details.
Configuration files can contain environment variable references. When
[config/registry_server.yaml](/config/registry-server.yaml) is specified, the
port configuration value can be set using the `PORT` environment variable.
Other useful environment variables are also defined there.

When no configuration is specified, `registry-server` runs in the evaluation
mode described in the [Usage](#usage) section above.

### Running the Registry API server with SQLite

To run `registry-server` with a SQLite backend, simply start it by running
`registry-server`. This creates and users a SQLite database in the default
location of `/tmp/registry.db`. This can be modified by specifying an alternate
configuration.

For example:

```
database:
  driver: sqlite3
  config: "data.db"
```

### Running the Registry API server with a PostgreSQL database

To run the `registry-server` with a PostgreSQL backend, ensure that you have
PostgreSQL [installed](https://www.postgresql.org/download/) and set up on your
machine. After it's ready, update the `database.driver` and `database.config`
values in your configuration.

For example:

```
database:
  driver: postgres
  config: host=localhost port=<dbport> user=<dbuser> dbname=<dbname> password=<dbpassword> sslmode=disable
```

### Running the Registry API server with a PostgreSQL database on Google Cloud SQL

The `registry-server` can also run with hosted PostgreSQL databases provided by
[Google Cloud SQL](https://cloud.google.com/sql). If you don't have an existing
PostgreSQL instance, you can follow
[these instructions](https://cloud.google.com/sql/docs/postgres/quickstart).
After your instance is ready, update the `database.driver` and
`database.config` values in your configuration.

For example:

```
database:
  driver: cloudsqlpostgres
  config: host=<project_id>:<region>:<instance_id> user=<dbuser> dbname=<dbname> password=<dbpassword> sslmode=disable
```

### Proxying a local service with Envoy

`registry-server` provides a gRPC service only. For a transcoded HTTP/JSON
interface, run the [Envoy](https://www.envoyproxy.io) proxy locally using the
configuration in the [deployments/envoy](/deployments/envoy) directory. With a
local installation of Envoy, this can be done by running the following inside
the [deployments/envoy](/deployments/envoy) directory.

`envoy -c envoy.yaml`

## Running the Registry API server in a container

The `containers` directory contains Dockerfiles and other configurations to
allow `registry-server` to be run in containers. To build a container that runs
`registry-server` standalone, use the following:

```
docker build -f containers/registry-server/Dockerfile -t registry-server .
```

Containers containing `registry-server` are also built automatically and made
available
[on GitHub](https://github.com/apigee/registry/pkgs/container/registry-server).

To run these images with docker, you'll need to provide configuration and
expose the port that the server uses inside the container (by default, this is
port 8080). Container builds read their configuration from an internal copy of
[config/registry_server.yaml](/config/registry-server.yaml), which allows you
to specify configuration by setting environment variables on the docker command
line (replacing `DBHOST`, `DBPORT`, `DBUSER`, `DBNAME`, and `DBPASSWORD` with
appropriate values for your database):

```
docker run \
  -p 8080:8080 \
  -e REGISTRY_DATABASE_DRIVER=postgres \
  -e REGISTRY_DATABASE_CONFIG="host=DBHOST port=DBPORT user=DBUSER dbname=DBNAME password=DBPASSWORD sslmode=disable" \
  ghcr.io/apigee/registry-server:latest
```

Alternately, you can use a docker
[bind mount](https://docs.docker.com/storage/bind-mounts/) to replace the
default configuration file with your own.

```
docker run \
  -p 8080:8080 \
  --mount type=bind,source="$(pwd)"/custom-config.yaml,target=/registry-config.yaml \
  ghcr.io/apigee/registry-server:latest
```

The invocation above assumes that `custom-config.yaml` is your custom server
configuration and that it configures the server port to 8080.

Be sure to verify that your PostgreSQL database server is configured to accept
remote connections (in `postgres.conf` and `pg_hba.conf`).

Note that SQLite databases are not supported in container builds. This is
because container builds exclude `CGO`, which is required by SQLite.
