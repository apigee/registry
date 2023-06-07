// Copyright 2021 Google LLC.
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

	longrunning "cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

// MigrateDatabase handles the corresponding API request.
func (s *RegistryServer) MigrateDatabase(ctx context.Context, req *rpc.MigrateDatabaseRequest) (*longrunning.Operation, error) {
	if req.Kind != "" && req.Kind != "auto" {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported migration kind %q", req.Kind)
	}
	db, err := s.getStorageClient(ctx)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	err = db.Migrate(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	metadata, err := anypb.New(&rpc.MigrateDatabaseMetadata{})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response, err := anypb.New(&rpc.MigrateDatabaseResponse{
		Message: "OK",
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &longrunning.Operation{
		Name:     "migrate",
		Metadata: metadata,
		Done:     true,
		Result:   &longrunning.Operation_Response{Response: response},
	}, nil
}
