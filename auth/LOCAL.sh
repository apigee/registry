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

# Be sure that the port setting below is correct. 8080 is the default.
APG_REGISTRY_ADDRESS="localhost:8080"

# Local calls don't use TLS.
APG_REGISTRY_INSECURE="true"

APG_REGISTRY_LOCATION="global"

registry config configurations create local \
  --registry.insecure="${APG_REGISTRY_INSECURE}" \
  --registry.address="${APG_REGISTRY_ADDRESS}" \
  --registry.location="${APG_REGISTRY_LOCATION}"
