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

. tools/PROTOS.sh
clone_common_protos

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/googleapis/gapic-generator-go/cmd/protoc-gen-go_cli@latest

echo "Generating Go client CLI for ${SERVICE_PROTOS[@]}"
protoc ${SERVICE_PROTOS[*]} \
	--proto_path='.' \
	--proto_path=$COMMON_PROTOS_PATH \
  	--go_cli_opt='root=apg' \
  	--go_cli_opt='gapic=github.com/apigee/registry/gapic' \
	--go_cli_out='cmd/registry/cmd/rpc'

# Patch the generated CLI to be importable as a module.
find cmd/registry/cmd/rpc -name "*.go" -type f -exec sed -i.bak "s/package main/package rpc/g" {} \;

# Remove local connection configuration flags from the generated Registry service CLI
# to avoid confusion. These flags are not supported by the non-generated subcommands
# of the main tool.
REGISTRY_SERVICE=cmd/registry/cmd/rpc/registry_service.go
sed -i.bak "/RegistryServiceCmd\.PersistentFlags/d" "${REGISTRY_SERVICE}"
sed -i.bak "/RegistryConfig\.Bind/d" "${REGISTRY_SERVICE}"

# Remove local connection configuration flags from the generated Admin service CLI.
ADMIN_SERVICE=cmd/registry/cmd/rpc/admin_service.go
sed -i.bak "/AdminServiceCmd\.PersistentFlags/d" "${ADMIN_SERVICE}"
sed -i.bak "/AdminConfig\.Bind/d" "${ADMIN_SERVICE}"

# Patch the generated CLI to use "APG_REGISTRY" as the prefix for
# Admin service configuration variables. This causes the Admin service
# client to be configured with the same variables that configure the
# Registry service client.
sed -i.bak "s/APG_ADMIN/APG_REGISTRY/" "${ADMIN_SERVICE}"
if grep --quiet APG_ADMIN "${ADMIN_SERVICE}"; then
  echo "Patching CLI failed."
  exit 1
fi

# Format the files with significant patches.
gofmt -w ${REGISTRY_SERVICE} ${ADMIN_SERVICE}

# Remove all the sed-generated backup files.
rm cmd/registry/cmd/rpc/*.bak