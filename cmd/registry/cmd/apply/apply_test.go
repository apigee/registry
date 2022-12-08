// Copyright 2022 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apply

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/names"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const sampleDir = "testdata/sample"

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if APG_REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestApply(t *testing.T) {
	project := names.Project{ProjectID: "apply-test"}
	parent := project.String() + "/locations/global"

	ctx := context.Background()
	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Setup: failed to create client: %+v", err)
	}
	defer adminClient.Close()

	if err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  project.String(),
		Force: true,
	}); err != nil && status.Code(err) != codes.NotFound {
		t.Errorf("Setup: failed to delete test project: %s", err)
	}

	if _, err := adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
		ProjectId: project.ProjectID,
		Project:   &rpc.Project{},
	}); err != nil {
		t.Fatalf("Setup: Failed to create test project: %s", err)
	}

	registryClient, err := connection.NewRegistryClient(ctx)
	if err != nil {
		t.Fatalf("Setup: Failed to create registry client: %s", err)
	}
	defer registryClient.Close()

	// Test API creation and export.
	// TODO: This should be split into two parts: 1) testing API creation, and 2) testing API export.
	// When API creation breaks we want to see something like FAIL: TestApply/Create_API or
	// FAIL: TestApplyAPIs/Create, not FAIL: TestApply/Create_and_Export_API, or worse FAIL: TestApply.
	{
		const filename = sampleDir + "/apis/registry.yaml"
		cmd := Command()
		cmd.SetArgs([]string{"-f", filename, "--parent", parent})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", cmd.Args, err)
		}
		expected, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read API YAML: %s", err)
		}

		got, err := registryClient.GetApi(ctx, &rpc.GetApiRequest{
			Name: project.Api("registry").String(),
		})
		if status.Code(err) == codes.NotFound {
			t.Fatalf("Expected API doesn't exist: %s", err)
		} else if err != nil {
			t.Fatalf("Failed to verify API existence: %s", err)
		}

		actual, _, err := patch.ExportAPI(ctx, registryClient, got, true)
		if err != nil {
			t.Fatalf("ExportApi(%+v) returned an error: %s", got, err)
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("GetApi(%q) returned unexpected diff: (-want +got):\n%s", got, diff)
		}
	}

	// Test artifact creation and export.
	// TODO: These should run as separate subtests to make it clear exactly which artifact types are failing.
	// Creation and export should also be separated ideally. The error message should at least make it
	// clear whether create or export is failing.
	artifacts := []string{"lifecycle", "manifest", "taxonomies", "styleguide"}
	for _, a := range artifacts {
		filename := fmt.Sprintf("%s/artifacts/%s.yaml", sampleDir, a)
		cmd := Command()
		cmd.SetArgs([]string{"-f", filename, "--parent", parent})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", cmd.Args, err)
		}
		expected, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read artifact YAML %s", err)
		}

		message, err := registryClient.GetArtifact(ctx, &rpc.GetArtifactRequest{
			Name: project.Artifact(a).String(),
		})
		if status.Code(err) == codes.NotFound {
			t.Fatalf("Expected artifact doesn't exist: %s", err)
		} else if err != nil {
			t.Fatalf("Failed to verify artifact existence: %s", err)
		}

		actual, _, err := patch.ExportArtifact(ctx, registryClient, message)
		if err != nil {
			t.Fatalf("ExportArtifact(%+v) returned an error: %s", message, err)
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("GetArtifact(%q) returned unexpected diff: (-want +got):\n%s", message, diff)
		}
	}

	if err := adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  project.String(),
		Force: true,
	}); err != nil {
		t.Logf("Cleanup: Failed to delete test project: %s", err)
	}
}

