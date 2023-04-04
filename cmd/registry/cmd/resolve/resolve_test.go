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

package resolve

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/apply"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func readAndGZipFile(t *testing.T, filename string) (*bytes.Buffer, error) {
	t.Helper()
	fileBytes, _ := os.ReadFile(filename)
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	_, err := zw.Write(fileBytes)
	if err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return &buf, nil
}

func TestResolve(t *testing.T) {
	tests := []struct {
		desc         string
		manifestPath string
		dryRun       bool
		listParent   string
		want         []string
	}{
		{
			desc:         "normal case",
			manifestPath: filepath.Join("testdata", "manifest.yaml"),
			dryRun:       false,
			listParent:   "projects/controller-demo/locations/global/apis/petstore/versions/-/specs/-",
			want: []string{
				"projects/controller-demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/complexity",
				"projects/controller-demo/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/complexity",
				"projects/controller-demo/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/complexity",
			},
		},
		{
			desc:         "receipt artifact",
			manifestPath: filepath.Join("testdata", "manifest_receipt.yaml"),
			dryRun:       false,
			listParent:   "projects/controller-demo/locations/global/apis/petstore/versions/-/specs/-",
			want: []string{
				"projects/controller-demo/locations/global/apis/petstore/versions/1.0.0/specs/openapi/artifacts/test-receipt-artifact",
				"projects/controller-demo/locations/global/apis/petstore/versions/1.0.1/specs/openapi/artifacts/test-receipt-artifact",
				"projects/controller-demo/locations/global/apis/petstore/versions/1.1.0/specs/openapi/artifacts/test-receipt-artifact",
			},
		},
		{
			desc:         "dry run",
			manifestPath: filepath.Join("testdata", "manifest.yaml"),
			dryRun:       true,
			listParent:   "projects/controller-demo/locations/global/apis/petstore/versions/-/specs/-",
			want:         []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			testProject := "controller-demo"
			client, adminClient := grpctest.SetupRegistry(ctx, t, testProject, nil)

			// Setup some resources in the project

			// Create API
			api, err := client.CreateApi(ctx, &rpc.CreateApiRequest{
				Parent: "projects/" + testProject + "/locations/global",
				ApiId:  "petstore",
				Api: &rpc.Api{
					DisplayName:  "petstore",
					Description:  "Sample Test API",
					Availability: "GENERAL",
				},
			})
			if err != nil {
				t.Fatalf("Failed CreateApi %s: %s", "petstore", err.Error())
			}

			// Create Versions 1.0.0, 1.0.1, 1.1.0
			v1, err := client.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
				Parent:       api.Name,
				ApiVersionId: "1.0.0",
				ApiVersion:   &rpc.ApiVersion{},
			})
			if err != nil {
				t.Fatalf("Failed CreateVersion 1.0.0: %s", err.Error())
			}

			v2, err := client.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
				Parent:       api.Name,
				ApiVersionId: "1.0.1",
				ApiVersion:   &rpc.ApiVersion{},
			})
			if err != nil {
				t.Fatalf("Failed CreateVersion 1.0.1: %s", err.Error())
			}

			v3, err := client.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
				Parent:       api.Name,
				ApiVersionId: "1.1.0",
				ApiVersion:   &rpc.ApiVersion{},
			})
			if err != nil {
				t.Fatalf("Failed CreateVersion 1.1.0: %s", err.Error())
			}

			// Create Spec in each of the versions
			buf, err := readAndGZipFile(t, filepath.Join("testdata", "openapi.yaml"))
			if err != nil {
				t.Fatalf("Failed reading API contents: %s", err.Error())
			}

			req := &rpc.CreateApiSpecRequest{
				Parent:    v1.Name,
				ApiSpecId: "openapi",
				ApiSpec: &rpc.ApiSpec{
					MimeType: "application/x.openapi+gzip;version=3.0.0",
					Contents: buf.Bytes(),
				},
			}
			s, err := client.CreateApiSpec(ctx, &rpc.CreateApiSpecRequest{
				Parent:    v1.Name,
				ApiSpecId: "openapi",
				ApiSpec: &rpc.ApiSpec{
					MimeType: "application/x.openapi+gzip;version=3.0.0",
					Contents: buf.Bytes(),
				},
			})
			if err != nil {
				t.Fatalf("Failed CreateApiSpec(%v): %s", req, err.Error())
			}

			addRevisionToWants := func(s *rpc.ApiSpec) {
				t.Helper()
				n, err := names.ParseSpec(s.Name)
				if err != nil {
					t.Fatal("ParseSpec", err)
				}
				specVer := fmt.Sprintf("%s/specs/%s", n.VersionID, n.SpecID)
				specRev := fmt.Sprintf("%s/specs/%s@%s", n.VersionID, n.SpecID, s.RevisionId)
				for i := range test.want { // inject dynamic spec revisions
					test.want[i] = strings.ReplaceAll(test.want[i], specVer, specRev)
				}
			}
			addRevisionToWants(s)

			req = &rpc.CreateApiSpecRequest{
				Parent:    v2.Name,
				ApiSpecId: "openapi",
				ApiSpec: &rpc.ApiSpec{
					MimeType: "application/x.openapi+gzip;version=3.0.0",
					Contents: buf.Bytes(),
				},
			}
			s, err = client.CreateApiSpec(ctx, &rpc.CreateApiSpecRequest{
				Parent:    v2.Name,
				ApiSpecId: "openapi",
				ApiSpec: &rpc.ApiSpec{
					MimeType: "application/x.openapi+gzip;version=3.0.0",
					Contents: buf.Bytes(),
				},
			})
			if err != nil {
				t.Fatalf("Failed CreateApiSpec(%v): %s", req, err.Error())
			}
			addRevisionToWants(s)

			req = &rpc.CreateApiSpecRequest{
				Parent:    v3.Name,
				ApiSpecId: "openapi",
				ApiSpec: &rpc.ApiSpec{
					MimeType: "application/x.openapi+gzip;version=3.0.0",
					Contents: buf.Bytes(),
				},
			}
			s, err = client.CreateApiSpec(ctx, &rpc.CreateApiSpecRequest{
				Parent:    v3.Name,
				ApiSpecId: "openapi",
				ApiSpec: &rpc.ApiSpec{
					MimeType: "application/x.openapi+gzip;version=3.0.0",
					Contents: buf.Bytes(),
				},
			})
			if err != nil {
				t.Fatalf("Failed CreateApiSpec(%v): %s", req, err.Error())
			}
			addRevisionToWants(s)

			// Apply the manifest to the registry
			args := []string{"-f", test.manifestPath, "--parent", "projects/" + testProject + "/locations/global"}
			applyCmd := apply.Command()
			applyCmd.SetArgs(args)
			if err = applyCmd.Execute(); err != nil {
				t.Fatalf("Failed to apply the manifest: %s", err)
			}

			resolveCmd := Command()

			args = []string{"projects/" + testProject + "/locations/global/artifacts/test-manifest"}
			if test.dryRun {
				args = append(args, "--dry-run")
			}
			resolveCmd.SetArgs(args)

			if err = resolveCmd.Execute(); err != nil {
				t.Fatalf("Execute() with args %v returned error: %s", args, err)
			}

			// List all the artifacts
			it := client.ListArtifacts(ctx, &rpc.ListArtifactsRequest{
				Parent: test.listParent,
			})

			got := make([]string, 0)
			for artifact, err := it.Next(); err != iterator.Done; artifact, err = it.Next() {
				if err != nil {
					continue // TODO: Handle errors.
				}

				got = append(got, artifact.Name)
			}

			sortStrings := cmpopts.SortSlices(func(a, b string) bool { return a < b })
			if diff := cmp.Diff(test.want, got, sortStrings); diff != "" {
				t.Errorf("Returned unexpected diff (-want +got):\n%s", diff)
			}

			// Delete the demo project
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
