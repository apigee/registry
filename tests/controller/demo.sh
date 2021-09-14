# Run all thhe commands from the root directory

# Deploy the registry server in GKE
make deploy-gke

# Setup auth
source auth/GKE.sh

# Create demo project
apg registry create-project --project_id demo --json

# Upload a manifest for the GKE job
registry upload manifest tests/controller/testdata/manifest.yaml --project_id=demo

# Deploy controller job
export REGISTRY_MANIFEST_ID=projects/demo/locations/global/artifacts/test-manifest
make deploy-controller-job

# Create APIs in the registry
./tests/controller/create_apis.sh 10

# Should return 10 results for 10 versions
registry list projects/demo/locations/global/apis/petstore/-/versions/-/specs/-/artifacts/lint-gnostic

# Should return 10 results for 10 versions
registry list projects/demo/locations/global/apis/petstore/-/versions/-/specs/-/artifacts/complexity

# Should return 10 results for 10 versions after the second run of the job
registry list projects/demo/locations/global/apis/petstore/-/versions/-/specs/-/artifacts/lintstats-gnostic