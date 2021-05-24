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
	"log"

	"github.com/apigee/registry/gapic"
	rpcpb "github.com/apigee/registry/rpc"
)

func EnsureProjectExists(ctx context.Context, client *gapic.RegistryClient, projectID string) {
	// if the project doesn't exist, create it
	req := &rpcpb.GetProjectRequest{Name: "projects/" + projectID}
	_, err := client.GetProject(ctx, req)
	if NotFound(err) {
		req := &rpcpb.CreateProjectRequest{
			ProjectId: projectID,
			Project:   &rpcpb.Project{},
		}
		_, err := client.CreateProject(ctx, req)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
	} else if err != nil {
		log.Fatalf("GetProject returned error during project existence check: %s", err.Error())
	}
}
