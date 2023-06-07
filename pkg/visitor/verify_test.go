// Copyright 2023 Google LLC.
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

package visitor

import (
	"context"
	"testing"

	"github.com/apigee/registry/pkg/connection/grpctest"

	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/seeder"
)

func TestVerify(t *testing.T) {
	ctx := context.Background()
	registryClient, adminClient := grpctest.SetupRegistry(ctx, t, "content-test", []seeder.RegistryResource{
		&rpc.Project{Name: "projects/verify-test"},
	})
	t.Cleanup(func() {
		if err := adminClient.DeleteProject(ctx, &rpc.DeleteProjectRequest{Name: "projects/verify-test"}); err != nil {
			t.Fatalf("failed to delete test project: %s", err)
		}
	})
	t.Run("verify-project-existing", func(t *testing.T) {
		err := VerifyLocation(ctx, registryClient, "projects/verify-test/locations/global")
		if err != nil {
			t.Errorf("Failed to verify existing project: %s", err)
		}
	})
	t.Run("verify-project-missing", func(t *testing.T) {
		err := VerifyLocation(ctx, registryClient, "projects/verify-test-missing/locations/global")
		if err == nil {
			t.Errorf("Failed to detect missing project: %s", err)
		}
	})
	t.Run("verify-project-invalid", func(t *testing.T) {
		err := VerifyLocation(ctx, registryClient, "projects/verify-test/locations/invalid")
		if err == nil {
			t.Errorf("Failed to detect invalid project: %s", err)
		}
	})
}
