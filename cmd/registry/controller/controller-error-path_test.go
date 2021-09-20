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
						Pattern: "$resource.spec", // Correct pattern should be: $resource.version/specs/-
					},
				},
				Action: "registry compute lint $0 --linter gnostic",
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
				Action: "registry compute lint $0 --linter gnostic",
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
				Action: "registry compute lintstats $0 --linter gnostic",
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
				Action: "registry compute lintstats $0.artifact --linter gnostic", // Correct reference should be: $0.spec/artifacts/lint-gnostic
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
				Action: "registry compute lintstats $0.artifact --linter gnostic",
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

			// Setup
			deleteProject(ctx, registryClient, t, "controller-test")
			createProject(ctx, registryClient, t, "controller-test")
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
