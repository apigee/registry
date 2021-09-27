#!/bin/bash
#
# Copyright 2021 Google LLC. All Rights Reserved.
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

SERVICE_PROTOS=(
	google/cloud/apigee/registry/v1/registry_models.proto
	google/cloud/apigee/registry/v1/registry_service.proto
)

echo "Generating Go client library for ${SERVICE_PROTOS[@]}"
protoc ${SERVICE_PROTOS[*]} \
	--proto_path='.' \
	--proto_path='third_party/api-common-protos' \
	--go_gapic_opt='go-gapic-package=github.com/apigee/registry/gapic;gapic' \
	--go_gapic_opt='grpc-service-config=gapic/grpc_service_config.json' \
	--go_gapic_opt='module=github.com/apigee/registry' \
	--go_gapic_out='.'

# Add an accessor for the underlying gRPC client of the GAPIC client.
cat >> gapic/registry_client.go <<END

func (c *RegistryClient) GrpcClient() rpcpb.RegistryClient {
	return c.internalClient.(*registryGRPCClient).registryClient
}
END

# Patch the generated GAPIC to send Authorization tokens with insecure requests.
# This allows the registry command line tools to test container builds.
FILE=gapic/doc.go
PATCH='\
insecure := os.Getenv("APG_REGISTRY_INSECURE")\
token := os.Getenv("APG_REGISTRY_TOKEN")\
if insecure == "1" \&\& token != "" {\
  out["authorization"] = append(out["authorization"], "Bearer "+token)\
}\
return metadata.NewOutgoingContext(ctx, out)\
' # No ' or / in the patch; escape & and newlines.
sed -i.bak "s/return metadata.NewOutgoingContext(ctx, out)/${PATCH}/" "${FILE}"
rm "${FILE}.bak"
gofmt -w "${FILE}"
if ! grep --quiet APG_REGISTRY_INSECURE "${FILE}"; then
  echo "Patching GAPIC library failed."
  exit 1
fi
