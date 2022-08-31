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

package registry

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/seeder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestConcurrentApiUpdates(t *testing.T) {
	test := struct {
		desc string
		seed *rpc.Project
		req  *rpc.UpdateApiRequest
	}{
		desc: "sample api",
		seed: &rpc.Project{Name: "projects/my-project"},
		req: &rpc.UpdateApiRequest{
			Api:          &rpc.Api{Name: "projects/my-project/locations/global/apis/api"},
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
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
				_, err := server.UpdateApi(ctx, test.req)
				code := status.Code(err)
				switch code {
				case codes.OK:
				case codes.Unavailable:
				default:
					t.Errorf("UpdateApi(%+v) returned status code %q", test.req, status.Code(err))
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
}
