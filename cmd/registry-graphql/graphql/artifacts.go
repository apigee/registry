// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package graphql

import (
	"errors"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/graphql-go/graphql"
)

var artifactType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Artifact",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"created": &graphql.Field{
				Type: timestampType,
			},
			"updated": &graphql.Field{
				Type: timestampType,
			},
		},
	},
)

func representationForArtifact(artifact *rpc.Artifact) map[string]interface{} {
	return map[string]interface{}{
		"id":      artifact.Name,
		"created": representationForTimestamp(artifact.CreateTime),
		"updated": representationForTimestamp(artifact.UpdateTime),
	}
}

func resolveArtifacts(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	req := &rpc.ListArtifactsRequest{
		Parent: getParentFromParams(p),
	}
	pageToken, isFound := p.Args["after"].(string)
	if isFound {
		req.PageToken = pageToken
	}
	pageSize, isFound := p.Args["first"].(int)
	if isFound {
		req.PageSize = int32(pageSize)
	} else {
		pageSize = 50
	}
	var response *rpc.ListArtifactsResponse
	edges := []map[string]interface{}{}
	for len(edges) < pageSize {
		response, _ = c.GrpcClient().ListArtifacts(ctx, req)
		for _, artifact := range response.GetArtifacts() {
			edges = append(edges, representationForEdge(representationForArtifact(artifact)))
		}
		req.PageToken = response.GetNextPageToken()
		if req.PageToken == "" {
			break
		}
	}
	return connectionForEdgesAndEndCursor(edges, response.GetNextPageToken()), nil
}

func resolveArtifact(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	name, isFound := p.Args["id"].(string)
	if !isFound {
		return nil, errors.New("missing id field")
	}
	req := &rpc.GetArtifactRequest{
		Name: name,
	}
	artifact, err := c.GetArtifact(ctx, req)
	if err != nil {
		return nil, err
	}
	return representationForArtifact(artifact), err
}