func TestApplyProject(t *testing.T) {
	// Each of these three imports should import an identical project that can be exported
	// into the structures in "sampleDir".
	tests := []struct {
		desc string
		root string
	}{
		{
			desc: "bundled",
			root: sampleDir,
		},
		{
			desc: "unbundled-hierarchical",
			root: "testdata/unbundled-hierarchical",
		},
		{
			desc: "unbundled-flat",
			root: "testdata/unbundled-flat",
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatalf("Setup: failed to create client: %+v", err)
			}
			defer adminClient.Close()
			project := names.Project{ProjectID: "apply-project-test"}
			if err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  project.String(),
				Force: true,
			}); err != nil && status.Code(err) != codes.NotFound {
				t.Errorf("Setup: failed to delete test project: %s", err)
			}

			if _, err := adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
				ProjectId: project.ProjectID,
				Project:   &rpc.Project{},
			}); err != nil {
				t.Fatalf("Setup: Failed to create test project: %s", err)
			}

			// set the configured registry.project to the test project
			config, err := connection.ActiveConfig()
			if err != nil {
				t.Fatalf("Setup: Failed to get registry configuration: %s", err)
			}
			config.Project = project.ProjectID
			connection.SetConfig(config)

			registryClient, err := connection.NewRegistryClient(ctx)
			if err != nil {
				t.Fatalf("Setup: Failed to create registry client: %s", err)
			}
			defer registryClient.Close()

			cmd := Command()
			cmd.SetArgs([]string{"-f", test.root, "-R"})
			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() with args %+v returned error: %s", cmd.Args, err)
			}

			filename := sampleDir + "/apis/registry.yaml"
			expected, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("Failed to read API YAML: %s", err)
			}

			got, err := registryClient.GetApi(ctx, &rpc.GetApiRequest{
				Name: project.Api("registry").String(),
			})
			if err != nil {
				t.Fatalf("Failed to get API: %s", err)
			}

			actual, _, err := patch.ExportAPI(ctx, registryClient, got, true)
			if err != nil {
				t.Fatalf("ExportApi(%+v) returned an error: %s", got, err)
			}

			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("GetApi(%q) returned unexpected diff: (-want +got):\n%s", got, diff)
			}

			artifacts := []string{"lifecycle", "manifest", "taxonomies"}
			for _, a := range artifacts {
				filename := fmt.Sprintf("%s/artifacts/%s.yaml", sampleDir, a)
				expected, err := os.ReadFile(filename)
				if err != nil {
					t.Fatalf("Failed to read artifact YAML %s", err)
				}

				message, err := registryClient.GetArtifact(ctx, &rpc.GetArtifactRequest{
					Name: project.Artifact(a).String(),
				})
				if status.Code(err) == codes.NotFound {
					t.Fatalf("Expected artifact doesn't exist: %s", err)
				} else if err != nil {
					t.Fatalf("Failed to verify artifact existence: %s", err)
				}

				actual, _, err := patch.ExportArtifact(ctx, registryClient, message)
				if err != nil {
					t.Fatalf("ExportArtifact(%+v) returned an error: %s", message, err)
				}

				if diff := cmp.Diff(expected, actual); diff != "" {
					t.Errorf("GetArtifact(%q) returned unexpected diff: (-want +got):\n%s", message, diff)
				}
			}

			if err := adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  project.String(),
				Force: true,
			}); err != nil {
				t.Logf("Cleanup: Failed to delete test project: %s", err)
			}
		})
	}
}

