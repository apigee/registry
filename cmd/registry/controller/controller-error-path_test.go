// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
)

// Tests for error paths in the controller

func TestControllerErrors(t *testing.T) {
	tests := []struct {
		desc              string
		generatedResource *rpc.GeneratedResource
	}{
		{
			desc: "Non-existing reference in dependencies",
			generatedResource: &rpc.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/lintstats-gnostic",
				Dependencies: []*rpc.Dependency{
					{
						Pattern: "$resource.spec", // Correct pattern should be: $resource.version
					},
				},
				Action: "registry compute lintstats $resource.spec --linter gnostic",
			},
		},
		{
			desc: "Incorrect reference keyword",
			generatedResource: &rpc.GeneratedResource{
				Pattern: "apis/-/versions/-/specs/-/artifacts/lint-gnostic",
				Dependencies: []*rpc.Dependency{
					{
						Pattern: "$resource.apispec", // Correct pattern should be: $resource.spec
					},
				},
				Action: "registry compute lint $resource.apispec --linter gnostic",
			},
		},
		{
			desc: "Non-existent dependency resource",
			generatedResource: &rpc.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/lintstats-gnostic",
				Dependencies: []*rpc.Dependency{
					{
						Pattern: "$resource.version/artifacts/lint-gnostic", // There is no version level lint-gnostic artifact in the registry
					},
				},
				//Correct action should be "registry compute lintstats $resource.version --linter gnostic"
				Action: "registry compute lintstats $resource.version/artifacts/lint-gnostic --linter gnostic",
			},
		},
		{
			desc: "Incorrect reference in action",
			generatedResource: &rpc.GeneratedResource{
				Pattern: "apis/-/versions/-/specs/-/artifacts/lintstats-gnostic",
				Dependencies: []*rpc.Dependency{
					{
						Pattern: "$resource.spec",
					},
				},
				Action: "registry compute lintstats $resource.artifact --linter gnostic", // Correct reference should be: $resource.spec
			},
		},
		{
			desc: "Incorrect resource pattern",
			generatedResource: &rpc.GeneratedResource{
				Pattern: "apis/-/specs/-/artifacts/lintstats-gnostic", // Correct pattern should be: apis/-/versions/-/specs/-/artifacts/lintstats-gnostic
				Dependencies: []*rpc.Dependency{
					{
						Pattern: "$resource.spec",
					},
				},
				Action: "registry compute lintstats $resource.specs --linter gnostic",
			},
		},
		{
			desc: "Incorrect matching",
			generatedResource: &rpc.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/summary",
				Dependencies: []*rpc.Dependency{
					{
						Pattern: "$resource.api/versions/-/artifacts/complexity", // Correct pattern should be: $resource.version/artifacts/vocabulary
					},
					{
						Pattern: "$resource.version/artifacts/vocabulary",
					},
				},
				Action: "registry compute summary $resource.version",
			},
		},
		{
			desc: "Incorrect action reference",
			generatedResource: &rpc.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/score",
				Dependencies: []*rpc.Dependency{
					{
						Pattern: "$resource.version/-/artifacts/complexity",
					},
				},
				// Correct action should be: "compute summary $resource.version/artifacts/complexity"
				Action: "registry compute summary $resource.api/versions/-/artifacts/complexity",
			},
		},
		{
			desc: "Missing reference",
			generatedResource: &rpc.GeneratedResource{
				Pattern: "apis/-/artifacts/summary",
				Dependencies: []*rpc.Dependency{
					{
						Pattern: "$resource.api/versions/-/artifacts/complexity",
					},
					{
						Pattern: "$resource.api/versions/-/artifacts/vocabulary",
					},
				},
				Action: "registry compute summary $resource", // Correct action should be: compute summary $resource.api
			},
		},
	}

	const projectID = "controller-error-demo"
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			registryClient, err := connection.NewClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer registryClient.Close()
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Logf("Failed to create client: %+v", err)
				t.FailNow()
			}
			defer adminClient.Close()

			// Setup
			deleteProject(ctx, adminClient, t, "controller-test")
			createProject(ctx, adminClient, t, "controller-test")
			createApi(ctx, registryClient, t, "projects/controller-test/locations/global", "petstore")
			// Version 1.0.0
			createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.0")
			createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.0", "openapi.yaml", gzipOpenAPIv3)
			// Version 1.0.1
			createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.0.1")
			createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.0.1", "openapi.yaml", gzipOpenAPIv3)
			// Version 1.1.0
			createVersion(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore", "1.1.0")
			createSpec(ctx, registryClient, t, "projects/controller-test/locations/global/apis/petstore/versions/1.1.0", "openapi.yaml", gzipOpenAPIv3)

			// Test GeneratedResource pattern
			actions, err := processManifestResource(ctx, registryClient, projectID, test.generatedResource)
			if err == nil {
				t.Errorf("Expected processManifestResource() to return an error, got: %v", actions)
			}
		})
	}
}
