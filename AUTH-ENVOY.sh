#!/bin/sh
#
# Configure an environment to run Registry clients with a local server through a local Envoy proxy.
#

### SERVER CONFIGURATION

# These steps are needed to enable local calls to the Cloud Datastore API.
# This is required when the registry-server is run locally.

# Optionally run this to update your application-default credentials.
#gcloud auth application-default login

# This assumes that the current gcloud project is the one where data is stored.
export REGISTRY_PROJECT_IDENTIFIER=$(gcloud config list --format 'value(core.project)')

### CLIENT CONFIGURATION

# Be sure that the port setting below is correct. 9999 is the default.
export APG_REGISTRY_ADDRESS=localhost:9999
export APG_REGISTRY_AUDIENCES=http://localhost:9999

# Local calls don't use TLS.
export APG_REGISTRY_INSECURE=1

# The auth token is generated for the gcloud logged-in user.
export APG_REGISTRY_CLIENT_EMAIL=$(gcloud config list account --format "value(core.account)")
export APG_REGISTRY_TOKEN=$(gcloud auth print-access-token ${APG_REGISTRY_CLIENT_EMAIL})
unset APG_REGISTRY_API_KEY

