#!/bin/bash
#
# Copyright 2020 Google LLC. All Rights Reserved.
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
echo through the API or using the automatically-generated apg command-line tool.

echo
echo Delete everything associated with any preexisting project named "demo".
apg registry delete-project --name projects/demo

echo
echo Create a project in the registry named "demo".
apg registry create-project --project_id demo --json

echo
echo Add a API to the registry.
apg registry create-api \
    --parent projects/demo \
    --api_id petstore \
    --api.availability GENERAL \
    --api.recommended_version "1.0.0" \
    --json

echo
echo Add a version of the API to the registry.
apg registry create-version \
    --parent projects/demo/apis/petstore \
    --version_id 1.0.0 \
    --version.state "PRODUCTION" \
    --json

echo
echo Add a spec for the API version that we just added to the registry.
apg registry create-spec \
    --parent projects/demo/apis/petstore/versions/1.0.0 \
    --spec_id openapi.yaml \
    --spec.contents `encode-spec < petstore/1.0.0/openapi.yaml@r0` \
    --json

echo
echo Get the API spec.
apg registry get-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --view full \
    --json

echo
echo Get the contents of the API spec.
apg registry get-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --view full --json | \
    jq '.contents' -r | \
    decode-spec

echo
echo Update an attribute of the spec.
apg registry update-spec \
	--spec.name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
	--spec.style "openapi/v3+gzip" \
    --json

echo
echo Get the modifed API spec.
apg registry get-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --view full \
    --json

echo
echo Update the spec to new contents.
apg registry update-spec \
	--spec.name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
	--spec.contents `encode-spec < petstore/1.0.0/openapi.yaml@r1` \
    --json

echo
echo Again update the spec to new contents.
apg registry update-spec \
	--spec.name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
	--spec.contents `encode-spec < petstore/1.0.0/openapi.yaml@r2` \
    --json

echo
echo Make a third update of the spec contents.
apg registry update-spec \
	--spec.name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
	--spec.contents `encode-spec < petstore/1.0.0/openapi.yaml@r3`

echo
echo Get the API spec.
apg registry get-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --view full \
    --json

echo
echo List the revisions of the spec.
apg registry list-spec-revisions \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json

echo
echo List just the names of the revisions of the spec.
apg registry list-spec-revisions \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[].name' -r 

echo
echo Get the latest revision of the spec.
apg registry list-spec-revisions \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[0].name' -r 

echo
echo Get the oldest revision of the spec.
apg registry list-spec-revisions \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[-1].name' -r 

ORIGINAL=`apg registry list-spec-revisions \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[-1].name' -r`

ORIGINAL_HASH=`apg registry list-spec-revisions \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[-1].hash' -r`

echo
echo Tag a spec revision.
apg registry tag-spec-revision --name $ORIGINAL --tag og --json

echo
echo Get a spec by its tag.
apg registry get-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml@og \
    --view basic \
    --json

echo
echo Print the hash of the current spec revision.
apg registry get-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.hash' -r

echo
echo Rollback to a prior spec revision.
apg registry rollback-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --revision_id og \
    --json

echo
echo Print the hash of the current spec revision after the rollback.
apg registry get-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.hash' -r

echo
echo Print the original hash. 
echo $ORIGINAL_HASH

echo
echo Delete a spec revision.
apg registry delete-spec-revision --name $ORIGINAL

ORIGINAL2=`apg registry list-spec-revisions \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[-1].name' -r`

echo
echo Verify that the spec has changed.
echo $ORIGINAL2 should not be $ORIGINAL

echo
echo List the revision tags of a spec.
apg registry list-spec-revision-tags \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json

echo
echo Verify that when listing specs, we only get the current revision of each spec.
apg registry list-specs \
    --parent projects/demo/apis/petstore/versions/1.0.0 \
    --json

echo
echo Add some labels to entities in the registry.
apg registry create-label \
    --parent projects/demo/apis/petstore \
    --label_id pets \
    --json

apg registry create-label \
    --parent projects/demo/apis/petstore \
    --label_id retail \
    --json

apg registry create-label \
    --parent projects/demo/apis/petstore \
    --label_id stock \
    --json

apg registry create-label \
    --parent projects/demo/apis/petstore/versions/1.0.0 \
    --label_id published \
    --json

echo
echo Set some properties on entities in the registry.
apg registry create-property \
    --parent projects/demo/apis/petstore \
    --property_id source \
    --property.value string_value \
    --property.value.string_value "https://github.com/OAI/OpenAPI-Specification" \
    --json

apg registry create-property \
    --parent projects/demo/apis/petstore/versions/1.0.0 \
    --property_id score \
    --property.value int64_value \
    --property.value.int64_value 100 \
    --json

apg registry create-property \
    --parent projects/demo/apis/petstore/versions/1.0.0 \
    --property_id boys \
    --property.value string_value \
    --property.value.string_value "Neil Tennant, Chris Lowe" \
    --json

echo
echo Export a YAML summary of the demo project.
registry export projects/demo > demo.yaml
cat demo.yaml
