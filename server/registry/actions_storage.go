// Copyright 2022 Google LLC.
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

package registry

import (
	"context"

	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GetStorage handles the corresponding API request.
func (s *RegistryServer) GetStorage(ctx context.Context, req *emptypb.Empty) (*rpc.Storage, error) {
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	tableNames, err := db.TableNames(ctx)
	if err != nil {
		return nil, err
	}
	collections := make([]*rpc.Storage_Collection, 0)
	for _, tableName := range tableNames {
		count, err := db.RowCount(ctx, tableName)
		if err != nil {
			return nil, err
		}
		collections = append(collections,
			&rpc.Storage_Collection{
				Name:  tableName,
				Count: count,
			},
		)
	}
	return &rpc.Storage{
		Description: db.DatabaseName(ctx),
		Collections: collections,
	}, nil
}
