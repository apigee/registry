#!/bin/bash

echo delete everything associated with a project
apg registry delete-project --name projects/demo

echo create a project in the registry
apg registry create-project --project_id demo

echo add a api to the registry
apg registry create-api \
    --parent projects/demo \
    --api_id petstore \
    --api.availability GENERAL \
    --api.recommended_version "1.0.0"

echo add a version to the registry
apg registry create-version \
    --parent projects/demo/apis/petstore \
    --version_id 1.0.0 \
    --version.state "PRODUCTION"

echo add a spec to the registry
apg registry create-spec \
    --parent projects/demo/apis/petstore/versions/1.0.0 \
    --spec_id openapi.yaml \
    --spec.contents `encode-spec < petstore/1.0.0/openapi.yaml@r0`

echo get a spec
apg registry get-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --view full \
    --json

echo get the spec contents
apg registry get-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --view full --json | \
    jq '.contents' -r | \
    base64 --decode | \
    gunzip

echo update a spec attribute
apg registry update-spec \
	--spec.name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
	--spec.style "openapi/v3+gzip"

echo get the spec
apg registry get-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --view full \
    --json

echo update the spec to new contents
apg registry update-spec \
	--spec.name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
	--spec.contents `encode-spec < petstore/1.0.0/openapi.yaml@r1`

echo update the spec to new contents
apg registry update-spec \
	--spec.name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
	--spec.contents `encode-spec < petstore/1.0.0/openapi.yaml@r2`

echo update the spec to new contents
apg registry update-spec \
	--spec.name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
	--spec.contents `encode-spec < petstore/1.0.0/openapi.yaml@r3`

echo get the spec
apg registry get-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --view full \
    --json

echo list spec revisions
apg registry list-spec-revisions \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json

echo list just the revision names
apg registry list-spec-revisions \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[].name' -r 

echo get the latest revision
apg registry list-spec-revisions \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[0].name' -r 

echo get the oldest revision 
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

echo tag a spec revision
apg registry tag-spec-revision --name $ORIGINAL --tag og

echo get a spec by its tag
apg registry get-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml@og \
    --view basic \
    --json

echo print current hash
apg registry get-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.hash' -r

echo rollback a spec revision
apg registry rollback-spec --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml --revision_id og

echo print current hash after rollback
apg registry get-spec \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.hash' -r

echo print original hash
echo $ORIGINAL_HASH

echo delete a spec revision
apg registry delete-spec-revision --name $ORIGINAL

ORIGINAL2=`apg registry list-spec-revisions \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[-1].name' -r`

echo $ORIGINAL2 should not be $ORIGINAL

echo list revision tags
apg registry list-spec-revision-tags \
    --name projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml \
    --json

echo list specs should only return the most current revision of each spec
apg registry list-specs \
    --parent projects/demo/apis/petstore/versions/1.0.0 \
    --json

echo add some labels
apg registry create-label \
    --parent projects/demo/apis/petstore \
    --label_id pets
apg registry create-label \
    --parent projects/demo/apis/petstore \
    --label_id retail
apg registry create-label \
    --parent projects/demo/apis/petstore \
    --label_id stock
apg registry create-label \
    --parent projects/demo/apis/petstore/versions/1.0.0 \
    --label_id published

echo set some properties
apg registry create-property \
    --parent projects/demo/apis/petstore \
    --property_id source \
    --property.value string_value \
    --property.value.string_value "https://github.com/OAI/OpenAPI-Specification"
apg registry create-property \
    --parent projects/demo/apis/petstore/versions/1.0.0 \
    --property_id score \
    --property.value int64_value \
    --property.value.int64_value 100
apg registry create-property \
    --parent projects/demo/apis/petstore/versions/1.0.0 \
    --property_id boys \
    --property.value string_value \
    --property.value.string_value "Neil Tennant, Chris Lowe"

registry export projects/demo

