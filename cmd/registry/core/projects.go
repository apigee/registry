// Copyright 2020 Google LLC. All Rights Reserved.
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

package core

import (
	"context"

	"github.com/apex/log"
	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
)

func EnsureProjectExists(ctx context.Context, client *gapic.RegistryClient, projectID string) {
	req := &rpc.GetProjectRequest{Name: "projects/" + projectID}
	if _, err := client.GetProject(ctx, req); NotFound(err) {
		req := &rpc.CreateProjectRequest{
			ProjectId: projectID,
			Project:   &rpc.Project{},
		}

		if _, err := client.CreateProject(ctx, req); err != nil {
			log.WithError(err).Fatal("Failed to create project")
		}
	} else if err != nil {
		log.WithError(err).Fatal("GetProject returned error during project existence check")
	}
}
