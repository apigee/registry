// Copyright 2022 Google LLC. All Rights Reserved.
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
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/protobuf/field_mask"
)

func apiId(i int) string {
	return fmt.Sprintf("test-%d", i)
}

func apiName(apiId string) string {
	return fmt.Sprintf("%s/apis/%s", root(), apiId)
}

func getApi(b *testing.B, ctx context.Context, client connection.Client, apiId string) {
	b.Helper()
	if _, err := client.GetApi(ctx, &rpc.GetApiRequest{Name: apiName(apiId)}); err != nil {
		b.Errorf("GetApi(%s) returned unexpected error: %s", apiId, err)
	}
}

func listApis(b *testing.B, ctx context.Context, client connection.Client) {
	b.Helper()
	it := client.ListApis(ctx, &rpc.ListApisRequest{Parent: root()})
	for _, err := it.Next(); err != iterator.Done; _, err = it.Next() {
		if err != nil {
			b.Errorf("ListApis(%s) returned unexpected error: %s", root(), err)
		}
	}
}

func createApi(b *testing.B, ctx context.Context, client connection.Client, apiId string) *rpc.Api {
	b.Helper()
	api, err := client.CreateApi(ctx, &rpc.CreateApiRequest{
		Parent: root(),
		ApiId:  apiId,
		Api: &rpc.Api{
			DisplayName: apiId,
			Description: fmt.Sprintf("Description for %s", apiId),
		},
	})
	if err != nil {
		b.Errorf("CreateApi(%s) returned unexpected error: %s", apiId, err)
	}
	return api
}

func updateApi(b *testing.B, ctx context.Context, client connection.Client, apiId string) {
	b.Helper()
	if _, err := client.UpdateApi(ctx, &rpc.UpdateApiRequest{
		Api: &rpc.Api{
			Name:        apiName(apiId),
			DisplayName: fmt.Sprintf("Updated %s", apiId),
		}, UpdateMask: &field_mask.FieldMask{
			Paths: []string{"display_name"},
		},
	}); err != nil {
		b.Errorf("UpdateApi(%s) returned unexpected error: %s", apiId, err)
	}
}

func deleteApi(b *testing.B, ctx context.Context, client connection.Client, apiId string) {
	b.Helper()
	if err := client.DeleteApi(ctx, &rpc.DeleteApiRequest{Name: apiName(apiId)}); err != nil {
		b.Errorf("DeleteApi(%s) returned unexpected error: %s", apiId, err)
	}
}

func BenchmarkGetApi(b *testing.B) {
	ctx, client := setup(b)
	b.Run("GetApi", func(b *testing.B) {
		b.StopTimer()
		for i := 1; i <= b.N; i++ {
			createApi(b, ctx, client, apiId(i))
		}
		b.StartTimer()
		for i := 1; i <= b.N; i++ {
			getApi(b, ctx, client, apiId(i))
		}
	})
	teardown(b, ctx, client)
}

func BenchmarkListApis(b *testing.B) {
	ctx, client := setup(b)
	b.Run("ListApis", func(b *testing.B) {
		b.StopTimer()
		for i := 1; i <= b.N; i++ {
			createApi(b, ctx, client, apiId(i))
		}
		b.StartTimer()
		for i := 1; i <= b.N; i++ {
			listApis(b, ctx, client)
		}
	})
	teardown(b, ctx, client)
}

func BenchmarkCreateApi(b *testing.B) {
	ctx, client := setup(b)
	b.Run("CreateApi", func(b *testing.B) {
		for i := 1; i <= b.N; i++ {
			createApi(b, ctx, client, apiId(i))
		}
	})
	teardown(b, ctx, client)
}

func BenchmarkUpdateApi(b *testing.B) {
	ctx, client := setup(b)
	b.Run("UpdateApi", func(b *testing.B) {
		b.StopTimer()
		for i := 1; i <= b.N; i++ {
			createApi(b, ctx, client, apiId(i))
		}
		b.StartTimer()
		for i := 1; i <= b.N; i++ {
			updateApi(b, ctx, client, apiId(i))
		}
	})
	teardown(b, ctx, client)
}

func BenchmarkDeleteApi(b *testing.B) {
	ctx, client := setup(b)
	b.Run("DeleteApi", func(b *testing.B) {
		b.StopTimer()
		for i := 1; i <= b.N; i++ {
			createApi(b, ctx, client, apiId(i))
		}
		b.StartTimer()
		for i := 1; i <= b.N; i++ {
			deleteApi(b, ctx, client, apiId(i))
		}
	})
	teardown(b, ctx, client)
}
