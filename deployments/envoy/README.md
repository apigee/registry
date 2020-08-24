# envoy

This directory contains configuration tools and other support files for running
Envoy locally alongside a locally-running version of `registry-server`. The
`envoy.yaml` file is also configured to use a locally-running `authz-server` as
an Envoy authorization filter.

## Instructions

1. Run `registry-server`. It will serve on its default port, 8080.

2. Run `authz-server`. It will serve on its default port, 50051.

3. Run Envoy with the `envoy.yaml` config file. You can do this with
   `envoy -c envoy.yaml` or by running the `GETENVOY.sh` script.

4. Configure your environment to send Registry requests through Envoy by
   running `source auth/ENVOY.sh` from the top of this repo.

5. Run any of the included tools or examples for accessing your local Registry,
   e.g. `apg registry get-status`.
