#!/bin/sh
#
# Initialize an environment for running and testing the flame server locally.
#

### SERVER CONFIGURATION

# GOOGLE_APPLICATION_CREDENTIALS should identify a file containing credentials
# for a service account that has the Cloud Datastore API enabled.
export GOOGLE_APPLICATION_CREDENTIALS=$HOME/.credentials/flamedemo.json

### CLIENT CONFIGURATION

# Here we set CLI_FLAME_ADDRESS to the address of the local flame server.
export CLI_FLAME_ADDRESS=localhost:8080

# Local runs don't use TLS.
export CLI_FLAME_INSECURE=1