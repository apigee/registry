#!/bin/sh
#
# Configure an environment to run flame clients with a local server.
#

### SERVER CONFIGURATION

# These steps are needed to enable local calls to the Cloud Datastore API.

# Optionally run this to update your application-default credentials.
#gcloud auth application-default login

# This assumes that the current gcloud project is the one where data is stored.
export FLAME_PROJECT_IDENTIFIER=$(gcloud config list --format 'value(core.project)')

### CLIENT CONFIGURATION

export CLI_FLAME_ADDRESS=flame-backend-yr4odda7na-uw.a.run.app:443

# Test calls use TLS.
unset CLI_FLAME_INSECURE

# Test calls don't need authentication.
unset CLI_FLAME_TOKEN
unset CLI_FLAME_API_KEY
