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
	"errors"
	"fmt"
	"net/http"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"google.golang.org/api/iterator"
)

// GraphQL types.

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: queryType,
	},
)

var argumentsForFilteredCollectionQuery = graphql.FieldConfigArgument{
	"filter": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
}

var argumentsForFilteredParentedCollectionQuery = graphql.FieldConfigArgument{
	"parent": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
	"filter": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
}

var argumentsForResourceQuery = graphql.FieldConfigArgument{
	"id": &graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	},
}

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"projects": &graphql.Field{
				Type:    graphql.NewList(projectType),
				Args:    argumentsForFilteredCollectionQuery,
				Resolve: resolveProjects,
			},
			"apis": &graphql.Field{
				Type:    graphql.NewList(apiType),
				Args:    argumentsForFilteredParentedCollectionQuery,
				Resolve: resolveAPIs,
			},
			"versions": &graphql.Field{
				Type:    graphql.NewList(versionType),
				Args:    argumentsForFilteredParentedCollectionQuery,
				Resolve: resolveVersions,
			},
			"specs": &graphql.Field{
				Type:    graphql.NewList(specType),
				Args:    argumentsForFilteredParentedCollectionQuery,
				Resolve: resolveSpecs,
			},
			"properties": &graphql.Field{
				Type:    graphql.NewList(propertyType),
				Args:    argumentsForFilteredParentedCollectionQuery,
				Resolve: resolveProperties,
			},
			"labels": &graphql.Field{
				Type:    graphql.NewList(apiType),
				Args:    argumentsForFilteredParentedCollectionQuery,
				Resolve: resolveAPIs,
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
			"property": &graphql.Field{
				Type:    propertyType,
				Args:    argumentsForResourceQuery,
				Resolve: resolveProperty,
			},
			"label": &graphql.Field{
				Type:    labelType,
				Args:    argumentsForResourceQuery,
				Resolve: resolveLabel,
			},
		},
	})

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
				Type:    graphql.NewList(apiType),
				Args:    argumentsForFilteredCollectionQuery,
				Resolve: resolveAPIs,
			},
			"labels": &graphql.Field{
				Type:    graphql.NewList(labelType),
				Args:    nil,
				Resolve: resolveLabels,
			},
			"properties": &graphql.Field{
				Type:    graphql.NewList(propertyType),
				Args:    nil,
				Resolve: resolveProperties,
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
				Type:    graphql.NewList(versionType),
				Args:    argumentsForFilteredCollectionQuery,
				Resolve: resolveVersions,
			},
			"labels": &graphql.Field{
				Type:    graphql.NewList(labelType),
				Args:    nil,
				Resolve: resolveLabels,
			},
			"properties": &graphql.Field{
				Type:    graphql.NewList(propertyType),
				Args:    nil,
				Resolve: resolveProperties,
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
				Type:    graphql.NewList(specType),
				Args:    argumentsForFilteredCollectionQuery,
				Resolve: resolveSpecs,
			},
			"labels": &graphql.Field{
				Type:    graphql.NewList(labelType),
				Args:    nil,
				Resolve: resolveLabels,
			},
			"properties": &graphql.Field{
				Type:    graphql.NewList(propertyType),
				Args:    nil,
				Resolve: resolveProperties,
			},
		},
	},
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
				Type:    graphql.NewList(labelType),
				Args:    nil,
				Resolve: resolveLabels,
			},
			"properties": &graphql.Field{
				Type:    graphql.NewList(propertyType),
				Args:    nil,
				Resolve: resolveProperties,
			},
		},
	},
)

var propertyType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Property",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var labelType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Label",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

// Convert proto objects to GraphQL representations.

