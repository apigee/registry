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

package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/graphql-go/graphql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func TestGraphQL(t *testing.T) {
	// Create a registry client.
	ctx := context.Background()
	registryClient, err := connection.NewClient(ctx)
	if err != nil {
		t.Logf("Failed to create client: %+v", err)
		t.FailNow()
	}
	defer registryClient.Close()
	// Create sample registry.
	projectID := "test-graphql"
	buildTestProject(ctx, registryClient, t, projectID, 20)
	// Build query.
	query := `
 	  query ($cursor: String){
		project(id: "projects/test-graphql") {
		  id
		  display_name
		  apis(first: 5, filter: "api_id.matches('[02468]$')", after: $cursor) {
			edges {
			  node {
				id
			  }
			}
			pageInfo {
			  endCursor
			}
		  }
		}
	  }`
	params := &graphql.Params{
		Schema:        Schema,
		RequestString: query,
		VariableValues: map[string]interface{}{
			"cursor": "",
		},
		Context: ctx,
	}
	// Run the query.
	payload := evaluateQuery(params)
	if len(payload.Data.Project.APIs.Edges) != 5 {
		t.Errorf("Unexpected number of APIs from query 1")
	}
	// Update the cursor and repeat the query.
	params.VariableValues["cursor"] = payload.Data.Project.APIs.PageInfo.EndCursor
	payload = evaluateQuery(params)
	if len(payload.Data.Project.APIs.Edges) != 5 {
		t.Errorf("Unexpected number of APIs from query 2")
	}
	// Delete the test project
	deleteTestProject(ctx, registryClient, t, projectID)
}

func evaluateQuery(params *graphql.Params) *Payload {
	r := graphql.Do(*params)
	if len(r.Errors) > 0 {
		log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	rJSON, _ := json.Marshal(r)
	payload := &Payload{}
	json.Unmarshal(rJSON, payload)
	return payload
}

type Payload struct {
	Data Data `json:"data"`
}

type Data struct {
	Project Project `json:"project"`
}

type Project struct {
	APIs APIConnection `json:"apis"`
}

type APIConnection struct {
	Edges    []APIEdge `json:"edges"`
	PageInfo PageInfo  `json:"pageInfo"`
}

type APIEdge struct {
	Node API `json:"node"`
}

type PageInfo struct {
	EndCursor string `json:"endCursor"`
}

type API struct {
	ID string `json:"id"`
}

func buildTestProject(ctx context.Context, registryClient connection.Client, t *testing.T, name string, apiCount int) {
	deleteTestProject(ctx, registryClient, t, name)
	// Create the test project.
	req := &rpc.CreateProjectRequest{
		ProjectId: name,
		Project: &rpc.Project{
			DisplayName: "Test",
			Description: "A test catalog",
		},
	}
	project, err := registryClient.CreateProject(ctx, req)
	check(t, "error creating project %s", err)
	// Create test APIs.
	for i := 0; i < apiCount; i++ {
		req := &rpc.CreateApiRequest{
			Parent: project.GetName() + "/locations/global",
			ApiId:  fmt.Sprintf("api%03d", i),
			Api: &rpc.Api{
				DisplayName:  fmt.Sprintf("API-%03d", i),
				Description:  "A sample API",
				Availability: "GENERAL",
			},
		}
		_, err := registryClient.CreateApi(ctx, req)
		check(t, "error creating api %s", err)
	}
}

func deleteTestProject(ctx context.Context, registryClient connection.Client, t *testing.T, name string) {
	req := &rpc.DeleteProjectRequest{
		Name: "projects/" + name,
	}
	if err := registryClient.DeleteProject(ctx, req); status.Code(err) != codes.NotFound {
		check(t, "Failed to delete test project: %+v", err)
	}
}
