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
	"encoding/base64"
	"errors"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
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
			"style": &graphql.Field{
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
			"labels": &graphql.Field{
				Type:    connectionType(labelType),
				Args:    argumentsForCollectionQuery,
				Resolve: resolveLabels,
			},
			"properties": &graphql.Field{
				Type:    connectionType(propertyType),
				Args:    argumentsForCollectionQuery,
				Resolve: resolveProperties,
			},
			"created": &graphql.Field{
				Type: timestampType,
			},
			"updated": &graphql.Field{
				Type: timestampType,
			},
			"contents": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

func representationForSpec(spec *rpc.Spec, view rpc.View) map[string]interface{} {
	result := map[string]interface{}{
		"id":          spec.Name,
		"filename":    spec.Filename,
		"description": spec.Description,
		"style":       spec.Style,
		"size_bytes":  spec.SizeBytes,
		"hash":        spec.Hash,
		"source_uri":  spec.SourceUri,
		"revision_id": spec.RevisionId,
		"created":     representationForTimestamp(spec.CreateTime),
		"updated":     representationForTimestamp(spec.UpdateTime),
	}
	if view == rpc.View_FULL {
		result["contents"] = base64.StdEncoding.EncodeToString([]byte(spec.Contents))
	}
	return result
}

func resolveSpecs(p graphql.ResolveParams) (interface{}, error) {
	// use the presence of "specs/edges/node/contents" in the request to determine the view
	view := rpc.View_BASIC
	if selectionSetContainsPath(p.Info.Operation.GetSelectionSet(), []string{"specs", "edges", "node", "contents"}) {
		view = rpc.View_FULL
	}
	ctx := p.Context
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	req := &rpc.ListSpecsRequest{
		Parent: getParentFromParams(p),
		View:   view,
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
	var response *rpc.ListSpecsResponse
	edges := []map[string]interface{}{}
	for len(edges) < pageSize {
		response, err = c.GrpcClient().ListSpecs(ctx, req)
		for _, spec := range response.GetSpecs() {
			edges = append(edges, representationForEdge(representationForSpec(spec, view)))
		}
		req.PageToken = response.GetNextPageToken()
		if req.PageToken == "" {
			break
		}
	}
	return connectionForEdgesAndEndCursor(edges, response.GetNextPageToken()), nil
}

// returns true if a selection set contains a path that ends with a specified path
func selectionSetContainsPath(selectionSet *ast.SelectionSet, path []string) bool {
	if len(path) == 0 {
		return true
	}
	if selectionSet == nil {
		return false
	}
	for _, s := range selectionSet.Selections {
		fieldName := s.(*ast.Field).Name.Value
		fieldSelection := s.(*ast.Field).GetSelectionSet()
		if fieldName == path[0] {
			return selectionSetContainsPath(fieldSelection, path[1:])
		} else {
			if selectionSetContainsPath(fieldSelection, path) {
				return true
			}
		}
	}
	return false
}

func resolveSpec(p graphql.ResolveParams) (interface{}, error) {
	// use the presence of "specs/contents" in the request to determine the view
	view := rpc.View_BASIC
	if selectionSetContainsPath(p.Info.Operation.GetSelectionSet(), []string{"spec", "contents"}) {
		view = rpc.View_FULL
	}
	ctx := p.Context
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	name, isFound := p.Args["id"].(string)
	if !isFound {
		return nil, errors.New("missing id field")
	}
	req := &rpc.GetSpecRequest{
		Name: name,
		View: view,
	}
	spec, err := c.GetSpec(ctx, req)
	if err != nil {
		return nil, err
	}
	return representationForSpec(spec, view), err
}
