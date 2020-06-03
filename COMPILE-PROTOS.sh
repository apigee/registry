#!/bin/sh
#
# Compile .proto files into the files needed to build the registry server and
# command-line tools.
#

echo "Updating tool dependencies."
go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
go get -u github.com/googleapis/gapic-generator-go/cmd/protoc-gen-go_gapic
go get -u github.com/googleapis/gapic-generator-go/cmd/protoc-gen-go_cli

echo "Clearing any previously-generated directories."
rm -rf rpc gapic cmd/cli
mkdir -p rpc gapic cmd/cli

export ANNOTATIONS="third_party/api-common-protos"

echo "Generating proto support code."
protoc --proto_path=./proto --proto_path=${ANNOTATIONS} \
	proto/registry_models.proto \
	proto/registry_properties.proto \
	proto/registry_service.proto \
	--go_out=plugins=grpc:rpc

# fix the location of proto output files
mv rpc/apigov.dev/registry/rpc/* rpc
rm -rf rpc/apigov.dev

echo "Generating GAPIC library."
protoc --proto_path=./proto --proto_path=${ANNOTATIONS} \
	proto/registry_models.proto \
	proto/registry_properties.proto \
	proto/registry_service.proto \
	--go_gapic_out gapic \
	--go_gapic_opt "go-gapic-package=.;gapic"

echo "Generating GAPIC-based CLI."
protoc --proto_path=./proto --proto_path=${ANNOTATIONS} \
	proto/registry_models.proto \
	proto/registry_properties.proto \
	proto/registry_service.proto \
  	--go_cli_out cmd/cli \
  	--go_cli_opt "root=cli" \
  	--go_cli_opt "gapic=apigov.dev/registry/gapic"

# fix a problem in a couple of generated CLI files
sed -i -e 's/anypb.Property_MessageValue/rpcpb.Property_MessageValue/g' \
	cmd/cli/create-property.go \
	cmd/cli/update-property.go

echo "Generating descriptor set for envoy."
protoc --proto_path=./proto --proto_path=${ANNOTATIONS} \
	proto/registry_models.proto \
	proto/registry_properties.proto \
	proto/registry_service.proto \
	--include_imports \
        --include_source_info \
        --descriptor_set_out=envoy/proto.pb
