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

package annotate

import (
	"context"
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAnnotate(t *testing.T) {
	const (
		projectID      = "annotate-test"
		projectName    = "projects/" + projectID
		apiID          = "sample"
		apiName        = projectName + "/locations/global/apis/" + apiID
		versionID      = "1.0.0"
		versionName    = apiName + "/versions/" + versionID
		specID         = "openapi.json"
		specName       = versionName + "/specs/" + specID
		deploymentId   = "deployment1"
		deploymentName = apiName + "/deployments/" + deploymentId
	)

	// Create a registry client.
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Fatalf("Error creating client: %+v", err)
	}
	defer registryClient.Close()
	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Error creating client: %+v", err)
	}
	defer adminClient.Close()
	// Clear the test project.
	err = adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{
		Name:  projectName,
		Force: true,
	})
	if err != nil && status.Code(err) != codes.NotFound {
		t.Fatalf("Error deleting test project: %+v", err)
	}
	// Create the test project.
	_, err = adminClient.CreateProject(ctx, &rpc.CreateProjectRequest{
		ProjectId: projectID,
		Project: &rpc.Project{
			DisplayName: "Test",
			Description: "A test catalog",
		},
	})
	if err != nil {
		t.Fatalf("Error creating project %s", err)
	}
	// Create a sample api.
	_, err = registryClient.CreateApi(ctx, &rpc.CreateApiRequest{
		Parent: projectName + "/locations/global",
		ApiId:  apiID,
		Api:    &rpc.Api{},
	})
	if err != nil {
		t.Fatalf("Error creating api %s", err)
	}
	// Create a sample version.
	_, err = registryClient.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
		Parent:       apiName,
		ApiVersionId: versionID,
		ApiVersion:   &rpc.ApiVersion{},
	})
	if err != nil {
		t.Fatalf("Error creating version %s", err)
	}
	// Create a sample spec.
	_, err = registryClient.CreateApiSpec(ctx, &rpc.CreateApiSpecRequest{
		Parent:    versionName,
		ApiSpecId: specID,
		ApiSpec: &rpc.ApiSpec{
			MimeType: "application/x.openapi;version=3.0.0",
			Contents: []byte(`{"openapi": "3.0.0", "info": {"title": "test", "version": "v1"}, "paths": {}}`),
		},
	})
	if err != nil {
		t.Fatalf("Error creating spec %s", err)
	}
	// Create a sample deployment.
	_, err = registryClient.CreateApiDeployment(ctx, &rpc.CreateApiDeploymentRequest{
		Parent:          apiName,
		ApiDeploymentId: deploymentId,
		ApiDeployment:   &rpc.ApiDeployment{},
	})
	if err != nil {
		t.Fatalf("Error creating deployment %s", err)
	}
	testCases := []struct {
		comment  string
		args     []string
		expected map[string]string
	}{
		{comment: "add some annotations",
			args:     []string{"a=1", "b=2"},
			expected: map[string]string{"a": "1", "b": "2"}},
		{comment: "remove one annotation and overwrite the other",
			args:     []string{"a=3", "b-", "--overwrite"},
			expected: map[string]string{"a": "3"}},
		{comment: "changing an annotation without --overwrite should be ignored",
			args:     []string{"a=4"},
			expected: map[string]string{"a": "3"}},
	}
	// test annotations for APIs.
	for _, tc := range testCases {
		cmd := Command(ctx)
		cmd.SetArgs(append([]string{apiName}, tc.args...))
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", tc.args, err)
		}
		api, err := registryClient.GetApi(ctx, &rpc.GetApiRequest{
			Name: apiName,
		})
		if err != nil {
			t.Errorf("Error getting api %s", err)
		} else {
			if diff := cmp.Diff(api.Annotations, tc.expected); diff != "" {
				t.Errorf("Annotations were incorrectly set %+v", api.Annotations)
			}
		}
	}
	// test annotations for versions.
	for _, tc := range testCases {
		cmd := Command(ctx)
		cmd.SetArgs(append([]string{versionName}, tc.args...))
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", tc.args, err)
		}
		version, err := registryClient.GetApiVersion(ctx, &rpc.GetApiVersionRequest{
			Name: versionName,
		})
		if err != nil {
			t.Errorf("Error getting version %s", err)
		} else {
			if diff := cmp.Diff(version.Annotations, tc.expected); diff != "" {
				t.Errorf("Annotations were incorrectly set %+v", version.Annotations)
			}
		}
	}
	// test annotations for specs.
	for _, tc := range testCases {
		cmd := Command(ctx)
		cmd.SetArgs(append([]string{specName}, tc.args...))
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", tc.args, err)
		}
		spec, err := registryClient.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
			Name: specName,
		})
		if err != nil {
			t.Errorf("Error getting api %s", err)
		} else {
			if diff := cmp.Diff(spec.Annotations, tc.expected); diff != "" {
				t.Errorf("Annotations were incorrectly set %+v", spec.Annotations)
			}
		}
	}
	// test annotations for Deployments.
	for _, tc := range testCases {
		cmd := Command(ctx)
		cmd.SetArgs(append([]string{deploymentName}, tc.args...))
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", tc.args, err)
		}
		deployment, err := registryClient.GetApiDeployment(ctx, &rpc.GetApiDeploymentRequest{
			Name: deploymentName,
		})
		if err != nil {
			t.Errorf("Error getting deployment %s", err)
		} else {
			if diff := cmp.Diff(deployment.Annotations, tc.expected); diff != "" {
				t.Errorf("Annotations were incorrectly set %+v", deployment.Annotations)
			}
		}
	}
	// Delete the test project.
	{
		req := &rpc.DeleteProjectRequest{
			Name:  projectName,
			Force: true,
		}
		err = adminClient.DeleteProject(ctx, req)
		if err != nil {
			t.Fatalf("Failed to delete test project: %s", err)
		}
	}
}
