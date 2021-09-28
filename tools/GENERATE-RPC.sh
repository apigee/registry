#!/bin/bash
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

set -e

if [ ! -d "third_party/api-common-protos" ]
then
  git clone https://github.com/googleapis/api-common-protos third_party/api-common-protos
fi

ALL_PROTOS=(
	google/cloud/apigee/registry/applications/v1alpha1/*.proto
	google/cloud/apigee/registry/internal/v1/*.proto
	google/cloud/apigee/registry/v1/*.proto
)

for proto in ${ALL_PROTOS[@]}; do
	echo "Generating Go types for $proto"
	protoc $proto --proto_path='.' --proto_path='third_party/api-common-protos' --go_opt='module=github.com/apigee/registry' --go_out='.'
done

SERVICE_PROTOS=(
	google/cloud/apigee/registry/v1/registry_models.proto
	google/cloud/apigee/registry/v1/registry_service.proto
)

echo "Generating Go gRPC client/server for ${SERVICE_PROTOS[@]}"
protoc ${SERVICE_PROTOS[*]} --proto_path='.' --proto_path='third_party/api-common-protos' --go-grpc_opt='module=github.com/apigee/registry' --go-grpc_out='.'
