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

# This points to the eval server.
export CLI_FLAME_ADDRESS=flame-backend-3rqz64w4vq-uw.a.run.app:443

# eval calls use TLS.
unset CLI_FLAME_INSECURE

# eval calls don't need authentication.
unset CLI_FLAME_TOKEN
unset CLI_FLAME_API_KEY
