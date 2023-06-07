#!/bin/bash
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

# This script uses Docker buildx to build multi-platform images of
# the registry server and related components.
 
# Available platforms will depend on the local installation of Docker.

# ORGANIZATION should be the dockerhub organization or gcr.io/project that will host the images

# TARGET selects one of the container groups below and adds a suffix (if set)

if [[ $TARGET == "dev" ]]
then
  # If TARGET is specified as "dev", a minimal set of containers are built.
  # Container names have the suffix "-dev".
  SUFFIX="-$TARGET"
  CONTAINERS=("registry-server" "registry-tools")
  PLATFORMS="linux/arm64"
else
  # Otherwise, all containers are built.
  SUFFIX=""
  CONTAINERS=("registry-server" "registry-tools")
  PLATFORMS="linux/amd64,linux/arm64"
fi

# This builds each desired container sequentially.
for CONTAINER in ${CONTAINERS[*]}; do
docker buildx build \
	--file containers/${CONTAINER}/Dockerfile \
	--tag ${ORGANIZATION}/${CONTAINER}${SUFFIX}:latest \
	--platform $PLATFORMS \
	--progress plain \
	--push \
	.
done
