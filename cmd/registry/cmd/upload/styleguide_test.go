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

package upload

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestStyleGuideUpload(t *testing.T) {
	tests := []struct {
		desc     string
		project  string
		filePath string
		want     *rpc.StyleGuide
	}{
		{
			desc:     "simple style guide upload",
			project:  "upload-styleguide-demo",
			filePath: filepath.Join("testdata", "styleguide.yaml"),
			want: &rpc.StyleGuide{
				Id: "test-styleguide",
				MimeTypes: []string{
					"application/x.openapi+gzip;version=2",
				},
				Guidelines: []*rpc.Guideline{
					{
						Id:          "refproperties",
						DisplayName: "Govern Ref Properties",
						Description: "This guideline governs properties for ref fields on specs.",
						Rules: []*rpc.Rule{
							{
								Id:             "norefsiblings",
								Description:    "An object exposing a $ref property cannot be further extended with additional properties.",
								Linter:         "spectral",
								LinterRulename: "no-$ref-siblings",
								Severity:       rpc.Rule_ERROR,
							},
						},
						Status: rpc.Guideline_ACTIVE,
					},
				},
				Linters: []*rpc.Linter{
					{
						Name: "spectral",
						Uri:  "https://github.com/stoplightio/spectral",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			client, err := connection.NewClient(ctx)
			if err != nil {
				t.Fatalf("Setup: Failed to create client: %s", err)
			}
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatalf("Setup: Failed to create client: %s", err)
			}
			err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  "projects/" + test.project,
				Force: true,
			})
			if err != nil && status.Code(err) != codes.NotFound {
				t.Fatalf("Setup: Failed to delete test project: %s", err)
			}
			_, err = adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
				ProjectId: test.project,
				Project: &rpc.Project{
					DisplayName: "Demo",
					Description: "A demo catalog",
				},
			})
			if err != nil {
				t.Fatalf("Failed to create project %s: %s", test.project, err.Error())
			}

			cmd := Command()
			args := []string{"styleguide", test.filePath, "--project-id", test.project}
			cmd.SetArgs(args)
			if err = cmd.Execute(); err != nil {
				t.Fatalf("Execute() with args %v returned error: %s", args, err)
			}

			req := &rpc.GetArtifactContentsRequest{
				Name: "projects/" + test.project + "/locations/global/artifacts/test-styleguide",
			}

			styleguide := &rpc.StyleGuide{}
			body, err := client.GetArtifactContents(ctx, req)
			if err != nil {
				t.Fatalf("GetArtifactContents() returned error: %s", err)
			}

			contents := body.GetData()
			err = proto.Unmarshal(contents, styleguide)
			if err != nil {
				t.Fatalf("Unmarshal() returned error: %s", err)
			}

			// Verify the style guide definition is correct
			opts := cmp.Options{
				protocmp.Transform(),
			}

			if diff := cmp.Diff(test.want, styleguide, opts); diff != "" {
				t.Errorf("GetArtifactContents returned unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}
