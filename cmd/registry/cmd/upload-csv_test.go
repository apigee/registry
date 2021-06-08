// Copyright 2021 Google LLC. All Rights Reserved.
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

package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/tests/seeding/fileseed"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
)

const (
	gzipOpenAPIv3 = "application/x.openapi+gzip;version=3.0.0"
	testProject   = "googleapis.com"
)

var (
	specContents1 = []byte(`{"openapi": "3.0.0", "info": {"title": "First Spec", "version": "v1"}, "paths": {}}`)
	specContents2 = []byte(`{"openapi": "3.0.0", "info": {"title": "Second Spec", "version": "v1"}, "paths": {}}`)
	specContents3 = []byte(`{"openapi": "3.0.0", "info": {"title": "Third Spec", "version": "v1"}, "paths": {}}`)
	specContents4 = []byte(`{"openapi": "3.0.0", "info": {"title": "Fourth Spec", "version": "v1"}, "paths": {}}`)
)

func TestUploadCSV(t *testing.T) {
	root := t.TempDir()
	tests := []struct {
		desc  string
		files []fileseed.File
		args  []string
		want  []*rpc.ApiSpec
		opts  cmp.Option
	}{
		{
			desc: "multiple spec upload",
			files: []fileseed.File{
				{
					Path:     fmt.Sprintf("%s/cloudtasks/v2beta2/openapi.yaml", root),
					Contents: specContents1,
				},
				{
					Path:     fmt.Sprintf("%s/cloudtasks/v2/openapi.yaml", root),
					Contents: specContents2,
				},
				{
					Path:     fmt.Sprintf("%s/datastore/v1beta1/openapi.yaml", root),
					Contents: specContents3,
				},
				{
					Path:     fmt.Sprintf("%s/datastore/v1/openapi.yaml", root),
					Contents: specContents4,
				},
				{
					Path: fmt.Sprintf("%s/specs.csv", root),
					Contents: []byte(strings.Join([]string{
						"api_id,version_id,spec_id,filepath",
						fmt.Sprintf("cloudtasks,v2beta2,openapi.yaml,%s/cloudtasks/v2beta2/openapi.yaml", root),
						fmt.Sprintf("cloudtasks,v2,openapi.yaml,%s/cloudtasks/v2/openapi.yaml", root),
						fmt.Sprintf("datastore,v1beta1,openapi.yaml,%s/datastore/v1beta1/openapi.yaml", root),
						fmt.Sprintf("datastore,v1,openapi.yaml,%s/datastore/v1/openapi.yaml", root),
					}, "\n")),
				},
			},
			args: []string{fmt.Sprintf("%s/specs.csv", root), "--project_id", testProject},
			want: []*rpc.ApiSpec{
				{
					Name:     "projects/googleapis.com/apis/cloudtasks/versions/v2beta2/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
					Contents: specContents1,
				},
				{
					Name:     "projects/googleapis.com/apis/cloudtasks/versions/v2/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
					Contents: specContents2,
				},
				{
					Name:     "projects/googleapis.com/apis/datastore/versions/v1beta1/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
					Contents: specContents3,
				},
				{
					Name:     "projects/googleapis.com/apis/datastore/versions/v1/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
					Contents: specContents4,
				},
			},
		},
		{
			desc: "out of order columns",
			files: []fileseed.File{
				{
					Path:     fmt.Sprintf("%s/openapi.yaml", root),
					Contents: specContents1,
				},
				{
					Path: fmt.Sprintf("%s/specs.csv", root),
					Contents: []byte(strings.Join([]string{
						"filepath,spec_id,version_id,api_id",
						fmt.Sprintf("%s/openapi.yaml,openapi.yaml,v2,cloudtasks", root),
					}, "\n")),
				},
			},
			args: []string{fmt.Sprintf("%s/specs.csv", root), "--project_id", testProject},
			want: []*rpc.ApiSpec{
				{
					Name:     "projects/googleapis.com/apis/cloudtasks/versions/v2/specs/openapi.yaml",
					MimeType: gzipOpenAPIv3,
					Contents: specContents1,
				},
			},
		},
		{
			desc: "empty sheet",
			files: []fileseed.File{
				{
					Path:     fmt.Sprintf("%s/specs.csv", root),
					Contents: []byte("api_id,version_id,spec_id,filepath"),
				},
			},
			args: []string{fmt.Sprintf("%s/specs.csv", root), "--project_id", testProject},
			want: []*rpc.ApiSpec{},
		},
		{
			desc: "empty ID fields",
			files: []fileseed.File{
				{
					Path:     fmt.Sprintf("%s/openapi.yaml", root),
					Contents: specContents1,
				},
				{
					Path: fmt.Sprintf("%s/specs.csv", root),
					Contents: []byte(strings.Join([]string{
						"api_id,version_id,spec_id,filepath",
						fmt.Sprintf(",,,%s/openapi.yaml", root),
					}, "\n")),
				},
			},
			args: []string{fmt.Sprintf("%s/specs.csv", root), "--project_id", testProject},
			want: []*rpc.ApiSpec{
				{
					MimeType: gzipOpenAPIv3,
					Contents: specContents1,
				},
			},
			// Ignore the name since it's randomly generated.
			opts: protocmp.IgnoreFields(&rpc.ApiSpec{}, "name"),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			client, err := connection.NewClient(ctx)
			if err != nil {
				t.Fatalf("Setup: Failed to create client: %s", err)
			}

			err = client.DeleteProject(ctx, &rpc.DeleteProjectRequest{
				Name: "projects/" + testProject,
			})
			if err != nil && status.Code(err) != codes.NotFound {
				t.Fatalf("Setup: Failed to delete test project: %s", err)
			}

			if err := fileseed.Write(test.files...); err != nil {
				t.Fatalf("Setup: Failed to write test files: %s", err)
			}

			args := append([]string{"upload", "csv"}, test.args...)
			rootCmd.SetArgs(args)
			if err := uploadCsvCmd.Execute(); err != nil {
				t.Fatalf("Execute() with args %v returned error: %s", args, err)
			}

			it := client.ListApiSpecs(ctx, &rpc.ListApiSpecsRequest{
				Parent:   fmt.Sprintf("projects/%s/apis/-/versions/-", testProject),
				PageSize: int32(len(test.want)),
			})

			got := make([]*rpc.ApiSpec, 0, len(test.want))
			for spec, err := it.Next(); err != iterator.Done; spec, err = it.Next() {
				if err != nil {
					t.Fatalf("Failed to read spec from server: %s", err)
				}

				body, err := client.GetApiSpecContents(ctx, &rpc.GetApiSpecContentsRequest{
					Name: fmt.Sprintf("%s/contents", spec.GetName()),
				})
				if err != nil {
					t.Fatalf("GetApiSpecContents(%q) returned error: %s", spec.GetName(), err)
				}

				spec.Contents = body.GetData()
				got = append(got, spec)
			}

			opts := cmp.Options{
				test.opts,
				protocmp.Transform(),
				// Ignore list ordering. We only want to verify that each spec exists.
				cmpopts.SortSlices(func(a, b *rpc.ApiSpec) bool { return a.GetName() < b.GetName() }),
				// Ignore randomly generated fields.
				protocmp.IgnoreFields(&rpc.ApiSpec{}, "revision_id", "hash", "size_bytes", "create_time", "revision_create_time", "revision_update_time"),
			}

			if diff := cmp.Diff(test.want, got, opts); diff != "" {
				t.Errorf("ListApiSpecs returned unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}
