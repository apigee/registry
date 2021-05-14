# containers

This directory contains configurations and support files for building
containers containing `registry-server` and related tools.

Use `registry-server/Dockerfile` to build a container with `registry-server`
only.

Use `registry-bundle/Dockerfile` to build a container with `registry-server`,
`envoy`, and `authz-server`. Envoy is configured to support grpc-web and to
perform authorization using the `authz-server` that is included with this
project, which is also included and run in the container.
