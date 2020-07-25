#!/bin/bash

# This is a directroy containing the registry protos.
export REGISTRY_PROTOS="google/cloud/apigee/registry/v1alpha1"

# This points to the .proto files distributed with protoc.
export PROTO_PROTOS="$HOME/local/include"

# This is a third_party directory containing .proto files used by many APIs.
export ANNOTATION_PROTOS="../../../third_party/api-common-protos"

# This is a third_party directory containing message protos used to store API metrics.
export METRICS_PROTOS="../../../third_party/gnostic/metrics"

mkdir -p lib/generated 

echo "Generating Dart support code."
protoc \
	--proto_path=../../.. \
	--proto_path=${ANNOTATION_PROTOS} \
	--proto_path=${PROTO_PROTOS} \
	${PROTO_PROTOS}/google/protobuf/any.proto \
	${PROTO_PROTOS}/google/protobuf/timestamp.proto \
	${PROTO_PROTOS}/google/protobuf/field_mask.proto \
	${PROTO_PROTOS}/google/protobuf/empty.proto \
	${REGISTRY_PROTOS}/registry_models.proto \
	${REGISTRY_PROTOS}/registry_service.proto \
	${METRICS_PROTOS}/complexity.proto \
	--dart_out=grpc:lib/generated
