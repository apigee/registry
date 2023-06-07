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

package benchmark

import (
	"context"
	"fmt"
	"testing"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/rpc"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/protobuf/field_mask"
)

func apiId(i int) string {
	return fmt.Sprintf("test-%d", i)
}

func apiName(apiId string) string {
	return fmt.Sprintf("%s/apis/%s", root().String()+"/locations/global", apiId)
}

func getApi(b *testing.B, ctx context.Context, client connection.RegistryClient, apiId string) error {
	b.Helper()
	_, err := client.GetApi(ctx, &rpc.GetApiRequest{Name: apiName(apiId)})
	return err
}

func listApis(b *testing.B, ctx context.Context, client connection.RegistryClient) error {
	b.Helper()
	it := client.ListApis(ctx, &rpc.ListApisRequest{Parent: root().String() + "/locations/global"})
	for _, err := it.Next(); err != iterator.Done; _, err = it.Next() {
		if err != nil {
			return err
		}
	}
	return nil
}

func updateApi(b *testing.B, ctx context.Context, client connection.RegistryClient, apiId string) error {
	b.Helper()
	_, err := client.UpdateApi(ctx, &rpc.UpdateApiRequest{
		Api: &rpc.Api{
			Name:        apiName(apiId),
			DisplayName: fmt.Sprintf("Updated %s", apiId),
		}, UpdateMask: &field_mask.FieldMask{
			Paths: []string{"display_name"},
		},
	})
	return err
}

func deleteApi(b *testing.B, ctx context.Context, client connection.RegistryClient, apiId string) error {
	b.Helper()
	return client.DeleteApi(ctx, &rpc.DeleteApiRequest{Name: apiName(apiId)})
}

func BenchmarkGetApi(b *testing.B) {
	ctx, client := setup(b)
	if err := createApi(b, ctx, client, root().Api(apiId(1))); err != nil {
		b.Fatalf("%s", err)
	}
	b.Run("GetApi", func(b *testing.B) {
		for i := 1; i <= b.N; i++ {
			if err := getApi(b, ctx, client, apiId(1)); err != nil {
				b.Fatalf("%s", err)
			}
		}
	})
	teardown(ctx, b, client)
}

func BenchmarkListApis(b *testing.B) {
	ctx, client := setup(b)
	for i := 1; i <= 10; i++ {
		if err := createApi(b, ctx, client, root().Api(apiId(i))); err != nil {
			b.Fatalf("%s", err)
		}
	}
	b.Run("ListApis", func(b *testing.B) {
		for i := 1; i <= b.N; i++ {
			if err := listApis(b, ctx, client); err != nil {
				b.Fatalf("%s", err)
			}
		}
	})
	teardown(ctx, b, client)
}

func BenchmarkCreateApi(b *testing.B) {
	ctx, client := setup(b)
	b.Run("CreateApi", func(b *testing.B) {
		for i := 1; i <= b.N; i++ {
			if err := createApi(b, ctx, client, root().Api(apiId(i))); err != nil {
				b.Fatalf("%s", err)
			}
		}
		b.StopTimer()
		for i := 1; i <= b.N; i++ {
			if err := deleteApi(b, ctx, client, apiId(i)); err != nil {
				b.Fatalf("%s", err)
			}
		}
		b.StartTimer()
	})
	teardown(ctx, b, client)
}

func BenchmarkUpdateApi(b *testing.B) {
	ctx, client := setup(b)
	for i := 1; i <= 1; i++ {
		if err := createApi(b, ctx, client, root().Api(apiId(i))); err != nil {
			b.Fatalf("%s", err)
		}
	}
	b.Run("UpdateApi", func(b *testing.B) {
		for i := 1; i <= b.N; i++ {
			if err := updateApi(b, ctx, client, apiId(1)); err != nil {
				b.Fatalf("%s", err)
			}
		}
	})
	teardown(ctx, b, client)
}

func BenchmarkDeleteApi(b *testing.B) {
	ctx, client := setup(b)
	b.Run("DeleteApi", func(b *testing.B) {
		b.StopTimer()
		for i := 1; i <= b.N; i++ {
			if err := createApi(b, ctx, client, root().Api(apiId(i))); err != nil {
				b.Fatalf("%s", err)
			}
		}
		b.StartTimer()
		for i := 1; i <= b.N; i++ {
			if err := deleteApi(b, ctx, client, apiId(i)); err != nil {
				b.Fatalf("%s", err)
			}
		}
	})
	teardown(ctx, b, client)
}
