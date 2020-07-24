#!/bin/bash

export ANNOTATIONS="../third_party/api-common-protos"

echo "Generating descriptor set for Endpoints."
protoc --proto_path=../proto --proto_path=${ANNOTATIONS} \
	../proto/registry_models.proto \
	../proto/registry_service.proto \
	${ANNOTATIONS}/google/api/annotations.proto \
	${ANNOTATIONS}/google/api/field_behavior.proto \
	${ANNOTATIONS}/google/api/resource.proto \
	${ANNOTATIONS}/google/api/http.proto \
	${ANNOTATIONS}/google/api/client.proto \
	--js_out=import_style=commonjs:. \
	--grpc-web_out=import_style=commonjs,mode=grpcwebtext:.
