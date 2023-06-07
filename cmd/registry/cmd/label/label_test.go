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

package label

import (
	"context"
	"testing"

	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/google/go-cmp/cmp"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestLabel(t *testing.T) {
	const (
		projectID      = "label-test"
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
	registryClient, adminClient := grpctest.SetupRegistry(ctx, t, projectID, nil)

	// Create a sample api.
	_, err := registryClient.CreateApi(ctx, &rpc.CreateApiRequest{
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
		{comment: "add some labels",
			args:     []string{"a=1", "b=2"},
			expected: map[string]string{"a": "1", "b": "2"}},
		{comment: "remove one label and overwrite the other",
			args:     []string{"a=3", "b-", "--overwrite"},
			expected: map[string]string{"a": "3"}},
		{comment: "changing a label without --overwrite should be ignored",
			args:     []string{"a=4"},
			expected: map[string]string{"a": "3"}},
	}
	// test labels for APIs.
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
			t.Errorf("Error getting api %s", err)
		} else {
			if diff := cmp.Diff(api.Labels, tc.expected); diff != "" {
				t.Errorf("labels were incorrectly set %+v", api.Labels)
			}
		}
	}
	// test labels for versions.
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
			t.Errorf("Error getting version %s", err)
		} else {
			if diff := cmp.Diff(version.Labels, tc.expected); diff != "" {
				t.Errorf("labels were incorrectly set %+v", version.Labels)
			}
		}
	}
	// test labels for specs.
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
			t.Errorf("Error getting api %s", err)
		} else {
			if diff := cmp.Diff(spec.Labels, tc.expected); diff != "" {
				t.Errorf("labels were incorrectly set %+v", spec.Labels)
			}
		}
	}
	// test labels for Deployments.
	for _, tc := range testCases {
		cmd := Command()
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
			if diff := cmp.Diff(deployment.Labels, tc.expected); diff != "" {
				t.Errorf("labels were incorrectly set %+v", deployment.Labels)
			}
		}
	}

	// Delete the test project.
	if false {
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
