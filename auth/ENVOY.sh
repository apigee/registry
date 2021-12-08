#!/bin/sh
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
# Configure an environment to run Registry clients with a local server through a local Envoy proxy.
#

if ! [ -x "$(command -v gcloud)" ]; then
  echo 'ERROR: This script requires the gcloud command. Please install it to continue.' >&2; return
fi

### SERVER CONFIGURATION

# This assumes that the current gcloud project is the one where the Cloud PubSub API is enabled and intended for use.
export REGISTRY_PROJECT_IDENTIFIER=$(gcloud config list --format 'value(core.project)')

### CLIENT CONFIGURATION

# Be sure that the port setting below is correct. 9999 is the default.
export APG_REGISTRY_ADDRESS=localhost:9999
export APG_REGISTRY_AUDIENCES=http://localhost:9999

# Local calls don't use TLS.
export APG_REGISTRY_INSECURE=1

# The auth token is generated for the gcloud logged-in user.
export APG_REGISTRY_CLIENT_EMAIL=$(gcloud config list account --format "value(core.account)")
export APG_REGISTRY_TOKEN=$(gcloud auth print-identity-token ${APG_REGISTRY_CLIENT_EMAIL})

# Calls don't use an API key.
unset APG_REGISTRY_API_KEY
