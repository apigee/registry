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

# Run all thhe commands from the root directory

# Deploy the registry server in GKE
make deploy-gke

# Setup auth
source auth/GKE.sh

# Create demo project
apg admin create-project --project_id demo --json

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