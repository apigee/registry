// Copyright 2022 Google LLC
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

package compute

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/names"
	metrics "github.com/google/gnostic/metrics"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestComplexity(t *testing.T) {
	tests := []struct {
		desc       string
		apiId      string
		versionId  string
		specId     string
		specFile   string
		mimeType   string
		getPattern string
		wantProto  *metrics.Complexity
	}{
		{
			desc:      "petstore-openapi",
			apiId:     "petstore",
			versionId: "1.0.0",
			specId:    "openapi",
			specFile:  "openapi.yaml",
			mimeType:  "application/x.openapi+gzip;version=3.0.0",
			wantProto: &metrics.Complexity{
				PathCount:           2,
				GetCount:            2,
				PostCount:           1,
				PutCount:            0,
				DeleteCount:         0,
				SchemaCount:         8,
				SchemaPropertyCount: 5,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			client, err := connection.NewRegistryClient(ctx)
			if err != nil {
				t.Fatalf("Setup: Failed to create client: %s", err)
			}
			adminClient, err := connection.NewAdminClient(ctx)
			if err != nil {
				t.Fatalf("Setup: Failed to create client: %s", err)
			}
			testProject := "complexity-test"
			err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  "projects/" + testProject,
				Force: true,
			})
			if err != nil && status.Code(err) != codes.NotFound {
				t.Fatalf("Setup: Failed to delete test project: %s", err)
			}
			project, err := adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
				ProjectId: testProject,
				Project: &rpc.Project{
					DisplayName: testProject,
					Description: "Test project",
				},
			})
			if err != nil {
				t.Fatalf("Failed to create project %s: %s", testProject, err.Error())
			}
			api, err := client.CreateApi(ctx, &rpc.CreateApiRequest{
				Parent: project.Name + "/locations/global",
				ApiId:  test.apiId,
				Api: &rpc.Api{
					DisplayName: test.apiId,
				},
			})
			if err != nil {
				t.Fatalf("Failed to create API %s: %s", test.apiId, err.Error())
			}
			v1, err := client.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
				Parent:       api.Name,
				ApiVersionId: test.versionId,
				ApiVersion:   &rpc.ApiVersion{},
			})
			if err != nil {
				t.Fatalf("Failed to create version %s: %s", test.versionId, err.Error())
			}
			buf, err := readAndGZipFile(t, filepath.Join("testdata", test.specFile))
			if err != nil {
				t.Fatalf("Failed reading spec contents: %s", err.Error())
			}
			req := &rpc.CreateApiSpecRequest{
				Parent:    v1.Name,
				ApiSpecId: test.specId,
				ApiSpec: &rpc.ApiSpec{
					MimeType: test.mimeType,
					Contents: buf.Bytes(),
				},
			}
			spec, err := client.CreateApiSpec(ctx, req)
			if err != nil {
				t.Fatalf("Failed CreateApiSpec(%v): %s", req, err.Error())
			}
			complexityCommand := Command()
			args := []string{"complexity", spec.Name}
			complexityCommand.SetArgs(args)
			if err = complexityCommand.Execute(); err != nil {
				t.Fatalf("Execute() with args %v returned error: %s", args, err)
			}
			specName, err := names.ParseSpec(spec.Name)
			if err != nil {
				t.Fatalf("Failed parsing spec name %s: %s", spec.Name, err)
			}
			contents, err := client.GetArtifactContents(ctx, &rpc.GetArtifactContentsRequest{
				Name: specName.Artifact("complexity").String(),
			})
			if err != nil {
				t.Fatalf("Failed getting artifact contents %s: %s", test.getPattern, err)
			}
			gotProto := &metrics.Complexity{}
			if err := proto.Unmarshal(contents.GetData(), gotProto); err != nil {
				t.Fatalf("Failed to unmarshal artifact: %s", err)
			}
			opts := cmp.Options{protocmp.Transform()}
			if !cmp.Equal(test.wantProto, gotProto, opts) {
				t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.wantProto, gotProto, opts))
			}
			err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name:  "projects/" + testProject,
				Force: true,
			})
			if err != nil && status.Code(err) != codes.NotFound {
				t.Fatalf("Setup: Failed to delete test project: %s", err)
			}
		})
	}
}
