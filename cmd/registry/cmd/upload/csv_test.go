// Copyright 2021 Google LLC.
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

package upload

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

const (
	gzipOpenAPIv2 = "application/x.openapi+gzip;version=2.0"
	gzipOpenAPIv3 = "application/x.openapi+gzip;version=3.0.0"
)

func TestUploadCSV(t *testing.T) {
	cloudtasksGA, err := os.ReadFile(filepath.Join("testdata", "csv", "cloudtasks", "v2", "openapi.yaml"))
	if err != nil {
		t.Fatalf("Setup: Failed to read spec contents: %s", err)
	}

	cloudtasksBeta, err := os.ReadFile(filepath.Join("testdata", "csv", "cloudtasks", "v2beta2", "openapi.yaml"))
	if err != nil {
		t.Fatalf("Setup: Failed to read spec contents: %s", err)
	}

	datastoreGA, err := os.ReadFile(filepath.Join("testdata", "csv", "datastore", "v1", "openapi.yaml"))
	if err != nil {
		t.Fatalf("Setup: Failed to read spec contents: %s", err)
	}

	datastoreBeta, err := os.ReadFile(filepath.Join("testdata", "csv", "datastore", "v1beta1", "openapi.yaml"))
	if err != nil {
		t.Fatalf("Setup: Failed to read spec contents: %s", err)
	}

	swagger, err := os.ReadFile(filepath.Join("testdata", "openapi", "petstore", "2.0", "swagger.yaml"))
	if err != nil {
		t.Fatalf("Setup: Failed to read spec contents: %s", err)
	}

	const testProject = "csv-demo"
	const testParent = "projects/" + testProject + "/locations/global"
	tests := []struct {
		desc string
		args []string
		want []*rpc.ApiSpec
	}{
		{
			desc: "multiple spec upload",
			args: []string{
				filepath.Join("testdata", "csv", "multiple-specs.csv"),
				"--parent", testParent,
			},
			want: []*rpc.ApiSpec{
				{
					Name:     fmt.Sprintf("projects/%s/locations/global/apis/cloudtasks/versions/v2beta2/specs/openapi", testProject),
					MimeType: gzipOpenAPIv3,
					Contents: cloudtasksBeta,
				},
				{
					Name:     fmt.Sprintf("projects/%s/locations/global/apis/cloudtasks/versions/v2/specs/openapi", testProject),
					MimeType: gzipOpenAPIv3,
					Contents: cloudtasksGA,
				},
				{
					Name:     fmt.Sprintf("projects/%s/locations/global/apis/datastore/versions/v1beta1/specs/openapi", testProject),
					MimeType: gzipOpenAPIv3,
					Contents: datastoreBeta,
				},
				{
					Name:     fmt.Sprintf("projects/%s/locations/global/apis/datastore/versions/v1/specs/openapi", testProject),
					MimeType: gzipOpenAPIv3,
					Contents: datastoreGA,
				},
				{
					Name:     fmt.Sprintf("projects/%s/locations/global/apis/swagger/versions/v1/specs/openapi", testProject),
					MimeType: gzipOpenAPIv2,
					Contents: swagger,
				},
			},
		},
		{
			desc: "out of order columns",
			args: []string{
				filepath.Join("testdata", "csv", "out-of-order-columns.csv"),
				"--parent", testParent,
			},
			want: []*rpc.ApiSpec{
				{
					Name:     fmt.Sprintf("projects/%s/locations/global/apis/cloudtasks/versions/v2/specs/openapi", testProject),
					MimeType: gzipOpenAPIv3,
					Contents: cloudtasksGA,
				},
			},
		},
		{
			desc: "empty sheet",
			args: []string{
				filepath.Join("testdata", "csv", "empty-sheet.csv"),
				"--parent", testParent,
			},
			want: []*rpc.ApiSpec{},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			client, _ := grpctest.SetupRegistry(ctx, t, testProject, nil)

			args := append([]string{"csv"}, test.args...)

			cmd := Command()
			cmd.SetArgs(args)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() with args %v returned error: %s", args, err)
			}

			it := client.ListApiSpecs(ctx, &rpc.ListApiSpecsRequest{
				Parent:   fmt.Sprintf("projects/%s/locations/global/apis/-/versions/-", testProject),
				PageSize: int32(len(test.want)),
			})

			got := make([]*rpc.ApiSpec, 0, len(test.want))
			for spec, err := it.Next(); err != iterator.Done; spec, err = it.Next() {
				if err != nil {
					t.Fatalf("Failed to read spec from server: %s", err)
				}

				body, err := client.GetApiSpecContents(ctx, &rpc.GetApiSpecContentsRequest{
					Name: spec.GetName(),
				})
				if err != nil {
					t.Fatalf("GetApiSpecContents(%q) returned error: %s", spec.GetName(), err)
				}

				spec.Contents = body.GetData()
				got = append(got, spec)
			}

			opts := cmp.Options{
				protocmp.Transform(),
				// Ignore list ordering. We only want to verify that each spec exists.
				cmpopts.SortSlices(func(a, b *rpc.ApiSpec) bool { return a.GetName() < b.GetName() }),
				// Ignore generated fields.
				protocmp.IgnoreFields(&rpc.ApiSpec{}, "revision_id", "hash", "size_bytes", "create_time", "revision_create_time", "revision_update_time"),
			}

			if diff := cmp.Diff(test.want, got, opts); diff != "" {
				t.Errorf("ListApiSpecs returned unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUploadCSVErrors(t *testing.T) {
	const testProject = "csv-errors"
	const testParent = "projects/" + testProject + "/locations/global"
	tests := []struct {
		desc string
		args []string
	}{
		{
			desc: "invalid parent",
			args: []string{
				filepath.Join("testdata", "csv", "multiple-specs.csv"),
				"--parent", "invalid",
			},
		},
		{
			desc: "missing parent",
			args: []string{
				filepath.Join("testdata", "csv", "out-of-order-columns.csv"),
				"--parent", "projects/missing/locations/global",
			},
		},
		{
			desc: "missing input file",
			args: []string{
				filepath.Join("testdata", "csv", "missing.csv"),
				"--parent", testParent,
			},
		},
	}

	ctx := context.Background()
	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Setup: Failed to create client: %s", err)
	}
	err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  "projects/" + testProject,
		Force: true,
	})
	if err != nil && status.Code(err) != codes.NotFound {
		t.Fatalf("Setup: Failed to delete test project: %s", err)
	}
	_, err = adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
		ProjectId: testProject,
		Project: &rpc.Project{
			DisplayName: "Test",
			Description: "A test catalog",
		},
	})
	if err != nil {
		t.Fatalf("Error creating project %s", err)
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			args := append([]string{"csv"}, test.args...)
			cmd := Command()
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			cmd.SetArgs(args)
			if err := cmd.Execute(); err == nil {
				t.Fatalf("Execute() with args %v succeeded and should have failed", args)
			}
		})
	}
}
