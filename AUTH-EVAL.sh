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

# Eval calls don't need authentication.
unset APG_REGISTRY_TOKEN
unset APG_REGISTRY_API_KEY
