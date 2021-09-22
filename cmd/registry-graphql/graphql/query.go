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

import "github.com/graphql-go/graphql"

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"projects": &graphql.Field{
				Type:    connectionType(projectType),
				Args:    argumentsForCollectionQuery,
				Resolve: resolveProjects,
			},
			"apis": &graphql.Field{
				Type:    connectionType(apiType),
				Args:    argumentsForParentedCollectionQuery,
				Resolve: resolveAPIs,
			},
			"versions": &graphql.Field{
				Type:    connectionType(versionType),
				Args:    argumentsForParentedCollectionQuery,
				Resolve: resolveVersions,
			},
			"specs": &graphql.Field{
				Type:    connectionType(specType),
				Args:    argumentsForParentedCollectionQuery,
				Resolve: resolveSpecs,
			},
			"artifacts": &graphql.Field{
				Type:    connectionType(artifactType),
				Args:    argumentsForParentedCollectionQuery,
				Resolve: resolveArtifacts,
			},
			"project": &graphql.Field{
				Type:    projectType,
				Args:    argumentsForResourceQuery,
				Resolve: resolveProject,
			},
			"api": &graphql.Field{
				Type:    apiType,
				Args:    argumentsForResourceQuery,
				Resolve: resolveAPI,
			},
			"version": &graphql.Field{
				Type:    versionType,
				Args:    argumentsForResourceQuery,
				Resolve: resolveVersion,
			},
			"spec": &graphql.Field{
				Type:    specType,
				Args:    argumentsForResourceQuery,
				Resolve: resolveSpec,
			},
			"artifact": &graphql.Field{
				Type:    artifactType,
				Args:    argumentsForResourceQuery,
				Resolve: resolveArtifact,
			},
		},
	})

var argumentsForCollectionQuery = graphql.FieldConfigArgument{
	"filter": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"after": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"first": &graphql.ArgumentConfig{
		Type: graphql.Int,
	},
}

var argumentsForParentedCollectionQuery = graphql.FieldConfigArgument{
	"parent": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
	"filter": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"after": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"first": &graphql.ArgumentConfig{
		Type: graphql.Int,
	},
}

var argumentsForResourceQuery = graphql.FieldConfigArgument{
	"id": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
}

func getParentFromParams(p graphql.ResolveParams) string {
	var parent string
	id := p.Source.(map[string]interface{})["id"]
	if id != nil {
		parent = id.(string)
	}
	name, isFound := p.Args["parent"]
	if isFound {
		parent = name.(string)
	}
	return parent
}
