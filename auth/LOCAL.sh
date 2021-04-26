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
# Configure an environment to run Registry clients with a local server.
#

### SERVER CONFIGURATION

if ! [ -x "$(command -v gcloud)" ]; then
  echo 'WARNING: The gcloud command is not installed.' >&2
  echo '  Without it, we are unable to automatically set REGISTRY_PROJECT_IDENTIFIER.' >&2
else
  # This assumes that the current gcloud project is the one where data is stored.
  export REGISTRY_PROJECT_IDENTIFIER=$(gcloud config list --format 'value(core.project)')
fi

### CLIENT CONFIGURATION

# Be sure that the port setting below is correct. 8080 is the default.
export APG_REGISTRY_ADDRESS=localhost:8080
export APG_REGISTRY_AUDIENCES=http://localhost:8080

# Local calls don't use TLS.
export APG_REGISTRY_INSECURE=1

# Local calls don't need authentication.
unset APG_REGISTRY_TOKEN
unset APG_REGISTRY_API_KEY
