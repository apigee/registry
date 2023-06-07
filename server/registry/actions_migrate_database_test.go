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

	longrunning "cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestMigrateDatabase(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	ctx := context.Background()
	server := defaultTestServer(t)

	req := &rpc.MigrateDatabaseRequest{}
	metadata, err := anypb.New(&rpc.MigrateDatabaseMetadata{})
	if err != nil {
		t.Fatalf("MigrateDatabase(%+v) test failed to build expected response metadata: %s", req, err)
	}
	response, err := anypb.New(&rpc.MigrateDatabaseResponse{
		Message: "OK",
	})
	if err != nil {
		t.Fatalf("MigrateDatabase(%+v) test failed to build expected response message: %s", req, err)
	}
	want := &longrunning.Operation{
		Name:     "migrate",
		Done:     true,
		Metadata: metadata,
		Result:   &longrunning.Operation_Response{Response: response},
	}

	got, err := server.MigrateDatabase(ctx, req)
	if err != nil {
		t.Fatalf("MigrateDatabase(%+v) returned error: %s", req, err)
	}

	opts := cmp.Options{
		protocmp.Transform(),
	}

	if !cmp.Equal(want, got, opts) {
		t.Errorf("MigrateDatabase(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(want, got, opts))
	}
}

func TestMigrateDatabaseKinds(t *testing.T) {
	if adminServiceUnavailable() {
		t.Skip(testRequiresAdminService)
	}
	ctx := context.Background()
	server := defaultTestServer(t)

	tests := []struct {
		desc string
		req  *rpc.MigrateDatabaseRequest
		want codes.Code
	}{
		{
			desc: "migrate with kind not specified",
			req:  &rpc.MigrateDatabaseRequest{Kind: ""},
			want: codes.OK,
		},
		{
			desc: "migrate with kind auto",
			req:  &rpc.MigrateDatabaseRequest{Kind: "auto"},
			want: codes.OK,
		},
		{
			desc: "migrate with kind unsupported",
			req:  &rpc.MigrateDatabaseRequest{Kind: "unsupported"},
			want: codes.InvalidArgument,
		},
		{
			desc: "migrate with kind invalid",
			req:  &rpc.MigrateDatabaseRequest{Kind: "invalid"},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if _, err := server.MigrateDatabase(ctx, test.req); status.Code(err) != test.want {
				t.Errorf("MigrateDatabase(%+v) returned status code %q, want %q: %v", test.req, status.Code(err), test.want, err)
			}
		})
	}
}
