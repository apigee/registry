#!/bin/sh
#
# Copyright 2021 Google LLC.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

#
# Create a configuration to run Apigee Registry clients against a Google-hosted service.
#
# The following assumes you have run `gcloud auth login` and that the current
# gcloud project is the one with your Apigee Registry instance.
#

if ! [ -x "$(command -v registry)" ]; then
  echo 'ERROR: This script requires the registry command. Please install it to continue.' >&2; exit 1
fi

if ! [ -x "$(command -v gcloud)" ]; then
  echo 'ERROR: This script requires the gcloud command. Please install it to continue.' >&2; exit 1
fi

# Set the service address.
REGISTRY_ADDRESS="apigeeregistry.googleapis.com:443"
REGISTRY_INSECURE="false"

REGISTRY_PROJECT="$(gcloud config get project)"
REGISTRY_LOCATION="global"
REGISTRY_CLIENT_EMAIL="$(gcloud auth list --filter=status:ACTIVE --format="value(account)")"
REGISTRY_TOKEN_SOURCE="gcloud auth print-access-token ${REGISTRY_CLIENT_EMAIL}"

registry config configurations create hosted \
  --registry.insecure="${REGISTRY_INSECURE}" \
  --registry.address="${REGISTRY_ADDRESS}" \
  --registry.project="${REGISTRY_PROJECT}" \
  --registry.location="${REGISTRY_LOCATION}" 
  
registry config set token-source "${REGISTRY_TOKEN_SOURCE}"
