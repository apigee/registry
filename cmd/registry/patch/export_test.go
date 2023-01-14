// Copyright 2023 Google LLC. All Rights Reserved.
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

package patch

import (
	"context"
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

func TestExport(t *testing.T) {
	tests := []struct {
		desc string
		root string
	}{
		{
			desc: "sample",
			root: "testdata/sample-hierarchical",
		},
	}
	for _, test := range tests {
		ctx := context.Background()
		adminClient, err := connection.NewAdminClient(ctx)
		if err != nil {
			t.Fatalf("Setup: failed to create client: %+v", err)
		}
		project := names.Project{ProjectID: "patch-export-test"}
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
		if err := Apply(ctx, registryClient, test.root, project.String()+"/locations/global", true, 1); err != nil {
			t.Fatalf("Apply() returned error: %s", err)
		}
		tempDir, err := os.MkdirTemp("", "sample-export-")
		if err != nil {
			t.Fatalf("Setup: Failed to create export directory: %s", err)
		}
		t.Cleanup(func() {
			if err := adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  project.String(),
				Force: true,
			}); err != nil {
				t.Logf("Cleanup: Failed to delete test project: %s", err)
			}
			adminClient.Close()
			registryClient.Close()
			os.RemoveAll(tempDir)
		})

		t.Run(test.desc+"-project", func(t *testing.T) {
			taskQueue, wait := core.WorkerPool(ctx, 1)
			err := ExportProject(ctx, registryClient, project, tempDir, taskQueue)
			if err != nil {
				t.Fatalf("Failed to export project: %s", err)
			}
			wait()
			compareExportedFiles(t, test.root, "", tempDir, project.ProjectID)
		})

		t.Run(test.desc+"-api", func(t *testing.T) {
			taskQueue, wait := core.WorkerPool(ctx, 1)
			err := ExportAPI(ctx, registryClient, project.Api("registry"), true, tempDir, taskQueue)
			if err != nil {
				t.Fatalf("Failed to export api: %s", err)
			}
			wait()
			compareExportedFiles(t, test.root, "apis/registry", tempDir, project.ProjectID)
		})

		t.Run(test.desc+"-version", func(t *testing.T) {
			taskQueue, wait := core.WorkerPool(ctx, 1)
			err := ExportAPIVersion(ctx, registryClient, project.Api("registry").Version("v1"), true, tempDir, taskQueue)
			if err != nil {
				t.Fatalf("Failed to export version: %s", err)
			}
			wait()
			compareExportedFiles(t, test.root, "apis/registry/versions/v1", tempDir, project.ProjectID)
		})

		t.Run(test.desc+"-spec", func(t *testing.T) {
			taskQueue, wait := core.WorkerPool(ctx, 1)
			err := ExportAPISpec(ctx, registryClient, project.Api("registry").Version("v1").Spec("openapi"), true, tempDir, taskQueue)
			if err != nil {
				t.Fatalf("Failed to export spec: %s", err)
			}
			wait()
			compareExportedFiles(t, test.root, "apis/registry/versions/v1/specs/openapi", tempDir, project.ProjectID)
		})

		t.Run(test.desc+"-deployment", func(t *testing.T) {
			taskQueue, wait := core.WorkerPool(ctx, 1)
			err := ExportAPIDeployment(ctx, registryClient, project.Api("registry").Deployment("prod"), true, tempDir, taskQueue)
			if err != nil {
				t.Fatalf("Failed to export deployment: %s", err)
			}
			wait()
			compareExportedFiles(t, test.root, "apis/registry/deployments/prod", tempDir, project.ProjectID)
		})

		t.Run(test.desc+"-artifact", func(t *testing.T) {
			taskQueue, wait := core.WorkerPool(ctx, 1)
			err := ExportArtifact(ctx, registryClient, project.Api("registry").Artifact("api-references"), tempDir, taskQueue)
			if err != nil {
				t.Fatalf("Failed to export artifact: %s", err)
			}
			wait()
			compareExportedFiles(t, test.root, "apis/registry/artifacts", tempDir, project.ProjectID)
		})
	}
}

func compareExportedFiles(t *testing.T, root, top, tempDir, projectID string) {
	if err := filepath.Walk(filepath.Join(root, top), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		refBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		path = strings.TrimPrefix(path, root)
		newFilename := filepath.Join(tempDir, projectID, path)
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
}
