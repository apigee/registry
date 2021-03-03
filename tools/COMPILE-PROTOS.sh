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
# Compile .proto files into the files needed to build the registry server and
# command-line tools.
#

echo "Updating tool dependencies."
go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
go get -u github.com/googleapis/gapic-generator-go/cmd/protoc-gen-go_gapic
go get -u github.com/googleapis/gapic-generator-go/cmd/protoc-gen-go_cli
go get -u github.com/googleapis/api-linter/cmd/api-linter

# add installed dependencies to PATH in case they aren't already
export PATH=$PATH:$(go env GOBIN):$(go env GOPATH)/bin

echo "Clearing any previously-generated files."
rm -rf rpc/*.go gapic/*.go cmd/apg/*.go
mkdir -p rpc gapic cmd/apg

ANNOTATIONS="third_party/api-common-protos"

PROTOS=( \
	google/cloud/apigee/registry/applications/v1alpha1/registry_index.proto \
	google/cloud/apigee/registry/applications/v1alpha1/registry_lint.proto \
	google/cloud/apigee/registry/internal/v1/registry_map.proto \
	google/cloud/apigee/registry/v1/registry_models.proto \
	google/cloud/apigee/registry/v1/registry_notifications.proto \
	google/cloud/apigee/registry/v1/registry_service.proto \
)

echo "Running the API linter."
for p in ${PROTOS[@]}; do
  echo "api-linter $p"
  api-linter -I ${ANNOTATIONS} $p
done

echo "Generating proto support code."
protoc --proto_path=. --proto_path=${ANNOTATIONS} \
	${PROTOS[*]} \
	--go_out=plugins=grpc:rpc

# fix the location of proto output files
mv rpc/github.com/apigee/registry/rpc/* rpc
rm -rf rpc/github.com

echo "Generating GAPIC library."
protoc --proto_path=. --proto_path=${ANNOTATIONS} \
	${PROTOS[*]} \
	--go_gapic_out gapic \
	--go_gapic_opt="go-gapic-package=github.com/apigee/registry/gapic;gapic" \
	--go_gapic_opt="grpc-service-config=gapic/grpc_service_config.json"

# fix the location of gapic output files
mv gapic/github.com/apigee/registry/gapic/* gapic
rm -rf gapic/github.com

# add an accessor for the underlying gRPC client of the GAPIC client
cat >> gapic/registry_client.go <<END

func (c *RegistryClient) GrpcClient() rpcpb.RegistryClient {
	return c.registryClient
}
END

# patch the generated GAPIC to send Authorization tokens with insecure requests
# (this allows the registry command line tools to test container builds)
sed -i 's/return metadata.NewOutgoingContext(ctx, out)/insecure := os.Getenv("APG_REGISTRY_INSECURE")\ntoken := os.Getenv("APG_REGISTRY_TOKEN")\nif insecure == "1" \&\& token != "" \{ \nout["authorization"] = append(out["authorization"], "Bearer "+token) \n\}\nreturn metadata.NewOutgoingContext(ctx, out)/' gapic/doc.go
gofmt -w gapic/doc.go

echo "Generating GAPIC-based CLI."
protoc --proto_path=. --proto_path=${ANNOTATIONS} \
	${PROTOS[*]} \
  	--go_cli_out cmd/apg \
  	--go_cli_opt "root=apg" \
  	--go_cli_opt "gapic=github.com/apigee/registry/gapic"

echo "Generating descriptor set for Envoy gRPC-JSON Transcoding."
protoc --proto_path=. --proto_path=${ANNOTATIONS} \
	${PROTOS[*]} \
	--include_imports \
    --include_source_info \
    --descriptor_set_out=deployments/envoy/proto.pb
