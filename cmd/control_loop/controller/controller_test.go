package controller

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"log"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func unavailable(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Code() == codes.Unavailable
}

func check(t *testing.T, message string, err error) {
	if unavailable(err) {
		t.Logf("Unable to connect to registry server. Is it running?")
		t.FailNow()
	}
	if err != nil {
		t.Errorf(message, err.Error())
	}
}

func checkActions(expectedActions, actions []string) string {
	opts := cmpopts.SortSlices(func(a, b string) bool { return a < b })
	diff := cmp.Diff(expectedActions, actions, opts)
	return diff

}

func readAndGZipFile(filename string) (*bytes.Buffer, error) {
	fileBytes, err := ioutil.ReadFile(filename)
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	_, err = zw.Write(fileBytes)
	if err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return &buf, nil
}

func DeleteProject(ctx context.Context, client connection.Client, t *testing.T, project_id string) {
	req := &rpc.DeleteProjectRequest{
		Name: "projects/" + project_id,
	}
	err := client.DeleteProject(ctx, req)
	if status.Code(err) != codes.NotFound {
		check(t, "Failed to delete demo project: %+v", err)
	}
}

func CreateProject(ctx context.Context, client connection.Client, t *testing.T, project_id string) {
	req := &rpc.CreateProjectRequest{
		ProjectId: project_id,
		Project: &rpc.Project{
			DisplayName: "Demo",
			Description: "A demo catalog",
		},
	}
	project, err := client.CreateProject(ctx, req)
	check(t, "error creating project %s", err)
	if project.GetName() != "projects/" + project_id {
		t.Errorf("Invalid project name %s", project.GetName())
	}
}

func CreateApi(ctx context.Context, client connection.Client, t *testing.T, parent, api_id string) {
	req := &rpc.CreateApiRequest{
			Parent: parent,
			ApiId:  api_id,
			Api: &rpc.Api{
				DisplayName:  api_id,
				Description:  fmt.Sprintf("Sample Test API: %s", api_id),
				Availability: "GENERAL",
			},
		}
		_, err := client.CreateApi(ctx, req)
		check(t, "error creating api %s", err)
}
func CreateVersion(ctx context.Context, client connection.Client, t *testing.T, parent, version_id string) {
	req := &rpc.CreateApiVersionRequest{
			Parent:       parent,
			ApiVersionId: version_id,
			ApiVersion:   &rpc.ApiVersion{},
		}
		_, err := client.CreateApiVersion(ctx, req)
		check(t, "error creating version %s", err)
}

func UploadSpec(ctx context.Context, client connection.Client, t *testing.T, file_path, parent, spec_id, mime_type string) {
	buf, err := readAndGZipFile(file_path)
	check(t, "error reading spec", err)
	req := &rpc.CreateApiSpecRequest{
		Parent:    parent,
		ApiSpecId: spec_id,
		ApiSpec: &rpc.ApiSpec{
			MimeType: mime_type,
			Contents: buf.Bytes(),
		},
	}
	_, err = client.CreateApiSpec(ctx, req)
	check(t, "error creating spec %s", err)
}

func TestSingleComputeLint(t *testing.T) {
	// Setup: Single spec in the project
	// Expect: One single command to compute artifact
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()
	// Setup
	project_id := "controller-demo"
	api_id := "petstore"
	version_id := "1.0.0"
	spec_id := "openapi.yaml"
	mime_type := "application/x.openapi+gzip;version=3.0.0"
	DeleteProject(ctx, registryClient, t, project_id)
	CreateProject(ctx, registryClient, t, project_id)
	CreateApi(ctx, registryClient, t, "projects/" + project_id, api_id)
	CreateVersion(ctx, registryClient, t, fmt.Sprintf("projects/%s/apis/%s", project_id, api_id), version_id)
	UploadSpec(ctx, registryClient, t, "petstore/1.0.0/openapi.yaml@r0", fmt.Sprintf("projects/%s/apis/%s/versions/%s", project_id, api_id, version_id), spec_id, mime_type)

	// Test the manifest
	manifest, err := ReadManifest(
		"../test/manifest_1.yaml")
	if err != nil {
		t.Error(err.Error())
	}

	actions, err := ProcessManifest(manifest)
	if err != nil {
		log.Printf(err.Error())
	}
	expectedActions := []string{"compute lint projects/controller-demo/apis/petstore/versions/1.0.0/specs/openapi.yaml --linter gnostic"}
	if diff := checkActions(actions, expectedActions); diff != "" {
		t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
	}
}

func TestMultipleComputeLint(t *testing.T) {
	// Setup: 3 specs in project
	// Expect: Create artifact command from scratch for each spec

	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()
	// Setup
	project_id := "controller-demo"
	api_id := "petstore"
	version_id_1 := "1.0.0"
	version_id_2 := "1.0.1"
	version_id_3 := "1.1.0"
	spec_id := "openapi.yaml"
	mime_type := "application/x.openapi+gzip;version=3.0.0"
	DeleteProject(ctx, registryClient, t, project_id)
	CreateProject(ctx, registryClient, t, project_id)
	CreateApi(ctx, registryClient, t, fmt.Sprintf("projects/%s", project_id), api_id)
	// Version 1.0.0
	CreateVersion(ctx, registryClient, t, fmt.Sprintf("projects/%s/apis/%s", project_id, api_id), version_id_1)
	UploadSpec(ctx, registryClient, t, "petstore/1.0.0/openapi.yaml@r0", fmt.Sprintf("projects/%s/apis/%s/versions/%s", project_id, api_id, version_id_1), spec_id, mime_type)
	// Version 1.0.1
	CreateVersion(ctx, registryClient, t, fmt.Sprintf("projects/%s/apis/%s", project_id, api_id), version_id_2)
	UploadSpec(ctx, registryClient, t, "petstore/1.0.0/openapi.yaml@r0", fmt.Sprintf("projects/%s/apis/%s/versions/%s", project_id, api_id, version_id_2), spec_id, mime_type)
	// Version 1.1.0
	CreateVersion(ctx, registryClient, t, fmt.Sprintf("projects/%s/apis/%s", project_id, api_id), version_id_3)
	UploadSpec(ctx, registryClient, t, "petstore/1.0.0/openapi.yaml@r0", fmt.Sprintf("projects/%s/apis/%s/versions/%s", project_id, api_id, version_id_3), spec_id, mime_type)

	// Test the manifest
	manifest, err := ReadManifest(
		"../test/manifest_1.yaml")
	if err != nil {
		t.Error(err.Error())
	}
	
	actions, err := ProcessManifest(manifest)
	if err != nil {
		log.Printf(err.Error())
	}
	expectedActions := []string{
	"compute lint projects/controller-demo/apis/petstore/versions/1.0.0/specs/openapi.yaml --linter gnostic",
	"compute lint projects/controller-demo/apis/petstore/versions/1.0.1/specs/openapi.yaml --linter gnostic",
	"compute lint projects/controller-demo/apis/petstore/versions/1.1.0/specs/openapi.yaml --linter gnostic"}
	if diff := checkActions(expectedActions, actions); diff != "" {
		t.Errorf("ProcessManifest(%+v) returned unexpected diff (-want +got):\n%s", manifest, diff)
	}

}

//TODO: Add tests for already existing artifacts



