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

var propertyType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Property",
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

func representationForProperty(property *rpc.Property) map[string]interface{} {
	return map[string]interface{}{
		"id":      property.Name,
		"created": representationForTimestamp(property.CreateTime),
		"updated": representationForTimestamp(property.UpdateTime),
	}
}

func resolveProperties(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	req := &rpc.ListPropertiesRequest{
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
	var response *rpc.ListPropertiesResponse
	edges := []map[string]interface{}{}
	for len(edges) < pageSize {
		response, err = c.GrpcClient().ListProperties(ctx, req)
		for _, property := range response.GetProperties() {
			edges = append(edges, representationForEdge(representationForProperty(property)))
		}
		req.PageToken = response.GetNextPageToken()
		if req.PageToken == "" {
			break
		}
	}
	return connectionForEdgesAndEndCursor(edges, response.GetNextPageToken()), nil
}

func resolveProperty(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	name, isFound := p.Args["id"].(string)
	if !isFound {
		return nil, errors.New("missing id field")
	}
	req := &rpc.GetPropertyRequest{
		Name: name,
	}
	property, err := c.GetProperty(ctx, req)
	return representationForProperty(property), err
}
