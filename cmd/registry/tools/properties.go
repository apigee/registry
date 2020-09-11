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

package tools

import (
	"context"

	"github.com/apigee/registry/gapic"
	"github.com/apigee/registry/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func SetProperty(ctx context.Context,
	client *gapic.RegistryClient,
	property *rpc.Property) error {
	request := &rpc.CreatePropertyRequest{}
	request.Property = property
	request.PropertyId = property.GetRelation()
	request.Parent = property.GetSubject()
	// First try setting a new property value.
	_, err := client.CreateProperty(ctx, request)
	if err == nil {
		return nil
	}
	// If that failed because the property already exists, update it.
	code := status.Code(err)
	if code == codes.AlreadyExists {
		request := &rpc.UpdatePropertyRequest{}
		request.Property = property
		_, err := client.UpdateProperty(ctx, request)
		return err
	}
	return err
}
