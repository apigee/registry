#!/bin/sh
#
# Configure an environment to run registry clients with a local server.
#

### SERVER CONFIGURATION

# These steps are needed to enable local calls to the Cloud Datastore API.

# Optionally run this to update your application-default credentials.
#gcloud auth application-default login

# This assumes that the current gcloud project is the one where data is stored.
export REGISTRY_PROJECT_IDENTIFIER=$(gcloud config list --format 'value(core.project)')

### CLIENT CONFIGURATION

# These point to the eval server.
export APG_REGISTRY_ADDRESS=registry-backend-3rqz64w4vq-uw.a.run.app:443
export APG_REGISTRY_AUDIENCES=https://registry-backend-3rqz64w4vq-uw.a.run.app

# eval calls use TLS.
unset APG_REGISTRY_INSECURE

# eval calls don't need authentication.
unset APG_REGISTRY_TOKEN
unset APG_REGISTRY_API_KEY
