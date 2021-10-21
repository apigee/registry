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

apg admin create-project --project_id demo --json

apg registry create-api \
    --parent projects/demo/locations/global \
    --api_id petstore \
    --api.availability GENERAL \
    --api.recommended_version "1.0.0" \
    --json

versions=$1

if [ -z "$versions" ]
then
    versions=5
fi

for i in $( seq 1 $versions)
do
    apg registry create-api-version \
    --parent projects/demo/locations/global/apis/petstore \
    --api_version_id $i.0.0 \
    --api_version.state "PRODUCTION" \
    --json

    apg registry create-api-spec \
    --parent projects/demo/locations/global/apis/petstore/versions/$i.0.0 \
    --api_spec_id openapi.yaml \
    --api_spec.contents `registry-encode-spec < tests/controller/testdata/openapi.yaml@r0` \
    --api_spec.mime_type "application/x.openapi+gzip;version=3" \
    --json
done
