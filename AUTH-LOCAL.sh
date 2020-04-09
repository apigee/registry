#!/bin/sh
#
# Configure an environment to run flame clients with a local server.
#

### SERVER CONFIGURATION

# These steps are needed to enable local calls to the Cloud Datastore API.

# Set this to the identifier for your cloud project.
export FLAME_PROJECT_IDENTIFIER=flame-demo

# Optionally run this to update your application-default credentials.
#gcloud auth application-default login

### CLIENT CONFIGURATION

# Be sure that the port setting below is correct. 8080 is the default.
export CLI_FLAME_ADDRESS=localhost:8080

# Local calls don't use TLS.
export CLI_FLAME_INSECURE=1

# Local calls don't need authentication.
unset CLI_FLAME_TOKEN
unset CLI_FLAME_API_KEY