func TestUnbundledExport(t *testing.T) {
	ctx := context.Background()
	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Setup: failed to create client: %+v", err)
	}
	defer adminClient.Close()
	project := names.Project{ProjectID: "apply-project-test"}
	if err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  project.String(),
		Force: true,
	}); err != nil && status.Code(err) != codes.NotFound {
		t.Errorf("Setup: failed to delete test project: %s", err)
	}

	if _, err := adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
		ProjectId: project.ProjectID,
		Project:   &rpc.Project{},
	}); err != nil {
		t.Fatalf("Setup: Failed to create test project: %s", err)
	}

	// set the configured registry.project to the test project
	config, err := connection.ActiveConfig()
	if err != nil {
		t.Fatalf("Setup: Failed to get registry configuration: %s", err)
	}
	config.Project = project.ProjectID
	connection.SetConfig(config)

	registryClient, err := connection.NewRegistryClient(ctx)
	if err != nil {
		t.Fatalf("Setup: Failed to create registry client: %s", err)
	}
	defer registryClient.Close()

	sampleDir := "testdata/unbundled-flat"

	cmd := Command()
	cmd.SetArgs([]string{"-f", sampleDir, "-R"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %+v returned error: %s", cmd.Args, err)
	}
	artifacts := []string{
		"artifacts/lifecycle",
		"artifacts/manifest",
		"artifacts/styleguide",
		"artifacts/taxonomies",
		"apis/registry/artifacts/api-references",
		"apis/registry/deployments/prod/artifacts/deployment-references",
		"apis/registry/versions/v1/artifacts/version-references",
		"apis/registry/versions/v1/specs/openapi/artifacts/spec-references",
	}
	for _, a := range artifacts {
		t.Run("read "+a, func(t *testing.T) {
			filename := fmt.Sprintf("%s/%s.yaml", sampleDir, strings.ReplaceAll(a, "/", "-"))
			expected, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("Failed to read artifact YAML %s", err)
			}
			message, err := registryClient.GetArtifact(ctx, &rpc.GetArtifactRequest{
				Name: "projects/apply-project-test/locations/global/" + a,
			})
			if status.Code(err) == codes.NotFound {
				t.Fatalf("Expected artifact doesn't exist: %s", err)
			} else if err != nil {
				t.Fatalf("Failed to verify artifact existence: %s", err)
			}
			actual, _, err := patch.ExportArtifact(ctx, registryClient, message)
			if err != nil {
				t.Fatalf("ExportArtifact(%+v) returned an error: %s", message, err)
			}
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("GetArtifact(%q) returned unexpected diff: (-want +got):\n%s", message, diff)
			}
		})
	}
	s := "apis/registry/versions/v1/specs/openapi"
	t.Run("read "+s, func(t *testing.T) {
		filename := fmt.Sprintf("%s/%s.yaml", sampleDir, strings.ReplaceAll(s, "/", "-"))
		expected, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read spec YAML %s", err)
		}
		message, err := registryClient.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
			Name: "projects/apply-project-test/locations/global/" + s,
		})
		if status.Code(err) == codes.NotFound {
			t.Fatalf("Expected spec doesn't exist: %s", err)
		} else if err != nil {
			t.Fatalf("Failed to verify spec existence: %s", err)
		}
		actual, _, err := patch.ExportAPISpec(ctx, registryClient, message, false)
		if err != nil {
			t.Fatalf("ExportAPISpec(%+v) returned an error: %s", message, err)
		}
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("ExportAPISpec(%q) returned unexpected diff: (-want +got):\n%s", message, diff)
		}
	})
	v := "apis/registry/versions/v1"
	t.Run("read "+v, func(t *testing.T) {
		filename := fmt.Sprintf("%s/%s.yaml", sampleDir, strings.ReplaceAll(v, "/", "-"))
		expected, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read spec YAML %s", err)
		}
		message, err := registryClient.GetApiVersion(ctx, &rpc.GetApiVersionRequest{
			Name: "projects/apply-project-test/locations/global/" + v,
		})
		if status.Code(err) == codes.NotFound {
			t.Fatalf("Expected version doesn't exist: %s", err)
		} else if err != nil {
			t.Fatalf("Failed to verify version existence: %s", err)
		}
		actual, _, err := patch.ExportAPIVersion(ctx, registryClient, message, false)
		if err != nil {
			t.Fatalf("ExportAPIVersion(%+v) returned an error: %s", message, err)
		}
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("ExportAPIVersion(%q) returned unexpected diff: (-want +got):\n%s", message, diff)
		}
	})
	d := "apis/registry/deployments/prod"
	t.Run("read "+d, func(t *testing.T) {
		filename := fmt.Sprintf("%s/%s.yaml", sampleDir, strings.ReplaceAll(d, "/", "-"))
		expected, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read deployment YAML %s", err)
		}
		message, err := registryClient.GetApiDeployment(ctx, &rpc.GetApiDeploymentRequest{
			Name: "projects/apply-project-test/locations/global/" + d,
		})
		if status.Code(err) == codes.NotFound {
			t.Fatalf("Expected version doesn't exist: %s", err)
		} else if err != nil {
			t.Fatalf("Failed to verify version existence: %s", err)
		}
		actual, _, err := patch.ExportAPIDeployment(ctx, registryClient, message, false)
		if err != nil {
			t.Fatalf("ExportAPIDeployment(%+v) returned an error: %s", message, err)
		}
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("ExportAPIDeployment(%q) returned unexpected diff: (-want +got):\n%s", message, diff)
		}
	})
	a := "apis/registry"
	t.Run("read "+a, func(t *testing.T) {
		filename := fmt.Sprintf("%s/%s.yaml", sampleDir, strings.ReplaceAll(a, "/", "-"))
		expected, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read api YAML %s", err)
		}
		message, err := registryClient.GetApi(ctx, &rpc.GetApiRequest{
			Name: "projects/apply-project-test/locations/global/" + a,
		})
		if status.Code(err) == codes.NotFound {
			t.Fatalf("Expected api doesn't exist: %s", err)
		} else if err != nil {
			t.Fatalf("Failed to verify api existence: %s", err)
		}
		actual, _, err := patch.ExportAPI(ctx, registryClient, message, false)
		if err != nil {
			t.Fatalf("ExportAPIDeployment(%+v) returned an error: %s", message, err)
		}
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("ExportAPIDeployment(%q) returned unexpected diff: (-want +got):\n%s", message, diff)
		}
	})
}
