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
	"sync"
	"testing"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/seeder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const concurrency = 20

func isRetryable(c codes.Code) bool {
	return (c == codes.OK) || (c == codes.Unavailable) || (c == codes.Aborted)
}

func TestConcurrentProjectUpdates(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}

	test := struct {
		desc string
		req  *rpc.UpdateProjectRequest
	}{
		desc: "sample",
		req: &rpc.UpdateProjectRequest{
			Project: &rpc.Project{
				Name:        "projects/my-project",
				DisplayName: "sample",
			},
			AllowMissing: true,
		},
	}

	t.Run(test.desc, func(t *testing.T) {
		ctx := context.Background()
		server := defaultTestServer(t)
		var wg sync.WaitGroup
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				_, err := server.UpdateProject(ctx, test.req)
				if !isRetryable(status.Code(err)) {
					t.Errorf("UpdateProject(%+v), wanted a retryable status code, got %q", test.req, status.Code(err))
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
}

func TestConcurrentApiUpdates(t *testing.T) {
	test := struct {
		desc string
		seed *rpc.Project
		req  *rpc.UpdateApiRequest
	}{
		desc: "sample api",
		seed: &rpc.Project{Name: "projects/my-project"},
		req: &rpc.UpdateApiRequest{
			Api:          &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
			AllowMissing: true,
		},
	}

	t.Run(test.desc, func(t *testing.T) {
		ctx := context.Background()
		server := defaultTestServer(t)
		if err := seeder.SeedProjects(ctx, server, test.seed); err != nil {
			t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
		}
		var wg sync.WaitGroup
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				_, err := server.UpdateApi(ctx, test.req)
				if !isRetryable(status.Code(err)) {
					t.Errorf("UpdateApi(%+v), wanted a retryable status code, got %q", test.req, status.Code(err))
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
}

func TestConcurrentApiVersionUpdates(t *testing.T) {
	test := struct {
		desc string
		seed *rpc.Api
		req  *rpc.UpdateApiVersionRequest
	}{
		desc: "sample api",
		seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
		req: &rpc.UpdateApiVersionRequest{
			ApiVersion:   &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/a/versions/v"},
			AllowMissing: true,
		},
	}

	t.Run(test.desc, func(t *testing.T) {
		ctx := context.Background()
		server := defaultTestServer(t)
		if err := seeder.SeedApis(ctx, server, test.seed); err != nil {
			t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
		}
		var wg sync.WaitGroup
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				_, err := server.UpdateApiVersion(ctx, test.req)
				if !isRetryable(status.Code(err)) {
					t.Errorf("UpdateApiVersion(%+v), wanted a retryable status code, got %q", test.req, status.Code(err))
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
}

func TestConcurrentApiSpecUpdates(t *testing.T) {
	test := struct {
		desc string
		seed *rpc.ApiVersion
		req  *rpc.UpdateApiSpecRequest
	}{
		desc: "sample api",
		seed: &rpc.ApiVersion{Name: "projects/my-project/locations/global/apis/a/versions/v"},
		req: &rpc.UpdateApiSpecRequest{
			ApiSpec:      &rpc.ApiSpec{Name: "projects/my-project/locations/global/apis/a/versions/v/specs/s"},
			AllowMissing: true,
		},
	}

	t.Run(test.desc, func(t *testing.T) {
		ctx := context.Background()
		server := defaultTestServer(t)
		if err := seeder.SeedVersions(ctx, server, test.seed); err != nil {
			t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
		}
		var wg sync.WaitGroup
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				_, err := server.UpdateApiSpec(ctx, test.req)
				if !isRetryable(status.Code(err)) {
					t.Errorf("UpdateApiSpec(%+v), wanted a retryable status code, got %q", test.req, status.Code(err))
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
}

func TestConcurrentApiDeploymentUpdates(t *testing.T) {
	test := struct {
		desc string
		seed *rpc.Api
		req  *rpc.UpdateApiDeploymentRequest
	}{
		desc: "sample api",
		seed: &rpc.Api{Name: "projects/my-project/locations/global/apis/a"},
		req: &rpc.UpdateApiDeploymentRequest{
			ApiDeployment: &rpc.ApiDeployment{Name: "projects/my-project/locations/global/apis/a/deployments/d"},
			AllowMissing:  true,
		},
	}

	t.Run(test.desc, func(t *testing.T) {
		ctx := context.Background()
		server := defaultTestServer(t)
		if err := seeder.SeedApis(ctx, server, test.seed); err != nil {
			t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
		}
		var wg sync.WaitGroup
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				_, err := server.UpdateApiDeployment(ctx, test.req)
				if !isRetryable(status.Code(err)) {
					t.Errorf("UpdateApiDeployment(%+v), wanted a retryable status code, got %q", test.req, status.Code(err))
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
}
