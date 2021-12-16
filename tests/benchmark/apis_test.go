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

	var (
		creates = make([]*rpc.CreateApiRequest, b.N)
		deletes = make([]*rpc.DeleteApiRequest, b.N)
	)

	for i := 0; i < b.N; i++ {
		creates[i] = &rpc.CreateApiRequest{
			Parent: rootResource,
			ApiId:  uniqueID(),
			Api:    &rpc.Api{},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.CreateApi(ctx, creates[i])
		if err != nil {
			b.Errorf("CreateApi(%+v) returned unexpected error: %s", creates[i], err)
		}

		deletes[i] = &rpc.DeleteApiRequest{
			Name: resp.GetName(),
		}
	}
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		if err := client.DeleteApi(ctx, deletes[i]); err != nil {
			b.Errorf("Cleanup: DeleteApi(%q) returned unexpected error: %s", deletes[i].GetName(), err)
		}
	}
}

func BenchmarkGetApi(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}

	var (
		gets    = make([]*rpc.GetApiRequest, b.N)
		deletes = make([]*rpc.DeleteApiRequest, b.N)
	)

	for i := 0; i < b.N; i++ {
		api, err := client.CreateApi(ctx, &rpc.CreateApiRequest{
			Parent: rootResource,
			ApiId:  uniqueID(),
			Api:    &rpc.Api{},
		})
		if err != nil {
			b.Fatalf("Setup: Failed to create API: %s", err)
		}

		gets[i] = &rpc.GetApiRequest{
			Name: api.GetName(),
		}

		deletes[i] = &rpc.DeleteApiRequest{
			Name: api.GetName(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.GetApi(ctx, gets[i]); err != nil {
			b.Errorf("GetApi(%q) returned unexpected error: %s", gets[i].GetName(), err)
		}
	}
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		if err := client.DeleteApi(ctx, deletes[i]); err != nil {
			b.Errorf("Cleanup: DeleteApi(%q) returned unexpected error: %s", deletes[i].GetName(), err)
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

	var (
		updates = make([]*rpc.UpdateApiRequest, b.N)
		deletes = make([]*rpc.DeleteApiRequest, b.N)
	)

	for i := 0; i < b.N; i++ {
		api, err := client.CreateApi(ctx, &rpc.CreateApiRequest{
			Parent: rootResource,
			ApiId:  uniqueID(),
			Api:    &rpc.Api{},
		})
		if err != nil {
			b.Fatalf("Setup: Failed to create API: %s", err)
		}

		updates[i] = &rpc.UpdateApiRequest{
			Api: &rpc.Api{
				Name:        api.GetName(),
				Description: uniqueID(),
			},
		}

		deletes[i] = &rpc.DeleteApiRequest{
			Name: api.GetName(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.UpdateApi(ctx, updates[i]); err != nil {
			b.Errorf("UpdateApi(%+v) returned unexpected error: %s", updates[i], err)
		}
	}
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		if err := client.DeleteApi(ctx, deletes[i]); err != nil {
			b.Errorf("Cleanup: DeleteApi(%q) returned unexpected error: %s", deletes[i].GetName(), err)
		}
	}
}

func BenchmarkDeleteApi(b *testing.B) {
	ctx := context.Background()
	client, err := connection.NewClient(ctx)
	if err != nil {
		b.Fatalf("Setup: Failed to create client: %s", err)
	}

	deletes := make([]*rpc.DeleteApiRequest, b.N)
	for i := 0; i < b.N; i++ {
		api, err := client.CreateApi(ctx, &rpc.CreateApiRequest{
			Parent: rootResource,
			ApiId:  uniqueID(),
			Api:    &rpc.Api{},
		})
		if err != nil {
			b.Fatalf("Setup: Failed to create API: %s", err)
		}

		deletes[i] = &rpc.DeleteApiRequest{
			Name: api.GetName(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := client.DeleteApi(ctx, deletes[i]); err != nil {
			b.Errorf("DeleteApi(%q) returned unexpected error: %s", deletes[i].GetName(), err)
		}
	}
}
