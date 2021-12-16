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

package benchmark

import (
	"context"
	"fmt"
	"testing"

	"github.com/apigee/registry/connection"
	"github.com/apigee/registry/rpc"
	"github.com/google/uuid"
)

const (
	rootResource = "projects/bench/locations/global"
)

func uniqueID() string {
	return fmt.Sprintf("%.8s", uuid.New())
}

func BenchmarkCreateApi(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}

	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &rpc.CreateApiRequest{
			Parent: rootResource,
			ApiId:  uniqueID(),
			Api:    &rpc.Api{},
		}

		b.StartTimer()
		api, err := client.CreateApi(ctx, req)
		b.StopTimer()

		if err != nil {
			b.Errorf("CreateApi(%+v) returned unexpected error: %s", req, err)
		}

		if err := client.DeleteApi(ctx, &rpc.DeleteApiRequest{Name: api.GetName()}); err != nil {
			b.Errorf("Cleanup: DeleteApi(%q) returned unexpected error: %s", api.GetName(), err)
		}
	}
}

func BenchmarkGetApi(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}

	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		api, err := client.CreateApi(ctx, &rpc.CreateApiRequest{
			Parent: rootResource,
			ApiId:  uniqueID(),
			Api:    &rpc.Api{},
		})
		if err != nil {
			b.Fatalf("Setup: Failed to create API: %s", err)
		}

		req := &rpc.GetApiRequest{
			Name: api.GetName(),
		}

		b.StartTimer()
		_, err = client.GetApi(ctx, req)
		b.StopTimer()

		if err != nil {
			b.Errorf("GetApi(%+v) returned unexpected error: %s", req, err)
		}

		if err := client.DeleteApi(ctx, &rpc.DeleteApiRequest{Name: api.GetName()}); err != nil {
			b.Errorf("Cleanup: DeleteApi(%q) returned unexpected error: %s", api.GetName(), err)
		}
	}
}

func BenchmarkListApis(b *testing.B) {
	b.Skip("TODO: Create ListApi benchmark test")
}

func BenchmarkUpdateApi(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}

	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		api, err := client.CreateApi(ctx, &rpc.CreateApiRequest{
			Parent: rootResource,
			ApiId:  uniqueID(),
			Api:    &rpc.Api{},
		})
		if err != nil {
			b.Fatalf("Setup: Failed to create API: %s", err)
		}

		req := &rpc.UpdateApiRequest{
			Api: &rpc.Api{
				Name:        api.GetName(),
				Description: uniqueID(),
			},
		}

		b.StartTimer()
		_, err = client.UpdateApi(ctx, req)
		b.StopTimer()

		if err != nil {
			b.Errorf("UpdateApi(%+v) returned unexpected error: %s", req, err)
		}

		if err := client.DeleteApi(ctx, &rpc.DeleteApiRequest{Name: api.GetName()}); err != nil {
			b.Errorf("Cleanup: DeleteApi(%q) returned unexpected error: %s", api.GetName(), err)
		}
	}
}

func BenchmarkDeleteApi(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}

	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		api, err := client.CreateApi(ctx, &rpc.CreateApiRequest{
			Parent: rootResource,
			ApiId:  uniqueID(),
			Api:    &rpc.Api{},
		})
		if err != nil {
			b.Fatalf("Setup: Failed to create API: %s", err)
		}

		req := &rpc.DeleteApiRequest{
			Name: api.GetName(),
		}

		b.StartTimer()
		err = client.DeleteApi(ctx, req)
		b.StopTimer()

		if err != nil {
			b.Errorf("DeleteApi(%+v) returned unexpected error: %s", req, err)
		}
	}
}
