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

package registry

import (
	"errors"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/graphql-go/graphql"
)

var versionType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Version",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"display_name": &graphql.Field{
				Type: graphql.String,
			},
			"description": &graphql.Field{
				Type: graphql.String,
			},
			"specs": &graphql.Field{
				Type:    connectionType(specType),
				Args:    argumentsForCollectionQuery,
				Resolve: resolveSpecs,
			},
			"artifacts": &graphql.Field{
				Type:    connectionType(artifactType),
				Args:    argumentsForCollectionQuery,
				Resolve: resolveArtifacts,
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

func representationForVersion(version *rpc.ApiVersion) map[string]interface{} {
	return map[string]interface{}{
		"id":           version.Name,
		"display_name": version.DisplayName,
		"description":  version.Description,
		"state":        version.State,
		"created":      representationForTimestamp(version.CreateTime),
		"updated":      representationForTimestamp(version.UpdateTime),
	}
}
func resolveVersions(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	req := &rpc.ListApiVersionsRequest{
		Parent: getParentFromParams(p),
	}
	filter, isFound := p.Args["filter"].(string)
	if isFound {
		req.Filter = filter
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
	var response *rpc.ListApiVersionsResponse
	edges := []map[string]interface{}{}
	for len(edges) < pageSize {
		response, err = c.GrpcClient().ListApiVersions(ctx, req)
		for _, version := range response.GetApiVersions() {
			edges = append(edges, representationForEdge(representationForVersion(version)))
		}
		req.PageToken = response.GetNextPageToken()
		if req.PageToken == "" {
			break
		}
	}
	return connectionForEdgesAndEndCursor(edges, response.GetNextPageToken()), nil
}

func resolveVersion(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	name, isFound := p.Args["id"].(string)
	if !isFound {
		return nil, errors.New("missing id field")
	}
	req := &rpc.GetApiVersionRequest{
		Name: name,
	}
	version, err := c.GetApiVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	return representationForVersion(version), err
}
