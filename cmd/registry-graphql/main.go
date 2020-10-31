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
	"errors"
	"fmt"
	"net/http"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

// GraphQL types.

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: queryType,
	},
)

var argumentsForCollectionQuery = graphql.FieldConfigArgument{
	"filter": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"page_token": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"page_size": &graphql.ArgumentConfig{
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
	"page_token": &graphql.ArgumentConfig{
		Type: graphql.String,
	},
	"page_size": &graphql.ArgumentConfig{
		Type: graphql.Int,
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
				Type:    pageType(projectType),
				Args:    argumentsForCollectionQuery,
				Resolve: resolveProjects,
			},
			"apis": &graphql.Field{
				Type:    pageType(apiType),
				Args:    argumentsForParentedCollectionQuery,
				Resolve: resolveAPIs,
			},
			"versions": &graphql.Field{
				Type:    pageType(versionType),
				Args:    argumentsForParentedCollectionQuery,
				Resolve: resolveVersions,
			},
			"specs": &graphql.Field{
				Type:    pageType(specType),
				Args:    argumentsForParentedCollectionQuery,
				Resolve: resolveSpecs,
			},
			"properties": &graphql.Field{
				Type:    pageType(propertyType),
				Args:    argumentsForParentedCollectionQuery,
				Resolve: resolveProperties,
			},
			"labels": &graphql.Field{
				Type:    pageType(labelType),
				Args:    argumentsForParentedCollectionQuery,
				Resolve: resolveLabels,
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
				Type:    pageType(apiType),
				Args:    argumentsForCollectionQuery,
				Resolve: resolveAPIs,
			},
			"labels": &graphql.Field{
				Type:    pageType(labelType),
				Args:    argumentsForCollectionQuery,
				Resolve: resolveLabels,
			},
			"properties": &graphql.Field{
				Type:    pageType(propertyType),
				Args:    argumentsForCollectionQuery,
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
				Type:    pageType(versionType),
				Args:    argumentsForCollectionQuery,
				Resolve: resolveVersions,
			},
			"labels": &graphql.Field{
				Type:    pageType(labelType),
				Args:    argumentsForCollectionQuery,
				Resolve: resolveLabels,
			},
			"properties": &graphql.Field{
				Type:    pageType(propertyType),
				Args:    argumentsForCollectionQuery,
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
				Type:    pageType(specType),
				Args:    argumentsForCollectionQuery,
				Resolve: resolveSpecs,
			},
			"labels": &graphql.Field{
				Type:    pageType(labelType),
				Args:    argumentsForCollectionQuery,
				Resolve: resolveLabels,
			},
			"properties": &graphql.Field{
				Type:    pageType(propertyType),
				Args:    argumentsForCollectionQuery,
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
				Type:    pageType(labelType),
				Args:    argumentsForCollectionQuery,
				Resolve: resolveLabels,
			},
			"properties": &graphql.Field{
				Type:    pageType(propertyType),
				Args:    argumentsForCollectionQuery,
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

// Paging support

// generated page types should be built only once.
var pageTypeCache map[string]*graphql.Object

// pageType generates a wrapper type that represents a page in a list of objects.
func pageType(t graphql.Type) *graphql.Object {
	if pageTypeCache == nil {
		pageTypeCache = make(map[string]*graphql.Object)
	}
	name := t.Name() + "Page"
	p, isFound := pageTypeCache[name]
	if isFound {
		return p
	}
	p = graphql.NewObject(
		graphql.ObjectConfig{
			Name: name,
			Fields: graphql.Fields{
				"values": &graphql.Field{
					Type: graphql.NewList(t),
				},
				"next_page_token": &graphql.Field{
					Type: graphql.String,
				},
			},
		},
	)
	pageTypeCache[name] = p
	return p
}

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

// Helper

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

// Resolvers for GraphQL fields.

func resolveProjects(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	req := &rpc.ListProjectsRequest{}
	filter, isFound := p.Args["filter"].(string)
	if isFound {
		req.Filter = filter
	}
	pageToken, isFound := p.Args["page_token"].(string)
	if isFound {
		req.PageToken = pageToken
	}
	pageSize, isFound := p.Args["page_size"].(int)
	if isFound {
		req.PageSize = int32(pageSize)
	}
	response, err := c.GrpcClient().ListProjects(ctx, req)
	projects := []map[string]interface{}{}
	for _, project := range response.GetProjects() {
		projects = append(projects, representationForProject(project))
	}
	page := make(map[string]interface{}, 0)
	page["values"] = projects
	page["next_page_token"] = response.GetNextPageToken()
	return page, nil
}

func resolveAPIs(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
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
	pageToken, isFound := p.Args["page_token"].(string)
	if isFound {
		req.PageToken = pageToken
	}
	pageSize, isFound := p.Args["page_size"].(int)
	if isFound {
		req.PageSize = int32(pageSize)
	}
	response, err := c.GrpcClient().ListApis(ctx, req)
	apis := []map[string]interface{}{}
	for _, api := range response.GetApis() {
		apis = append(apis, representationForAPI(api))
	}
	page := make(map[string]interface{}, 0)
	page["values"] = apis
	page["next_page_token"] = response.GetNextPageToken()
	return page, nil
}

func resolveVersions(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
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
	pageToken, isFound := p.Args["page_token"].(string)
	if isFound {
		req.PageToken = pageToken
	}
	pageSize, isFound := p.Args["page_size"].(int)
	if isFound {
		req.PageSize = int32(pageSize)
	}
	response, err := c.GrpcClient().ListVersions(ctx, req)
	versions := []map[string]interface{}{}
	for _, version := range response.GetVersions() {
		versions = append(versions, representationForVersion(version))
	}
	page := make(map[string]interface{}, 0)
	page["values"] = versions
	page["next_page_token"] = response.GetNextPageToken()
	return page, nil
}

func resolveSpecs(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
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
	pageToken, isFound := p.Args["page_token"].(string)
	if isFound {
		req.PageToken = pageToken
	}
	pageSize, isFound := p.Args["page_size"].(int)
	if isFound {
		req.PageSize = int32(pageSize)
	}
	response, err := c.GrpcClient().ListSpecs(ctx, req)
	specs := []map[string]interface{}{}
	for _, spec := range response.GetSpecs() {
		specs = append(specs, representationForSpec(spec))
	}
	page := make(map[string]interface{}, 0)
	page["values"] = specs
	page["next_page_token"] = response.GetNextPageToken()
	return page, nil
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
	pageToken, isFound := p.Args["page_token"].(string)
	if isFound {
		req.PageToken = pageToken
	}
	pageSize, isFound := p.Args["page_size"].(int)
	if isFound {
		req.PageSize = int32(pageSize)
	}
	response, err := c.GrpcClient().ListProperties(ctx, req)
	properties := []map[string]interface{}{}
	for _, property := range response.GetProperties() {
		properties = append(properties, representationForProperty(property))
	}
	page := make(map[string]interface{}, 0)
	page["values"] = properties
	page["next_page_token"] = response.GetNextPageToken()
	return page, nil
}

func resolveLabels(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	c, err := connection.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	req := &rpc.ListLabelsRequest{
		Parent: getParentFromParams(p),
	}
	pageToken, isFound := p.Args["page_token"].(string)
	if isFound {
		req.PageToken = pageToken
	}
	pageSize, isFound := p.Args["page_size"].(int)
	if isFound {
		req.PageSize = int32(pageSize)
	}
	response, err := c.GrpcClient().ListLabels(ctx, req)
	labels := []map[string]interface{}{}
	for _, label := range response.GetLabels() {
		labels = append(labels, representationForLabel(label))
	}
	page := make(map[string]interface{}, 0)
	page["values"] = labels
	page["next_page_token"] = response.GetNextPageToken()
	return page, nil
}

func resolveProject(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
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
	ctx := p.Context
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
	ctx := p.Context
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
	}
	spec, err := c.GetSpec(ctx, req)
	return representationForSpec(spec), err
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

func resolveLabel(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
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
