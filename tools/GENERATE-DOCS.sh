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
# Automatically generate API documentation.
#

echo "Updating tool dependencies."
go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc

ANNOTATIONS="third_party/api-common-protos"

PROTOS=( \
	google/cloud/apigee/registry/v1/registry_models.proto \
	google/cloud/apigee/registry/v1/registry_service.proto \
)

echo "Generating documentation."
mkdir -p ./docs
protoc --proto_path=. --proto_path=${ANNOTATIONS} \
	${PROTOS[*]} \
	--doc_out=./docs \
	--doc_opt=html,api.html
