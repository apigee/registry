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

ALL_PROTOS=(
	google/cloud/apigeeregistry/internal/v1/*.proto
	google/cloud/apigeeregistry/v1/*.proto
	google/cloud/apigeeregistry/v1/controller/*.proto
	google/cloud/apigeeregistry/v1/apihub/*.proto
	google/cloud/apigeeregistry/v1/scoring/*.proto
	google/cloud/apigeeregistry/v1/style/*.proto
	google/cloud/apigeeregistry/v1/check/*.proto
)

SERVICE_PROTOS=(
	google/cloud/apigeeregistry/v1/registry_models.proto
	google/cloud/apigeeregistry/v1/registry_service.proto
	google/cloud/apigeeregistry/v1/admin_models.proto
	google/cloud/apigeeregistry/v1/admin_service.proto
	google/cloud/apigeeregistry/v1/provisioning_service.proto
)

HOSTED_PROTOS=(
	google/cloud/apigeeregistry/v1/registry_models.proto
	google/cloud/apigeeregistry/v1/registry_service.proto
	google/cloud/apigeeregistry/v1/provisioning_service.proto
)

COMMON_PROTOS_PATH='third_party/api-common-protos'

function clone_common_protos {
	if [ ! -d $COMMON_PROTOS_PATH ]
	then
		git clone https://github.com/googleapis/api-common-protos $COMMON_PROTOS_PATH
	fi
}

# Require a specific version of protoc for generating files.
# This stabilizes the generated file output, which includes the protoc version.
. tools/PROTOC-VERSION.sh
if [ "$(protoc --version)" != "libprotoc 3.$PROTOC_VERSION" ]; then
    echo "Please update your protoc to version 3.$PROTOC_VERSION, the current required version as specified in tools/PROTOC_VERSION.sh"
    exit
fi
