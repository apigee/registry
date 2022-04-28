# envoy

This directory contains configuration tools and other support files for running
Envoy locally alongside a locally-running version of `registry-server`.

## Instructions

1. Run `registry-server`. It will serve on its default port, 8080.

2. Run Envoy with the `envoy.yaml` config file. You can do this with
   `envoy -c envoy.yaml`.

3. Configure your environment to send Registry requests through Envoy by
   running `. auth/ENVOY.sh` from the top of this repo.

4. Run any of the included tools or examples for accessing your local Registry,
   e.g. `apg admin get-status`.

## Running Envoy with authz

The `envoy-auth.yaml` file includes an `authz` filter configuration. To run
Envoy without authz, use this configuration in step 2 above and run the
`authz-server` from the
[https://github.com/apigee/registry-experimental](apigee/registry-experimental)
repo.
