#!/bin/sh
#
# Configure an environment to run flame clients with a server published with Cloud Endpoints.
#
# The following assumes you have run `gcloud auth login` and that the current
# gcloud project is the one with your Cloud Endpoints service.
#

### CLIENT CONFIGURATION

# Calls to the Cloud Endpoints service are secure.
unset CLI_FLAME_INSECURE

# Get the service address from the gcloud tool.
export CLI_FLAME_AUDIENCES=$(gcloud run services describe flame --platform managed --format="value(status.address.url)")
export CLI_FLAME_ADDRESS=${CLI_FLAME_AUDIENCES#https://}:443

# The auth token is generated for the gcloud logged-in user.
export CLI_FLAME_CLIENT_EMAIL=$(gcloud config list account --format "value(core.account)")
export CLI_FLAME_TOKEN=$(gcloud auth print-identity-token ${CLI_FLAME_CLIENT_EMAIL})