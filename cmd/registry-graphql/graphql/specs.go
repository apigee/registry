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

var specType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Spec",
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
			"mime_type": &graphql.Field{
				Type: graphql.String,
			},
			"size_bytes": &graphql.Field{
				Type: graphql.Int,
			},
			"hash": &graphql.Field{
				Type: graphql.String,
			},
			"source_uri": &graphql.Field{
				Type: graphql.String,
			},
			"revision_id": &graphql.Field{
				Type: graphql.String,
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

func representationForSpec(spec *rpc.ApiSpec) map[string]interface{} {
	result := map[string]interface{}{
		"id":          spec.Name,
		"filename":    spec.Filename,
		"description": spec.Description,
		"mime_type":   spec.MimeType,
		"size_bytes":  spec.SizeBytes,
		"hash":        spec.Hash,
		"source_uri":  spec.SourceUri,
		"revision_id": spec.RevisionId,
		"created":     representationForTimestamp(spec.CreateTime),
		"updated":     representationForTimestamp(spec.RevisionUpdateTime),
	}
	return result
}

func resolveSpecs(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	req := &rpc.ListApiSpecsRequest{
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
	var response *rpc.ListApiSpecsResponse
	edges := []map[string]interface{}{}
	for len(edges) < pageSize {
		response, _ = c.GrpcClient().ListApiSpecs(ctx, req)
		for _, spec := range response.GetApiSpecs() {
			edges = append(edges, representationForEdge(representationForSpec(spec)))
		}
		req.PageToken = response.GetNextPageToken()
		if req.PageToken == "" {
			break
		}
	}
	return connectionForEdgesAndEndCursor(edges, response.GetNextPageToken()), nil
}

func resolveSpec(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	name, isFound := p.Args["id"].(string)
	if !isFound {
		return nil, errors.New("missing id field")
	}
	req := &rpc.GetApiSpecRequest{
		Name: name,
	}
	spec, err := c.GetApiSpec(ctx, req)
	if err != nil {
		return nil, err
	}
	return representationForSpec(spec), err
}
