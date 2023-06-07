// Copyright 2023 Google LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package patch

import (
	"context"
	"testing"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/pkg/names"
	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestArtifactLists(t *testing.T) {
	tests := []struct {
		desc string
		root string
	}{
		{
			desc: "artifacts",
			root: "testdata/lists",
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			project := names.Project{ProjectID: "patch-lists-test"}
			registryClient, adminClient := grpctest.SetupRegistry(ctx, t, project.ProjectID, nil)

			// set the configured registry.project to the test project
			config, err := connection.ActiveConfig()
			if err != nil {
				t.Fatalf("Setup: Failed to get registry configuration: %s", err)
			}
			config.Project = project.ProjectID
			connection.SetConfig(config)

			if err := Apply(ctx, registryClient, adminClient, nil, project.String()+"/locations/global", true, 10, test.root); err != nil {
				t.Fatalf("Apply() returned error: %s", err)
			}

			artifacts := []string{
				"apihub-display-settings",
				"apihub-lifecycle",
				"apihub-lint-errors",
				"apihub-lint-summary",
				"apihub-lint-warnings",
				"apihub-styleguide",
				"apihub-taxonomies",
			}
			for _, a := range artifacts {
				_, err := registryClient.GetArtifact(ctx, &rpc.GetArtifactRequest{
					Name: project.Artifact(a).String(),
				})
				if status.Code(err) == codes.NotFound {
					t.Fatalf("Expected artifact doesn't exist: %s", err)
				} else if err != nil {
					t.Fatalf("Failed to verify artifact existence: %s", err)
				}
			}
		})
	}
}
