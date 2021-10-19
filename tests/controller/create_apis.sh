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
