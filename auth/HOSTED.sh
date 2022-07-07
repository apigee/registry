#!/bin/sh
#
# Copyright 2021 Google LLC. All Rights Reserved.
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
# Configure an environment to run Apigee Registry clients with a Google-hosted service.
#
# The following assumes you have run `gcloud auth login` and that the current
# gcloud project is the one with your Apigee Registry instance.
#

if ! [ -x "$(command -v gcloud)" ]; then
  echo 'ERROR: This script requires the gcloud command. Please install it to continue.' >&2; return
fi

# Calls to the hosted service are secure.
unset APG_REGISTRY_INSECURE

# Get the service address.
export APG_REGISTRY_ADDRESS=apigeeregistry.googleapis.com:443

# The auth token is generated for the gcloud logged-in user.
export APG_REGISTRY_CLIENT_EMAIL=$(gcloud config list account --format "value(core.account)")
export APG_REGISTRY_TOKEN=$(gcloud auth print-access-token ${APG_REGISTRY_CLIENT_EMAIL})
