# Copyright 2021 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

echo 'Delete existing project'
apg registry delete-project --name projects/capabilities-demo

echo 'Create a new project capabilities-demo'
apg registry create-project --project_id capabilities-demo --json

echo 'Create an API'
apg registry create-api \
    --parent projects/capabilities-demo \
    --api_id petstore \
    --api.availability GENERAL \
    --api.recommended_version "1.0.0" \
    --json

echo 'Create an API version'
apg registry create-api-version \
    --parent projects/capabilities-demo/apis/petstore \
    --api_version_id 1.0.0 \
    --api_version.state "PRODUCTION" \
    --json

echo 'Create a spec'
apg registry create-api-spec \
    --parent projects/capabilities-demo/apis/petstore/versions/1.0.0 \
    --api_spec_id openapi.yaml \
    --api_spec.contents `registry-encode-spec < cmd/capabilities/testdata/openapi.yaml` \
    --api_spec.mime_type "application/x.openapi+gzip;version=3" \
    --json

sleep 10

echo 'Check if the lint artifact is calculated for this spec'
registry list projects/capabilities-demo/apis/petstore/versions/1.0.0/specs/-/artifacts/lint-gnostic