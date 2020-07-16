#!/bin/bash
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

echo "Clearing any previously-generated directories."
rm -rf rpc gapic cmd/apg
mkdir -p rpc gapic cmd/apg

ANNOTATIONS="third_party/api-common-protos"

PROTOS=( \
	proto/registry_models.proto \
	proto/registry_service.proto \
	proto/registry_notifications.proto \
)

echo "Running the API linter."
for p in ${PROTOS[@]}; do
  echo "api-linter $p"
  api-linter -I ./proto -I ${ANNOTATIONS} $p
done

echo "Generating proto support code."
protoc --proto_path=./proto --proto_path=${ANNOTATIONS} \
	${PROTOS[*]} \
	--go_out=plugins=grpc:rpc

# fix the location of proto output files
mv rpc/apigov.dev/registry/rpc/* rpc
rm -rf rpc/apigov.dev

echo "Generating GAPIC library."
protoc --proto_path=./proto --proto_path=${ANNOTATIONS} \
	${PROTOS[*]} \
	--go_gapic_out gapic \
	--go_gapic_opt "go-gapic-package=.;gapic"

echo "Generating GAPIC-based CLI."
protoc --proto_path=./proto --proto_path=${ANNOTATIONS} \
	${PROTOS[*]} \
  	--go_cli_out cmd/apg \
  	--go_cli_opt "root=apg" \
  	--go_cli_opt "gapic=apigov.dev/registry/gapic"

# fix a problem in a couple of generated CLI files
sed -i -e 's/anypb.Property_MessageValue/rpcpb.Property_MessageValue/g' \
	cmd/apg/create-property.go \
	cmd/apg/update-property.go

echo "Generating descriptor set for envoy."
protoc --proto_path=./proto --proto_path=${ANNOTATIONS} \
	${PROTOS[*]} \
	--include_imports \
        --include_source_info \
        --descriptor_set_out=envoy/proto.pb
