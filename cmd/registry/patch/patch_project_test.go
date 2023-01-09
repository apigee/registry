package patch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProjectImports(t *testing.T) {
	// Each of these three imports should import an identical project that can be exported
	// into the structures in "sampleDir".
	const sampleDir = "testdata/sample-nested"
	tests := []struct {
		desc string
		root string
	}{
		{
			desc: "sample-nested",
			root: sampleDir,
		},
		{
			desc: "sample-hierarchical",
			root: "testdata/sample-hierarchical",
		},
		{
			desc: "sample-flat",
			root: "testdata/sample-flat",
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
			project := names.Project{ProjectID: "patch-project-test"}
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

			if err := Apply(ctx, registryClient, test.root, project.String()+"/locations/global", true, 10); err != nil {
				t.Fatalf("Apply() returned error: %s", err)
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

			model, err := NewApi(ctx, registryClient, got, true)
			if err != nil {
				t.Fatalf("NewApi(%+v) returned an error: %s", got, err)
			}
			actual, err := Encode(model)
			if err != nil {
				t.Fatalf("Encode(%+v) returned an error: %s", model, err)
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

				model, err := NewArtifact(ctx, registryClient, message)
				if err != nil {
					t.Fatalf("NewArtifact(%+v) returned an error: %s", message, err)
				}
				actual, err := Encode(model)
				if err != nil {
					t.Fatalf("Encode(%+v) returned an error: %s", model, err)
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

func TestProjectExport(t *testing.T) {
	tests := []struct {
		desc string
		root string
	}{
		{
			desc: "sample-hierarchical",
			root: "testdata/sample-hierarchical",
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
			project := names.Project{ProjectID: "patch-export-project-test"}
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

			if err := Apply(ctx, registryClient, test.root, project.String()+"/locations/global", true, 1); err != nil {
				t.Fatalf("Apply() returned error: %s", err)
			}

			tempDir, err := os.MkdirTemp("", "sample-export-")
			if err != nil {
				t.Fatalf("Setup: Failed to create export directory: %s", err)
			}
			defer os.RemoveAll(tempDir)

			taskQueue, wait := core.WorkerPool(ctx, 1)
			err = ExportProject(ctx, registryClient, project, tempDir, taskQueue, false)
			if err != nil {
				t.Fatalf("Setup: Failed to export project: %s", err)
			}
			wait()

			// does the exported directory contain everything?
			if err := filepath.Walk(test.root, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				refBytes, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				path = strings.TrimPrefix(path, test.root)
				newFilename := filepath.Join(tempDir, project.ProjectID, path)
				newBytes, err := os.ReadFile(newFilename)
				if err != nil {
					return err
				}
				if diff := cmp.Diff(newBytes, refBytes); diff != "" {
					t.Errorf("mismatched export %s %+v", newFilename, diff)
				}
				return nil
			}); err != nil {
				t.Fatalf("Setup: Failed to export project: %s", err)
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
