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

echo "Clearing any previously-generated files."
rm -rf rpc/*.go gapic/*.go cmd/apg/*.go
mkdir -p rpc gapic cmd/apg

ANNOTATIONS="third_party/api-common-protos"

PROTOS=( \
	google/cloud/apigee/registry/v1alpha1/registry_models.proto \
	google/cloud/apigee/registry/v1alpha1/registry_service.proto \
	google/cloud/apigee/registry/v1alpha1/registry_notifications.proto \
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
	--go_gapic_opt "go-gapic-package=github.com/apigee/registry/gapic;gapic"

# fix the location of gapic output files
mv gapic/github.com/apigee/registry/gapic/* gapic
rm -rf gapic/github.com

echo "Generating GAPIC-based CLI."
protoc --proto_path=. --proto_path=${ANNOTATIONS} \
	${PROTOS[*]} \
  	--go_cli_out cmd/apg \
  	--go_cli_opt "root=apg" \
  	--go_cli_opt "gapic=github.com/apigee/registry/gapic"

# fix a problem in a couple of generated CLI files
sed -i -e 's/anypb.Property_MessageValue/rpcpb.Property_MessageValue/g' \
	cmd/apg/create-property.go \
	cmd/apg/update-property.go

echo "Generating descriptor set for envoy."
protoc --proto_path=. --proto_path=${ANNOTATIONS} \
	${PROTOS[*]} \
	--include_imports \
    --include_source_info \
    --descriptor_set_out=deployments/envoy/proto.pb
