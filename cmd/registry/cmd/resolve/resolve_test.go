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
	"io/ioutil"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/upload"
	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/names"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func readAndGZipFile(t *testing.T, filename string) (*bytes.Buffer, error) {
	t.Helper()
	fileBytes, _ := ioutil.ReadFile(filename)
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
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		t.Fatalf("Setup: Failed to create client: %s", err)
	}

	testProject := "controller-demo"

	err = client.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name: "projects/" + testProject,
	})
	if err != nil && status.Code(err) != codes.NotFound {
		t.Fatalf("Setup: Failed to delete test project: %s", err)
	}

	project, err := client.CreateProject(ctx, &rpc.CreateProjectRequest{
		ProjectId: testProject,
		Project: &rpc.Project{
			DisplayName: "Demo",
			Description: "A demo catalog",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create project %s: %s", testProject, err.Error())
	}

	// Setup some resources in the project

	// Create API
	api, err := client.CreateApi(ctx, &rpc.CreateApiRequest{
		Parent: project.Name,
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
	buf, err := readAndGZipFile(t, "testdata/openapi.yaml")
	if err != nil {
		t.Fatalf("Failed reading API contents: %s", err.Error())
	}

	req := &rpc.CreateApiSpecRequest{
		Parent:    v1.Name,
		ApiSpecId: "openapi.yaml",
		ApiSpec: &rpc.ApiSpec{
			MimeType: "application/x.openapi+gzip;version=3.0.0",
			Contents: buf.Bytes(),
		},
	}
	v1spec, err := client.CreateApiSpec(ctx, &rpc.CreateApiSpecRequest{
		Parent:    v1.Name,
		ApiSpecId: "openapi.yaml",
		ApiSpec: &rpc.ApiSpec{
			MimeType: "application/x.openapi+gzip;version=3.0.0",
			Contents: buf.Bytes(),
		},
	})
	if err != nil {
		t.Fatalf("Failed CreateApiSpec(%v): %s", req, err.Error())
	}

	req = &rpc.CreateApiSpecRequest{
		Parent:    v2.Name,
		ApiSpecId: "openapi.yaml",
		ApiSpec: &rpc.ApiSpec{
			MimeType: "application/x.openapi+gzip;version=3.0.0",
			Contents: buf.Bytes(),
		},
	}
	v2spec, err := client.CreateApiSpec(ctx, &rpc.CreateApiSpecRequest{
		Parent:    v2.Name,
		ApiSpecId: "openapi.yaml",
		ApiSpec: &rpc.ApiSpec{
			MimeType: "application/x.openapi+gzip;version=3.0.0",
			Contents: buf.Bytes(),
		},
	})
	if err != nil {
		t.Fatalf("Failed CreateApiSpec(%v): %s", req, err.Error())
	}

	req = &rpc.CreateApiSpecRequest{
		Parent:    v3.Name,
		ApiSpecId: "openapi.yaml",
		ApiSpec: &rpc.ApiSpec{
			MimeType: "application/x.openapi+gzip;version=3.0.0",
			Contents: buf.Bytes(),
		},
	}
	v3spec, err := client.CreateApiSpec(ctx, &rpc.CreateApiSpecRequest{
		Parent:    v3.Name,
		ApiSpecId: "openapi.yaml",
		ApiSpec: &rpc.ApiSpec{
			MimeType: "application/x.openapi+gzip;version=3.0.0",
			Contents: buf.Bytes(),
		},
	})
	if err != nil {
		t.Fatalf("Failed CreateApiSpec(%v): %s", req, err.Error())
	}

	// Upload the manifest to registry
	args := []string{"manifest", "testdata/manifest.yaml", "--project_id=" + testProject}
	uploadCmd := upload.Command(ctx)
	uploadCmd.SetArgs(args)
	if err = uploadCmd.Execute(); err != nil {
		t.Fatalf("Failed to upload the manifest: %s", err)
	}

	// Call the controller update command
	resolveCmd := Command(ctx)
	args = []string{"projects/" + testProject + "/artifacts/test-manifest"}
	resolveCmd.SetArgs(args)
	if err = resolveCmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %v returned error: %s", args, err)
	}

	// List all the artifacts
	got := make([]string, 0)
	listPattern := "projects/controller-demo/apis/petstore/versions/-/specs/-/artifacts/-"
	segments := names.ArtifactRegexp().FindStringSubmatch(listPattern)
	_ = core.ListArtifacts(ctx, client, segments, "", false,
		func(artifact *rpc.Artifact) {
			got = append(got, artifact.Name)
		},
	)

	want := []string{
		v1spec.Name + "/artifacts/complexity",
		v2spec.Name + "/artifacts/complexity",
		v3spec.Name + "/artifacts/complexity",
	}

	sortStrings := cmpopts.SortSlices(func(a, b string) bool { return a < b })
	if diff := cmp.Diff(want, got, sortStrings); diff != "" {
		t.Errorf("Returned unexpected diff (-want +got):\n%s", diff)
	}

	// Delete the demo project
	err = client.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name: "projects/" + testProject,
	})
	if err != nil && status.Code(err) != codes.NotFound {
		t.Fatalf("Setup: Failed to delete test project: %s", err)
	}

}
