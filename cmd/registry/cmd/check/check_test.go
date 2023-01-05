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

package check

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/apigee/registry/server/registry/test/seeder"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if APG_REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestCheck(t *testing.T) {
	ctx := context.Background()
	registryClient, err := connection.NewRegistryClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %+v", err)
	}
	t.Cleanup(func() { registryClient.Close() })
	adminClient, err := connection.NewAdminClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %+v", err)
	}
	t.Cleanup(func() { adminClient.Close() })
	client := seeder.Client{
		RegistryClient: registryClient,
		AdminClient:    adminClient,
	}
	spec := &rpc.ApiSpec{
		Name:     "projects/my-project/locations/global/apis/a/versions/v/specs/bad",
		MimeType: "application/html",
		Contents: []byte("some text"),
	}
	if err := seeder.SeedSpecs(ctx, client, spec); err != nil {
		t.Fatalf("Setup/Seeding: Failed to seed registry: %s", err)
	}

	buf := &bytes.Buffer{}
	cmd := Command()
	args := []string{"projects/my-project"}
	cmd.SetArgs(args)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() with args %v returned error: %s", args, err)
	}
	if !strings.Contains(buf.String(), `Unexpected mime_type "application/html"`) {
		t.Errorf("unexpected result: %s", buf)
	}
}
