// Copyright 2023 Google LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package patch

import (
	"context"
	"testing"

	"github.com/apigee/registry/cmd/registry/compress"
	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry/test/seeder"
)

func TestAutomaticallyZippedSpecs(t *testing.T) {
	ctx := context.Background()
	registryClient, adminClient := grpctest.SetupRegistry(ctx, t, "patch-autozip-test", []seeder.RegistryResource{
		&rpc.Project{
			Name: "projects/patch-autozip-test",
		},
	})
	testpath := "testdata/sample-protos"
	if err := Apply(ctx, registryClient, adminClient, nil, "projects/patch-autozip-test/locations/global", true, 1, testpath); err != nil {
		t.Fatalf("Apply() failed with error %s", err)
	}
	// verify the spec
	specname := "projects/patch-autozip-test/locations/global/apis/apigeeregistry/versions/v1/specs/protos"
	spec, err := registryClient.GetApiSpec(ctx, &rpc.GetApiSpecRequest{
		Name: specname,
	})
	if err != nil {
		t.Fatalf("Failed to get applied spec: %s", err)
	}
	if spec.SizeBytes == 0 {
		t.Fatalf("Applied spec is empty")
	}
	contents, err := registryClient.GetApiSpecContents(ctx, &rpc.GetApiSpecContentsRequest{
		Name: specname,
	})
	if err != nil {
		t.Fatalf("Failed to get contents of applied spec: %s", err)
	}
	mimeType := "application/x.proto+zip"
	if contents.ContentType != mimeType {
		t.Errorf("Expected content type of %s, got %s", mimeType, contents.ContentType)
	}
	files, err := compress.UnzipArchiveToMap(contents.Data)
	if err != nil {
		t.Fatalf("Failed to unzip contents of applied spec: %s", err)
	}
	fileCount := 13
	if len(files) != fileCount {
		t.Errorf("Expected %d files, got %d", fileCount, len(files))
	}
}
