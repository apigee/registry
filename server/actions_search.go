// Copyright 2021 Google LLC. All Rights Reserved.
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

package server

import (
	"context"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/dao"
)

func (s *RegistryServer) SearchAll(ctx context.Context, in *rpc.SearchRequest) (*rpc.SearchResponse, error) {
	if !s.searchAvailable {
		return &rpc.SearchResponse{Message: "search is not available"}, nil
	}

	client, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, unavailableError(err)
	}
	defer s.releaseStorageClient(client)
	db := dao.NewDAO(client)

	rows, err := db.ListLexemes(ctx, in.GetQ())
	if err != nil {
		return nil, err
	}

	var results []*rpc.SearchResult
	for _, row := range rows.Rows {
		results = append(results, &rpc.SearchResult{
			Key:     row.Key,
			Excerpt: row.Raw,
		})
	}
	return &rpc.SearchResponse{
		Results:       results,
		NextPageToken: "",
	}, nil
}
