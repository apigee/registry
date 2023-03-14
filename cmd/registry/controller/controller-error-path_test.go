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

	"github.com/apigee/registry/pkg/application/controller"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/seeder"
)

// Tests for error paths in the controller

func TestControllerErrors(t *testing.T) {
	const projectID = "controller-error-demo"
	ctx := context.Background()
	registryClient, _ := grpctest.SetupRegistry(ctx, t, projectID, []seeder.RegistryResource{
		&rpc.ApiSpec{
			Name:     "projects/controller-test/locations/global/apis/petstore/versions/1.0.0/specs/openapi",
			MimeType: gzipOpenAPIv3,
		},
		&rpc.ApiSpec{
			Name:     "projects/controller-test/locations/global/apis/petstore/versions/1.0.1/specs/openapi",
			MimeType: gzipOpenAPIv3,
		},
		&rpc.ApiSpec{
			Name:     "projects/controller-test/locations/global/apis/petstore/versions/1.1.0/specs/openapi",
			MimeType: gzipOpenAPIv3,
		},
	})
	lister := &RegistryLister{RegistryClient: registryClient}

	tests := []struct {
		desc              string
		generatedResource *controller.GeneratedResource
	}{
		{
			desc: "Non-existing reference in dependencies",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/lintstats-gnostic",
				Dependencies: []*controller.Dependency{
					{
						Pattern: "$resource.spec", // Correct pattern should be: $resource.version
					},
				},
				Action: "registry compute lintstats $resource.spec --linter gnostic",
			},
		},
		{
			desc: "Incorrect reference keyword",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/specs/-/artifacts/lint-gnostic",
				Dependencies: []*controller.Dependency{
					{
						Pattern: "$resource.apispec", // Correct pattern should be: $resource.spec
					},
				},
				Action: "registry compute lint $resource.apispec --linter gnostic",
			},
		},
		{
			desc: "Nonexistent dependency resource",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/lintstats-gnostic",
				Dependencies: []*controller.Dependency{
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
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/specs/-/artifacts/lintstats-gnostic",
				Dependencies: []*controller.Dependency{
					{
						Pattern: "$resource.spec",
					},
				},
				Action: "registry compute lintstats $resource.artifact --linter gnostic", // Correct reference should be: $resource.spec
			},
		},
		{
			desc: "Incorrect resource pattern",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/specs/-/artifacts/lintstats-gnostic", // Correct pattern should be: apis/-/versions/-/specs/-/artifacts/lintstats-gnostic
				Dependencies: []*controller.Dependency{
					{
						Pattern: "$resource.spec",
					},
				},
				Action: "registry compute lintstats $resource.specs --linter gnostic",
			},
		},
		{
			desc: "Incorrect matching",
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/summary",
				Dependencies: []*controller.Dependency{
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
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/versions/-/artifacts/score",
				Dependencies: []*controller.Dependency{
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
			generatedResource: &controller.GeneratedResource{
				Pattern: "apis/-/artifacts/summary",
				Dependencies: []*controller.Dependency{
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

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Test GeneratedResource pattern
			actions, err := processManifestResource(ctx, lister, projectID, test.generatedResource)
			if err == nil {
				t.Errorf("Expected processManifestResource() to return an error, got: %v", actions)
			}
		})
	}
}
