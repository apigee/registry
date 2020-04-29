#!/bin/bash

export ANNOTATIONS="../third_party/api-common-protos"

echo "Generating dart support code."
protoc --proto_path=../proto --proto_path=${ANNOTATIONS} \
	../proto/flame_models.proto \
	../proto/flame_service.proto \
	/usr/include/google/protobuf/timestamp.proto \
	/usr/include/google/protobuf/field_mask.proto \
	/usr/include/google/protobuf/empty.proto \
	--dart_out=grpc:lib/src/generated
