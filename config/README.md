# config

This directory contains YAML files that can be used to configure the
`registry-server` to use alternate storage backends.

If no configuration is specified, the SQLite configuration is used.

Configuration files can be specified using the `-c` option.

- [sqlite.yaml](sqlite.yaml) configures `registry-server` to use a SQLite
  database in its local filesystem at the location specifed in the `dbconfig`
  parameter.
- [postgres.yaml](postgres.yaml) configures `registry-server` to use a
  PostgreSQL database that can be reached using the options specified in the
  `dbconfig` parameter.
- [cloudsql-postgres.yaml](cloudsql-postgres.yaml) configures `registry-server`
  to use a CloudSQL PostgreSQL database that can be reached using the options
  specified in the `dbconfig` parameter.
