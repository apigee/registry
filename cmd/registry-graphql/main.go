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

package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"google.golang.org/api/iterator"
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
		},
	},
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
				Type: graphql.NewList(specType),
				Args: graphql.FieldConfigArgument{
					"filter": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					parent := p.Source.(map[string]interface{})["id"].(string)
					ctx := context.TODO()
					c, err := connection.NewClient(ctx)
					if err != nil {
						return nil, err
					}
					req := &rpc.ListSpecsRequest{
						Parent: parent,
					}
					it := c.ListSpecs(ctx, req)
					specs := []map[string]interface{}{}
					for {
						spec, err := it.Next()
						if err == iterator.Done {
							break
						} else if err != nil {
							return nil, err
						}
						specs = append(specs,
							representationForSpec(spec))
					}
					return specs, nil
				},
			},
		},
	},
)

var apiType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "API",
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
			"versions": &graphql.Field{
				Type: graphql.NewList(versionType),
				Args: graphql.FieldConfigArgument{
					"filter": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					parent := p.Source.(map[string]interface{})["id"].(string)
					ctx := context.TODO()
					c, err := connection.NewClient(ctx)
					if err != nil {
						return nil, err
					}
					req := &rpc.ListVersionsRequest{
						Parent: parent,
					}
					it := c.ListVersions(ctx, req)
					versions := []map[string]interface{}{}
					for {
						version, err := it.Next()
						if err == iterator.Done {
							break
						} else if err != nil {
							return nil, err
						}
						versions = append(versions,
							representationForVersion(version))
					}
					return versions, nil
				},
			},
		},
	},
)

var projectType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Project",
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
			"apis": &graphql.Field{
				Type: graphql.NewList(apiType),
				Args: graphql.FieldConfigArgument{
					"filter": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					parent := p.Source.(map[string]interface{})["id"].(string)
					ctx := context.TODO()
					c, err := connection.NewClient(ctx)
					if err != nil {
						return nil, err
					}
					req := &rpc.ListApisRequest{
						Parent: parent,
					}
					filter, isFound := p.Args["filter"].(string)
					if isFound {
						req.Filter = filter
					}
					it := c.ListApis(ctx, req)
					apis := []map[string]interface{}{}
					count := 0
					for {
						api, err := it.Next()
						if err == iterator.Done {
							break
						} else if err != nil {
							return nil, err
						}
						apis = append(apis,
							representationForAPI(api))
						count++
						if count > 20 {
							break
						}
					}
					return apis, nil
				},
			},
		},
	},
)

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"projects": &graphql.Field{
				Type: graphql.NewList(projectType),
				Args: graphql.FieldConfigArgument{
					"filter": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := context.TODO()
					c, err := connection.NewClient(ctx)
					if err != nil {
						return nil, err
					}
					req := &rpc.ListProjectsRequest{}
					filter, isFound := p.Args["filter"].(string)
					if isFound {
						req.Filter = filter
					}
					it := c.ListProjects(ctx, req)
					projects := []map[string]interface{}{}
					for {
						project, err := it.Next()
						if err == iterator.Done {
							break
						} else if err != nil {
							return nil, err
						}
						projects = append(projects,
							representationForProject(project))
					}
					return projects, nil
				},
			},
		},
	})

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: queryType,
	},
)

func representationForSpec(spec *rpc.Spec) map[string]interface{} {
	return map[string]interface{}{
		"id":          spec.Name,
		"filename":    spec.Filename,
		"description": spec.Description,
		"style":       spec.Style,
		"size_bytes":  spec.SizeBytes,
		"hash":        spec.Hash,
		"source_uri":  spec.SourceUri,
		"revision_id": spec.RevisionId,
	}
}

func representationForVersion(version *rpc.Version) map[string]interface{} {
	return map[string]interface{}{
		"id":           version.Name,
		"display_name": version.DisplayName,
		"description":  version.Description,
		"state":        version.State,
	}
}

func representationForAPI(api *rpc.Api) map[string]interface{} {
	return map[string]interface{}{
		"id":                  api.Name,
		"display_name":        api.DisplayName,
		"description":         api.Description,
		"availability":        api.Availability,
		"recommended_version": api.RecommendedVersion,
		"owner":               api.Owner,
	}
}

func representationForProject(project *rpc.Project) map[string]interface{} {
	return map[string]interface{}{
		"id":           project.Name,
		"display_name": project.DisplayName,
		"description":  project.Description,
	}
}

func main() {
	// graphql handler
	h := handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	})
	http.Handle("/graphql", h)

	// static file server for Graphiql in-browser editor
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	fmt.Println("Running server on port 8088")
	http.ListenAndServe(":8088", nil)
}
