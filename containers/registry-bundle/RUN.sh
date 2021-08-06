#!/usr/bin/env bash
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
# This script runs in a container and starts the registry-server in the
# background before running envoy.
#

# This causes bash to exit immediately if anything fails.
set -e

# run the authz server on its default port.
/authz-server -c authz.yaml &

# run the registry server on a fixed port.
REGISTRY_SERVER_PORT=8081
PORT=$REGISTRY_SERVER_PORT /registry-server -c registry-server.yaml &

# update envoy.yaml to look for the registry-server on the port we just set.
sed -i "s/8080/${REGISTRY_SERVER_PORT}/g" /etc/envoy/envoy.yaml

# update envoy.yaml to point to the container-assigned port.
sed -i "s/9999/${PORT}/g" /etc/envoy/envoy.yaml

# run envoy.
/usr/local/bin/envoy -c /etc/envoy/envoy.yaml &

# wait until any child process exits.
wait -n
