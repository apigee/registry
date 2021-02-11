# auth

This directory contains scripts that can be used to set environment variables
containing credentials and server information. These environment variables are
used by `apg`, `registry`, and other tools to configure their connections to
the `registry-server`.

- [LOCAL.sh](LOCAL.sh) configures clients to work with a locally-running
  `registry-server`.
- [CLOUDRUN.sh](CLOUDRUN.sh) configures clients to work with a
  `registry-server` deployed to Cloud Run using 'make build' and 'make deploy'
  from the top-level [Makefile](../Makefile).
- [ENVOY.sh](ENVOY.sh) configures clients to work with a locally-running
  `registry-server` that is proxied behind a local Envoy instance.
- [GKE.sh](GKE.sh) configures clients to work with a `registry-server` deployed
  to GKE. For more details about GKE deployments, please refer to
  [deployments/gke/README.md](../deployments/gke/README.md).
