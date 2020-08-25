#!/bin/sh
#
# Configure an environment to run Registry clients with an evaluation server.
#

### CLIENT CONFIGURATION

# These point to the eval server.
export APG_REGISTRY_ADDRESS=registry-backend-3rqz64w4vq-uw.a.run.app:443
export APG_REGISTRY_AUDIENCES=https://registry-backend-3rqz64w4vq-uw.a.run.app

# Eval calls use TLS.
unset APG_REGISTRY_INSECURE

# The auth token is generated for the gcloud logged-in user.
export APG_REGISTRY_CLIENT_EMAIL=$(gcloud config list account --format "value(core.account)")
export APG_REGISTRY_TOKEN=$(gcloud auth print-identity-token ${APG_REGISTRY_CLIENT_EMAIL})

# Calls don't use an API key.
unset APG_REGISTRY_API_KEY