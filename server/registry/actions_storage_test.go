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
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestGetStorage(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	ctx := context.Background()
	server := defaultTestServer(t)

	req := &emptypb.Empty{}

	resp, err := server.GetStorage(ctx, req)
	if err != nil {
		t.Fatalf("GetStorage(%+v) returned error: %s", req, err)
	}

	// Ensure that we get the set of tables that we expect.
	// Tables should be returned in alphabetical order.
	want := []string{"apis", "artifacts", "blobs", "deployment_revision_tags", "deployments", "projects", "spec_revision_tags", "specs", "versions"}
	got := make([]string, 0)
	for _, c := range resp.Collections {
		got = append(got, c.Name)
	}
	if !cmp.Equal(want, got) {
		t.Errorf("GetStorage(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(want, got, protocmp.Transform()))
	}
}
