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
	"github.com/graphql-go/graphql"
)

// Paging support

var pageInfoType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "PageInfo",
		Fields: graphql.Fields{
			"endCursor": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

// generated page types should be built only once.
var connectionTypeCache map[string]*graphql.Object

// connectionType generates a wrapper type that represents a page in a list of objects.
func connectionType(t graphql.Type) *graphql.Object {
	if connectionTypeCache == nil {
		connectionTypeCache = make(map[string]*graphql.Object)
	}
	name := t.Name() + "Connection"
	p, isFound := connectionTypeCache[name]
	if isFound {
		return p
	}
	p = graphql.NewObject(
		graphql.ObjectConfig{
			Name: name,
			Fields: graphql.Fields{
				"edges": &graphql.Field{
					Type: graphql.NewList(
						graphql.NewObject(
							graphql.ObjectConfig{
								Name: t.Name() + "Edges",
								Fields: graphql.Fields{
									"node": &graphql.Field{
										Type: t,
									},
								},
							},
						),
					),
				},
				"pageInfo": &graphql.Field{
					Type: pageInfoType,
				},
			},
		},
	)
	connectionTypeCache[name] = p
	return p
}

func representationForPageInfo(endCursor string) map[string]interface{} {
	return map[string]interface{}{
		"endCursor": endCursor,
	}
}

func representationForEdge(node interface{}) map[string]interface{} {
	return map[string]interface{}{
		"node": node,
	}
}

func connectionForEdgesAndEndCursor(edges []map[string]interface{}, cursor string) map[string]interface{} {
	return map[string]interface{}{
		"edges":    edges,
		"pageInfo": representationForPageInfo(cursor),
	}
}
