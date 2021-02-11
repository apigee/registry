#!/bin/sh
#
# Configure an environment to access the Registry server running on GKE.
#

if ! [ -x "$(command -v kubectl)" ] || ! [ -x "$(command -v gcloud)" ]; then
  echo 'ERROR: This script requires `kubectl` and `gcloud`. Please install to continue.' >&2; return
fi

### SERVER CONFIGURATION

# These steps are needed to enable local calls to the GKE service.

# Optionally run this to update your application-default credentials.
#gcloud auth application-default login

# This assumes that the current gcloud project is the one where data is stored.
export REGISTRY_PROJECT_IDENTIFIER=$(gcloud config list --format 'value(core.project)')

### CLIENT CONFIGURATION
# Below configuration assumes the server is running on the GKE cluter
# `registry-backend` under zone `us-central1-a`, and is exposed by the
# service `registry-backend`. Ensure the cluster, service and zone are
# correct.
gcloud container clusters get-credentials registry-backend --zone us-central1-a || return

ingress_ip=$(kubectl get service registry-backend -o jsonpath="{.status.loadBalancer.ingress[0].ip}")
service_port=$(kubectl get service registry-backend -o jsonpath="{.spec.ports[0].port}")
if [ -z "${ingress_ip}" ]; then
  echo "External IP not found for service 'registry-backend'. Pleasee try later."
  return
fi

export APG_REGISTRY_ADDRESS="${ingress_ip}:${service_port}"
export APG_REGISTRY_AUDIENCES="http://${APG_REGISTRY_ADDRESS}"

# The auth token is generated for the gcloud logged-in user.
export APG_REGISTRY_CLIENT_EMAIL=$(gcloud config list account --format "value(core.account)")
export APG_REGISTRY_TOKEN=$(gcloud auth print-identity-token ${APG_REGISTRY_CLIENT_EMAIL})

# Calls don't use an API key.
unset APG_REGISTRY_API_KEY
unset APG_REGISTRY_INSECURE