func representationForProject(project *rpc.Project) map[string]interface{} {
	return map[string]interface{}{
		"id":           project.Name,
		"display_name": project.DisplayName,
		"description":  project.Description,
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

func representationForVersion(version *rpc.Version) map[string]interface{} {
	return map[string]interface{}{
		"id":           version.Name,
		"display_name": version.DisplayName,
		"description":  version.Description,
		"state":        version.State,
	}
}

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

func representationForProperty(property *rpc.Property) map[string]interface{} {
	return map[string]interface{}{
		"id": property.Name,
	}
}

func representationForLabel(label *rpc.Label) map[string]interface{} {
	return map[string]interface{}{
		"id": label.Name,
	}
}

// Resolvers for GraphQL fields.

func resolveProjects(p graphql.ResolveParams) (interface{}, error) {
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
		projects = append(projects, representationForProject(project))
	}
	return projects, nil
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

func resolveAPIs(p graphql.ResolveParams) (interface{}, error) {
	ctx := context.TODO()
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	req := &rpc.ListApisRequest{
		Parent: getParentFromParams(p),
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
		apis = append(apis, representationForAPI(api))
		count++
		if count > 10 {
			break
		}
	}
	return apis, nil
}

func resolveVersions(p graphql.ResolveParams) (interface{}, error) {
	ctx := context.TODO()
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	req := &rpc.ListVersionsRequest{
		Parent: getParentFromParams(p),
	}
	filter, isFound := p.Args["filter"].(string)
	if isFound {
		req.Filter = filter
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
		versions = append(versions, representationForVersion(version))
	}
	return versions, nil
}

func resolveSpecs(p graphql.ResolveParams) (interface{}, error) {
	ctx := context.TODO()
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	req := &rpc.ListSpecsRequest{
		Parent: getParentFromParams(p),
	}
	filter, isFound := p.Args["filter"].(string)
	if isFound {
		req.Filter = filter
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
		specs = append(specs, representationForSpec(spec))
	}
	return specs, nil
}

func resolveProperties(p graphql.ResolveParams) (interface{}, error) {
	ctx := context.TODO()
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	req := &rpc.ListPropertiesRequest{
		Parent: getParentFromParams(p),
	}
	it := c.ListProperties(ctx, req)
	properties := []map[string]interface{}{}
	count := 0
	for {
		property, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return nil, err
		}
		properties = append(properties, representationForProperty(property))
		count++
		if count > 10 {
			break
		}
	}
	return properties, nil
}

func resolveLabels(p graphql.ResolveParams) (interface{}, error) {
	ctx := context.TODO()
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	req := &rpc.ListLabelsRequest{
		Parent: getParentFromParams(p),
	}
	it := c.ListLabels(ctx, req)
	labels := []map[string]interface{}{}
	count := 0
	for {
		label, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return nil, err
		}
		labels = append(labels, representationForLabel(label))
		count++
		if count > 10 {
			break
		}
	}
	return labels, nil
}

func resolveProject(p graphql.ResolveParams) (interface{}, error) {
	ctx := context.TODO()
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	name, isFound := p.Args["id"].(string)
	if !isFound {
		return nil, errors.New("missing id field")
	}
	req := &rpc.GetProjectRequest{
		Name: name,
	}
	api, err := c.GetProject(ctx, req)
	return representationForProject(api), err
}

func resolveAPI(p graphql.ResolveParams) (interface{}, error) {
	ctx := context.TODO()
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	name, isFound := p.Args["id"].(string)
	if !isFound {
		return nil, errors.New("missing id field")
	}
	req := &rpc.GetApiRequest{
		Name: name,
	}
	api, err := c.GetApi(ctx, req)
	return representationForAPI(api), err
}

func resolveVersion(p graphql.ResolveParams) (interface{}, error) {
	ctx := context.TODO()
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	name, isFound := p.Args["id"].(string)
	if !isFound {
		return nil, errors.New("missing id field")
	}
	req := &rpc.GetVersionRequest{
		Name: name,
	}
	version, err := c.GetVersion(ctx, req)
	return representationForVersion(version), err
}

func resolveSpec(p graphql.ResolveParams) (interface{}, error) {
	ctx := context.TODO()
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
	}
	spec, err := c.GetSpec(ctx, req)
	return representationForSpec(spec), err
}

func resolveProperty(p graphql.ResolveParams) (interface{}, error) {
	ctx := context.TODO()
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

func resolveLabel(p graphql.ResolveParams) (interface{}, error) {
	ctx := context.TODO()
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	name, isFound := p.Args["id"].(string)
	if !isFound {
		return nil, errors.New("missing id field")
	}
	req := &rpc.GetLabelRequest{
		Name: name,
	}
	label, err := c.GetLabel(ctx, req)
	return representationForLabel(label), err
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

	// run the server
	port := "8088"
	fmt.Println("Running server on port " + port)
	http.ListenAndServe(":"+port, nil)
}
