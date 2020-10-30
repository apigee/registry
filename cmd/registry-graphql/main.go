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
	"log"
	"net/http"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"google.golang.org/api/iterator"
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
				Args: nil,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					log.Printf("we need some apis")

					parent := p.Source.(map[string]interface{})["id"]
					log.Printf("%+v", parent)

					return nil, nil
				},
			},
		},
	},
)

func representationForProject(project *rpc.Project) map[string]interface{} {
	return map[string]interface{}{
		"id":           project.Name,
		"display_name": project.DisplayName,
		"description":  project.Description,
	}
}

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"projects": &graphql.Field{
				Type: graphql.NewList(projectType),
				Args: nil,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := context.TODO()
					c, err := connection.NewClient(ctx)
					if err != nil {
						return nil, err
					}
					req := &rpc.ListProjectsRequest{}
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
