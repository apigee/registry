#!/bin/bash
#
# Copyright 2020 Google LLC.
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

echo This walkthrough script demonstrates key registry operations that can be performed
echo through the API or using the registry command-line tool.

if ! type "jq" > /dev/null; then
  echo
  echo "Error: this script requires jq (https://stedolan.github.io/jq/)"
  exit 1
fi

PROJECT=demo

echo
echo Delete everything associated with any preexisting project named $PROJECT.
registry rpc admin delete-project --name projects/$PROJECT --force

echo
echo Create a registry project with id $PROJECT.
registry rpc admin create-project --project_id $PROJECT --json

echo
echo Add a API to the registry.
registry rpc create-api \
    --parent projects/$PROJECT/locations/global \
    --api_id petstore \
    --api.availability GENERAL \
    --api.recommended_version "projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0" \
    --json

echo
echo Add a version of the API to the registry.
registry rpc create-api-version \
    --parent projects/$PROJECT/locations/global/apis/petstore \
    --api_version_id 1.0.0 \
    --api_version.state "PRODUCTION" \
    --json

echo
echo Add a spec for the API version that we just added to the registry.
registry rpc create-api-spec \
    --parent projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0 \
    --api_spec_id openapi \
    --api_spec.contents `registry-encode-spec < testdata/openapi.yaml@r0` \
    --json

echo
echo Get the API spec.
registry rpc get-api-spec \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
    --json

echo
echo Get the contents of the API spec.
registry rpc get-api-spec-contents \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
    --json | \
    jq '.data' -r | \
    registry-decode-spec

echo
echo Update an attribute of the spec.
registry rpc update-api-spec \
	--api_spec.name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
	--api_spec.mime_type "application/x.openapi+gzip;version=3" \
    --json

echo
echo Get the modified API spec.
registry rpc get-api-spec \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
    --json

echo
echo Update the spec to new contents.
registry rpc update-api-spec \
	--api_spec.name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
	--api_spec.contents `registry-encode-spec < testdata/openapi.yaml@r1` \
    --json

echo
echo Again update the spec to new contents.
registry rpc update-api-spec \
	--api_spec.name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
	--api_spec.contents `registry-encode-spec < testdata/openapi.yaml@r2` \
    --json

echo
echo Make a third update of the spec contents.
registry rpc update-api-spec \
	--api_spec.name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
	--api_spec.contents `registry-encode-spec < testdata/openapi.yaml@r3`

echo
echo Get the API spec.
registry rpc get-api-spec \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
    --json

echo
echo List the revisions of the spec.
registry rpc list-api-spec-revisions \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
    --json

echo
echo List just the names of the revisions of the spec.
registry rpc list-api-spec-revisions \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
    --json | \
    jq '.apiSpecs[].name' -r 

echo
echo Get the latest revision of the spec.
registry rpc list-api-spec-revisions \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
    --json | \
    jq '.apiSpecs[0].name' -r 

echo
echo Get the oldest revision of the spec.
registry rpc list-api-spec-revisions \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
    --json | \
    jq '.apiSpecs[-1].name' -r 

ORIGINAL=`registry rpc list-api-spec-revisions \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
    --json | \
    jq '.apiSpecs[-1].name' -r`

ORIGINAL_HASH=`registry rpc list-api-spec-revisions \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
    --json | \
    jq '.apiSpecs[-1].hash' -r`

echo
echo Tag a spec revision.
registry rpc tag-api-spec-revision --name $ORIGINAL --tag og --json

echo
echo Get a spec by its tag.
registry rpc get-api-spec \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi@og \
    --json

echo
echo Print the hash of the current spec revision.
registry rpc get-api-spec \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
    --json | \
    jq '.hash' -r

echo
echo Rollback to a prior spec revision.
registry rpc rollback-api-spec \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
    --revision_id og \
    --json

echo
echo Print the hash of the current spec revision after the rollback.
registry rpc get-api-spec \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
    --json | \
    jq '.hash' -r

echo
echo Print the original hash. 
echo $ORIGINAL_HASH

echo
echo Delete a spec revision.
registry rpc delete-api-spec-revision --name $ORIGINAL

ORIGINAL2=`registry rpc list-api-spec-revisions \
    --name projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0/specs/openapi \
    --json | \
    jq '.specs[-1].name' -r`

echo
echo Verify that the spec has changed.
echo $ORIGINAL2 should not be $ORIGINAL

echo
echo Verify that when listing specs, we only get the current revision of each spec.
registry rpc list-api-specs \
    --parent projects/$PROJECT/locations/global/apis/petstore/versions/1.0.0 \
    --json

echo
echo Set some artifacts on entities in the registry.
echo The contents below is the hex-encoding of "https://github.com/OAI/OpenAPI-Specification"
registry rpc create-artifact \
    --parent projects/$PROJECT/locations/global/apis/petstore \
    --artifact_id source \
    --artifact.mime_type "text/plain" \
    --artifact.contents "68747470733a2f2f6769746875622e636f6d2f4f41492f4f70656e4150492d53706563696669636174696f6e0a" \
    --json
