#!/bin/sh

# First get some dependencies.
go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go

export ANNOTATIONS="third-party/api-common-protos"

# generate proto support code
protoc --proto_path=./proto --proto_path=${ANNOTATIONS} \
	proto/flame_models.proto \
	proto/flame_service.proto \
	--go_out=plugins=grpc:rpc

# generate gapic
protoc --proto_path=./proto --proto_path=${ANNOTATIONS} \
	proto/flame_models.proto \
	proto/flame_service.proto \
	--go_gapic_out gapic \
	--go_gapic_opt "go-gapic-package=.;gapic"

# generate gapic-based CLI
protoc --proto_path=./proto --proto_path=${ANNOTATIONS} \
	proto/flame_models.proto \
	proto/flame_service.proto \
  	--go_cli_out cli \
  	--go_cli_opt "root=flame" \
  	--go_cli_opt "gapic=apigov.dev/flame/gapic"
