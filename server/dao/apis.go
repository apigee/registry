// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

import (
	"context"

	"github.com/apigee/registry/server/models"
	"github.com/apigee/registry/server/names"
	"github.com/apigee/registry/server/storage"
	"github.com/apigee/registry/server/storage/filtering"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ApiList contains a page of api resources.
type ApiList struct {
	Apis  []models.Api
	Token string
}

var apiFields = []filtering.Field{
	{Name: "name", Type: filtering.String},
	{Name: "project_id", Type: filtering.String},
	{Name: "api_id", Type: filtering.String},
	{Name: "display_name", Type: filtering.String},
	{Name: "description", Type: filtering.String},
	{Name: "create_time", Type: filtering.Timestamp},
	{Name: "update_time", Type: filtering.Timestamp},
	{Name: "availability", Type: filtering.String},
	{Name: "recommended_version", Type: filtering.String},
	{Name: "labels", Type: filtering.StringMap},
}

func (d *DAO) ListApis(ctx context.Context, parent names.Project, opts PageOptions) (ApiList, error) {
	q := d.NewQuery(storage.ApiEntityName)

	token, err := decodeToken(opts.Token)
	if err != nil {
		return ApiList{}, status.Errorf(codes.InvalidArgument, "invalid page token %q: %s", opts.Token, err.Error())
	}

	if err := token.ValidateFilter(opts.Filter); err != nil {
		return ApiList{}, status.Errorf(codes.InvalidArgument, "invalid filter %q: %s", opts.Filter, err)
	} else {
		token.Filter = opts.Filter
	}

	q = q.ApplyOffset(token.Offset)

	if parent.ProjectID != "-" {
		q = q.Require("ProjectID", parent.ProjectID)
		if _, err := d.GetProject(ctx, parent); err != nil {
			return ApiList{}, err
		}
	}

	filter, err := filtering.NewFilter(opts.Filter, apiFields)
	if err != nil {
		return ApiList{}, err
	}

	it := d.Run(ctx, q)
	response := ApiList{
		Apis: make([]models.Api, 0, opts.Size),
	}

	api := new(models.Api)
	for _, err = it.Next(api); err == nil; _, err = it.Next(api) {
		apiMap, err := apiMap(*api)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}

		match, err := filter.Matches(apiMap)
		if err != nil {
			return response, err
		} else if !match {
			token.Offset++
			continue
		} else if len(response.Apis) == int(opts.Size) {
			break
		}

		response.Apis = append(response.Apis, *api)
		token.Offset++
	}
	if err != nil && err != iterator.Done {
		return response, status.Error(codes.Internal, err.Error())
	}

	if err == nil {
		response.Token, err = encodeToken(token)
		if err != nil {
			return response, status.Error(codes.Internal, err.Error())
		}
	}

	return response, nil
}

func apiMap(api models.Api) (map[string]interface{}, error) {
	labels, err := api.LabelsMap()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":                api.Name(),
		"project_id":          api.ProjectID,
		"api_id":              api.ApiID,
		"display_name":        api.DisplayName,
		"description":         api.Description,
		"create_time":         api.CreateTime,
		"update_time":         api.UpdateTime,
		"availability":        api.Availability,
		"recommended_version": api.RecommendedVersion,
		"labels":              labels,
	}, nil
}

func (d *DAO) GetApi(ctx context.Context, name names.Api) (*models.Api, error) {
	api := new(models.Api)
	k := d.NewKey(storage.ApiEntityName, name.String())
	if err := d.Get(ctx, k, api); d.IsNotFound(err) {
		return nil, status.Errorf(codes.NotFound, "api %q not found in database", name)
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return api, nil
}

func (d *DAO) SaveApi(ctx context.Context, api *models.Api) error {
	k := d.NewKey(storage.ApiEntityName, api.Name())
	if _, err := d.Put(ctx, k, api); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (d *DAO) DeleteApi(ctx context.Context, name names.Api) error {
	if err := d.DeleteChildrenOfApi(ctx, name); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	k := d.NewKey(storage.ApiEntityName, name.String())
	if err := d.Delete(ctx, k); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}
