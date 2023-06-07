// Copyright 2022 Google LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package patch

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/pkg/encoding"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
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
			project := names.Project{ProjectID: "patch-project-test"}
			registryClient, adminClient := grpctest.SetupRegistry(ctx, t, project.ProjectID, nil)

			if err := Apply(ctx, registryClient, adminClient, nil, project.String()+"/locations/global", true, 10, test.root); err != nil {
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
			actual, err := encoding.EncodeYAML(model)
			if err != nil {
				t.Fatalf("encoding.EncodeYAML(%+v) returned an error: %s", model, err)
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
				actual, err := encoding.EncodeYAML(model)
				if err != nil {
					t.Fatalf("encoding.EncodeYAML(%+v) returned an error: %s", model, err)
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
