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

# Be sure that the port setting below is correct. 8080 is the default.
export CLI_REGISTRY_ADDRESS=localhost:8080

# Local calls don't use TLS.
export CLI_REGISTRY_INSECURE=1

# Local calls don't need authentication.
unset CLI_REGISTRY_TOKEN
unset CLI_REGISTRY_API_KEY