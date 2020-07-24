#!/bin/bash

echo delete everything associated with a project
apg registry delete-project --name projects/sample

echo create a project in the registry
apg registry create-project --project_id sample

echo add a product to the registry
apg registry create-product --parent projects/sample --product_id petstore

echo add a version to the registry
apg registry create-version --parent projects/sample/products/petstore --version_id 1.0.0

echo add a spec to the registry
apg registry create-spec \
    --parent projects/sample/products/petstore/versions/1.0.0 \
    --spec_id openapi.yaml \
    --spec.contents `encode-spec < petstore/1.0.0/openapi.yaml@r0`

echo get a spec
apg registry get-spec \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --view full \
    --json

echo get the spec contents
apg registry get-spec \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --view full --json | \
    jq '.contents' -r | \
    base64 --decode | \
    gunzip

echo update a spec attribute
apg registry update-spec \
	--spec.name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
	--spec.style "openapi/v3+gzip"

echo get the spec
apg registry get-spec \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --view full \
    --json

echo update the spec to new contents
apg registry update-spec \
	--spec.name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
	--spec.contents `encode-spec < petstore/1.0.0/openapi.yaml@r1`

echo update the spec to new contents
apg registry update-spec \
	--spec.name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
	--spec.contents `encode-spec < petstore/1.0.0/openapi.yaml@r2`

echo update the spec to new contents
apg registry update-spec \
	--spec.name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
	--spec.contents `encode-spec < petstore/1.0.0/openapi.yaml@r3`

echo get the spec
apg registry get-spec \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --view full \
    --json

echo list spec revisions
apg registry list-spec-revisions \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --json

echo list just the revision names
apg registry list-spec-revisions \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[].name' -r 

echo get the latest revision
apg registry list-spec-revisions \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[0].name' -r 

echo get the oldest revision 
apg registry list-spec-revisions \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[-1].name' -r 

ORIGINAL=`apg registry list-spec-revisions \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[-1].name' -r`

ORIGINAL_HASH=`apg registry list-spec-revisions \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[-1].hash' -r`

echo tag a spec revision
apg registry tag-spec-revision --name $ORIGINAL --tag og

echo get a spec by its tag
apg registry get-spec \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml@og \
    --view basic \
    --json

echo print current hash
apg registry get-spec \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.hash' -r

echo rollback a spec revision
apg registry rollback-spec --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml --revision_id og

echo print current hash after rollback
apg registry get-spec \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.hash' -r

echo print original hash
echo $ORIGINAL_HASH

echo delete a spec revision
apg registry delete-spec-revision --name $ORIGINAL

ORIGINAL2=`apg registry list-spec-revisions \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --json | \
    jq '.specs[-1].name' -r`

echo $ORIGINAL2 should not be $ORIGINAL

echo list revision tags
apg registry list-spec-revision-tags \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --json

echo list specs should only return the most current revision of each spec
apg registry list-specs \
    --parent projects/sample/products/petstore/versions/1.0.0 \
    --json

echo delete the spec
apg registry delete-spec  --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml

echo list spec-revisions should return nothing now
apg registry list-spec-revisions \
    --name projects/sample/products/petstore/versions/1.0.0/specs/openapi.yaml \
    --json

