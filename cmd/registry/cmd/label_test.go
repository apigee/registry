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
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLabel(t *testing.T) {
	var err error

	projectId := "label-test"
	projectName := "projects/" + projectId
	apiId := "sample"
	apiName := projectName + "/apis/" + apiId
	versionId := "1.0.0"
	versionName := apiName + "/versions/" + versionId
	specId := "openapi.json"
	specName := versionName + "/specs/" + specId

	// Create a registry client.
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Fatalf("error creating client: %+v", err)
	}
	defer registryClient.Close()
	// Clear the test project.
	err = registryClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name: projectName,
	})
	if err != nil && status.Code(err) != codes.NotFound {
		t.Fatalf("error deleting test project: %+v", err)
	}
	// Create the test project.
	_, err = registryClient.CreateProject(ctx, &rpc.CreateProjectRequest{
		ProjectId: projectId,
		Project: &rpc.Project{
			DisplayName: "Test",
			Description: "A test catalog",
		},
	})
	if err != nil {
		t.Fatalf("error creating project %s", err)
	}
	// Create a sample api.
	_, err = registryClient.CreateApi(ctx, &rpc.CreateApiRequest{
		Parent: projectName,
		ApiId:  apiId,
		Api:    &rpc.Api{},
	})
	if err != nil {
		t.Fatalf("error creating api %s", err)
	}
	// Create a sample version.
	_, err = registryClient.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
		Parent:       apiName,
		ApiVersionId: versionId,
		ApiVersion:   &rpc.ApiVersion{},
	})
	if err != nil {
		t.Fatalf("error creating version %s", err)
	}
	// Create a sample spec.
	_, err = registryClient.CreateApiSpec(ctx, &rpc.CreateApiSpecRequest{
		Parent:    versionName,
		ApiSpecId: specId,
		ApiSpec: &rpc.ApiSpec{
			MimeType: "application/x.openapi;version=3.0.0",
			Contents: []byte(`{"openapi": "3.0.0", "info": {"title": "test", "version": "v1"}, "paths": {}}`),
		},
	})
	if err != nil {
		t.Fatalf("error creating spec %s", err)
	}

	// Add some labels to the test api.
	rootCmd.SetArgs([]string{"label", apiName, "a=1", "b=2"})
	if err := labelCmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %v returned error: %s", rootCmd.Args, err)
	}
	api, err := registryClient.GetApi(ctx, &rpc.GetApiRequest{
		Name: apiName,
	})
	if diff := cmp.Diff(api.Labels, map[string]string{"a": "1", "b": "2"}); diff != "" {
		t.Errorf("labels incorrectly set %+v", api.Labels)
	}
	// Remove one label and overwrite the other.
	rootCmd.SetArgs([]string{"label", apiName, "a=3", "b-", "--overwrite"})
	if err = labelCmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %v returned error: %s", rootCmd.Args, err)
	}
	api, err = registryClient.GetApi(ctx, &rpc.GetApiRequest{
		Name: apiName,
	})
	if diff := cmp.Diff(api.Labels, map[string]string{"a": "3"}); diff != "" {
		t.Errorf("labels incorrectly set %+v", api.Labels)
	}
	// Changing a label without --overwrite should be ignored.
	rootCmd.SetArgs([]string{"label", apiName, "a=4"})
	if err = labelCmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %v returned error: %s", rootCmd.Args, err)
	}
	if diff := cmp.Diff(api.Labels, map[string]string{"a": "3"}); diff != "" {
		t.Errorf("labels incorrectly set %+v", api.Labels)
	}

	// Add some labels to the test version.
	rootCmd.SetArgs([]string{"label", versionName, "a=1", "b=2"})
	if err := labelCmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %v returned error: %s", rootCmd.Args, err)
	}
	version, err := registryClient.GetApiVersion(ctx, &rpc.GetApiVersionRequest{
		Name: versionName,
	})
	if diff := cmp.Diff(version.Labels, map[string]string{"a": "1", "b": "2"}); diff != "" {
		t.Errorf("labels incorrectly set %+v", version.Labels)
	}
	// Remove one label and overwrite the other.
	rootCmd.SetArgs([]string{"label", versionName, "a=3", "b-", "--overwrite"})
	if err = labelCmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %v returned error: %s", rootCmd.Args, err)
	}
	version, err = registryClient.GetApiVersion(ctx, &rpc.GetApiVersionRequest{
		Name: versionName,
	})
	if diff := cmp.Diff(version.Labels, map[string]string{"a": "3"}); diff != "" {
		t.Errorf("labels incorrectly set %+v", version.Labels)
	}
	// Changing a label without --overwrite should be ignored.
	rootCmd.SetArgs([]string{"label", versionName, "a=4"})
	if err = labelCmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %v returned error: %s", rootCmd.Args, err)
	}
	if diff := cmp.Diff(version.Labels, map[string]string{"a": "3"}); diff != "" {
		t.Errorf("labels incorrectly set %+v", version.Labels)
	}

	// Add some labels to the test spec.
	rootCmd.SetArgs([]string{"label", specName, "a=1", "b=2"})
	if err := labelCmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %v returned error: %s", rootCmd.Args, err)
	}
	spec, err := registryClient.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
		Name: specName,
	})
	if diff := cmp.Diff(spec.Labels, map[string]string{"a": "1", "b": "2"}); diff != "" {
		t.Errorf("labels incorrectly set %+v", spec.Labels)
	}
	// Remove one label and overwrite the other.
	rootCmd.SetArgs([]string{"label", specName, "a=3", "b-", "--overwrite"})
	if err = labelCmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %v returned error: %s", rootCmd.Args, err)
	}
	spec, err = registryClient.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
		Name: specName,
	})
	if diff := cmp.Diff(spec.Labels, map[string]string{"a": "3"}); diff != "" {
		t.Errorf("labels incorrectly set %+v", spec.Labels)
	}
	// Changing a label without --overwrite should be ignored.
	rootCmd.SetArgs([]string{"label", specName, "a=4"})
	if err = labelCmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %v returned error: %s", rootCmd.Args, err)
	}
	if diff := cmp.Diff(spec.Labels, map[string]string{"a": "3"}); diff != "" {
		t.Errorf("labels incorrectly set %+v", spec.Labels)
	}

	// Delete the test project.
	{
		req := &rpc.DeleteProjectRequest{
			Name: projectName,
		}
		err = registryClient.DeleteProject(ctx, req)
		if err != nil {
			t.Fatalf("failed to delete test project: %s", err)
		}
	}
}
