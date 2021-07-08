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
		projectID   = "annotate-test"
		projectName = "projects/" + projectID
		apiID       = "sample"
		apiName     = projectName + "/apis/" + apiID
		versionID   = "1.0.0"
		versionName = apiName + "/versions/" + versionID
		specID      = "openapi.json"
		specName    = versionName + "/specs/" + specID
	)

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
		ProjectId: projectID,
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
		ApiId:  apiID,
		Api:    &rpc.Api{},
	})
	if err != nil {
		t.Fatalf("error creating api %s", err)
	}
	// Create a sample version.
	_, err = registryClient.CreateApiVersion(ctx, &rpc.CreateApiVersionRequest{
		Parent:       apiName,
		ApiVersionId: versionID,
		ApiVersion:   &rpc.ApiVersion{},
	})
	if err != nil {
		t.Fatalf("error creating version %s", err)
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
		t.Fatalf("error creating spec %s", err)
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
		cmd := Command()
		cmd.SetArgs(append([]string{apiName}, tc.args...))
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", tc.args, err)
		}
		api, err := registryClient.GetApi(ctx, &rpc.GetApiRequest{
			Name: apiName,
		})
		if err != nil {
			t.Errorf("error getting api %s", err)
		} else {
			if diff := cmp.Diff(api.Annotations, tc.expected); diff != "" {
				t.Errorf("annotations were incorrectly set %+v", api.Annotations)
			}
		}
	}
	// test annotations for versions.
	for _, tc := range testCases {
		cmd := Command()
		cmd.SetArgs(append([]string{versionName}, tc.args...))
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", tc.args, err)
		}
		version, err := registryClient.GetApiVersion(ctx, &rpc.GetApiVersionRequest{
			Name: versionName,
		})
		if err != nil {
			t.Errorf("error getting version %s", err)
		} else {
			if diff := cmp.Diff(version.Annotations, tc.expected); diff != "" {
				t.Errorf("annotations were incorrectly set %+v", version.Annotations)
			}
		}
	}
	// test annotations for specs.
	for _, tc := range testCases {
		cmd := Command()
		cmd.SetArgs(append([]string{specName}, tc.args...))
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() with args %+v returned error: %s", tc.args, err)
		}
		spec, err := registryClient.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
			Name: specName,
		})
		if err != nil {
			t.Errorf("error getting api %s", err)
		} else {
			if diff := cmp.Diff(spec.Annotations, tc.expected); diff != "" {
				t.Errorf("annotations were incorrectly set %+v", spec.Annotations)
			}
		}
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
