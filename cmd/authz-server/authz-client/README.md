# authz-client

This directory contains a tiny command-line tool that can be used to test the
`authz-server`. It assumes that the server is running locally on the default
port and uses the `Check` method to check a request that is authorized with the
token stored in the `APG_REGISTRY_TOKEN` environment variable.
