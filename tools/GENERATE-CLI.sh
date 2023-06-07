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

set -e

. tools/PROTOS.sh
clone_common_protos

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/googleapis/gapic-generator-go/cmd/protoc-gen-go_cli@latest

# This directory contains the generated CLI
GENERATED='cmd/registry/cmd/rpc/generated'

echo "Generating Go client CLI for ${SERVICE_PROTOS[@]}"
protoc ${SERVICE_PROTOS[*]} \
	--proto_path='.' \
	--proto_path=$COMMON_PROTOS_PATH \
  	--go_cli_opt='root=apg' \
  	--go_cli_opt='gapic=github.com/apigee/registry/gapic' \
	--go_cli_out=$GENERATED

# Patch the generated CLI to be importable as a module.
find $GENERATED -name "*.go" -type f -exec sed -i.bak "s/package main/package generated/g" {} \;

# Patch the generated CLI cmds to use the `pkg/connection` clients so that
# we have a consistent story around flags, env vars, and config files.

REGISTRY_SERVICE=${GENERATED}/registry_service.go
sed -i.bak "/RegistryServiceCmd\.PersistentFlags/d" "${REGISTRY_SERVICE}"
sed -i.bak "/RegistryConfig/d" "${REGISTRY_SERVICE}"
sed -i.bak "/fmt/d" "${REGISTRY_SERVICE}"
sed -i.bak "/viper/,/grpc/d" "${REGISTRY_SERVICE}"
sed -i.bak "/option\.ClientOption/,/NewRegistryClient/d" "${REGISTRY_SERVICE}"
sed -i.bak "s/return/RegistryClient, err = connection.NewRegistryClient\(ctx\)\n\treturn/" "${REGISTRY_SERVICE}"
sed -i.bak "s/apigee\/registry\/gapic.*/apigee\/registry\/gapic\"\n\"github\.com\/apigee\/registry\/pkg\/connection\"/" "${REGISTRY_SERVICE}"

ADMIN_SERVICE=${GENERATED}/admin_service.go
sed -i.bak "/AdminServiceCmd\.PersistentFlags/d" "${ADMIN_SERVICE}"
sed -i.bak "/AdminConfig/d" "${ADMIN_SERVICE}"
sed -i.bak "/fmt/d" "${ADMIN_SERVICE}"
sed -i.bak "/viper/,/grpc/d" "${ADMIN_SERVICE}"
sed -i.bak "/option\.ClientOption/,/NewAdminClient/d" "${ADMIN_SERVICE}"
sed -i.bak "s/return/AdminClient, err = connection.NewAdminClient\(ctx\)\n\treturn/" "${ADMIN_SERVICE}"
sed -i.bak "s/apigee\/registry\/gapic.*/apigee\/registry\/gapic\"\n\"github\.com\/apigee\/registry\/pkg\/connection\"/" "${ADMIN_SERVICE}"

PROVISIONING_SERVICE=${GENERATED}/provisioning_service.go
sed -i.bak "/ProvisioningServiceCmd\.PersistentFlags/d" "${PROVISIONING_SERVICE}"
sed -i.bak "/ProvisioningConfig/d" "${PROVISIONING_SERVICE}"
sed -i.bak "/fmt/d" "${PROVISIONING_SERVICE}"
sed -i.bak "/viper/,/grpc/d" "${PROVISIONING_SERVICE}"
sed -i.bak "/option\.ClientOption/,/NewProvisioningClient/d" "${PROVISIONING_SERVICE}"
sed -i.bak "s/return/ProvisioningClient, err = connection.NewProvisioningClient\(ctx\)\n\treturn/" "${PROVISIONING_SERVICE}"
sed -i.bak "s/apigee\/registry\/gapic.*/apigee\/registry\/gapic\"\n\"github\.com\/apigee\/registry\/pkg\/connection\"/" "${PROVISIONING_SERVICE}"

# Format the files with significant patches.
gofmt -s -w ${REGISTRY_SERVICE} ${ADMIN_SERVICE} ${PROVISIONING_SERVICE}

# Remove all the sed backup files.
rm ${GENERATED}/*.bak
