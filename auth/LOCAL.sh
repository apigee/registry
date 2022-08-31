#!/bin/bash
#
# Copyright 2020 Google LLC. All Rights Reserved.
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
# Create a configuration to run Registry clients against a local server.
#

### SERVER CONFIGURATION

if ! [ -x "$(command -v gcloud)" ]; then
  echo 'WARNING: The gcloud command is not installed.' >&2
  echo '  Without it, we are unable to automatically set REGISTRY_PROJECT_IDENTIFIER.' >&2
else
  # These steps are needed to enable local calls to the Cloud PubSub API.
  
  # Optionally run this to update your application-default credentials.
  # gcloud auth application-default login

  # This assumes that the current gcloud project is the one where the Cloud PubSub API is enabled and intended for use.
  export REGISTRY_PROJECT_IDENTIFIER=$(gcloud config list --format 'value(core.project)')
fi

### CLIENT CONFIGURATION

# Be sure that the port setting below is correct. 8080 is the default.
APG_REGISTRY_ADDRESS="localhost:8080"

# Local calls don't use TLS.
APG_REGISTRY_INSECURE="true"

APG_REGISTRY_PROJECT="${REGISTRY_PROJECT_IDENTIFIER}"
APG_REGISTRY_LOCATION="global"
CLIENT_EMAIL=$(gcloud config list account --format "value(core.account)")

registry config configurations create local \
  --registry.insecure="${APG_REGISTRY_INSECURE}" \
  --registry.address="${APG_REGISTRY_ADDRESS}" \
  --registry.project="${REGISTRY_PROJECT_IDENTIFIER}" \
  --registry.location="${APG_REGISTRY_LOCATION}"
