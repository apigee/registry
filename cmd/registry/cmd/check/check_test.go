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
	"google.golang.org/protobuf/encoding/protojson"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if APG_REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestCheck(t *testing.T) {
	var score, _ = protojson.Marshal(&rpc.Score{
		Id:   "score",
		Kind: "Score",
		Value: &rpc.Score_IntegerValue{
			IntegerValue: &rpc.IntegerValue{
				Value:    1,
				MinValue: 0,
				MaxValue: 10,
			},
		},
	})
	artifacts := []*rpc.Artifact{{
		Name:     "projects/my-project/locations/global/apis/b/deployments/d/artifacts/bad",
		MimeType: "application/html",
		Contents: score,
	}}
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
	if err := seeder.SeedArtifacts(ctx, client, artifacts...); err != nil {
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
	if !strings.Contains(buf.String(), `message: Unexpected mime_type "application/html" for contents`) {
		t.Errorf("unexpected result: %s", buf)
	}
}
