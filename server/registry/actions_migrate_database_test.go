// Copyright 2020 Google LLC. All Rights Reserved.
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

	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/genproto/googleapis/longrunning"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestMigrateDatabase(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)

	req := &rpc.MigrateDatabaseRequest{}
	metadata, err := anypb.New(&rpc.MigrateDatabaseMetadata{})
	if err != nil {
		t.Fatalf("MigrateDatabase(%+v) test failed to build expected metadata: %s", req, err)
	}
	response, err := anypb.New(&rpc.MigrateDatabaseResponse{
		Message: "OK",
	})
	if err != nil {
		t.Fatalf("MigrateDatabase(%+v) test failed to build expected response: %s", req, err)
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
		// Ignore fields that are build-dependent or excluded from Go 	< 1.18.
		protocmp.IgnoreFields(&rpc.Status{}, "build"),
	}

	if !cmp.Equal(want, got, opts) {
		t.Errorf("MigrateDatabase(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(want, got, opts))
	}
}

func TestMigrateDatabaseKinds(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)

	tests := []struct {
		desc string
		req  *rpc.MigrateDatabaseRequest
		want codes.Code
	}{
		{
			req:  &rpc.MigrateDatabaseRequest{Kind: "auto"},
			want: codes.OK,
		},
		{
			req:  &rpc.MigrateDatabaseRequest{Kind: ""},
			want: codes.OK,
		},
		{
			req:  &rpc.MigrateDatabaseRequest{Kind: "unsupported"},
			want: codes.InvalidArgument,
		},
		{
			req:  &rpc.MigrateDatabaseRequest{Kind: "invalid"},
			want: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		_, err := server.MigrateDatabase(ctx, test.req)
		if test.want == codes.OK {
			if err != nil {
				t.Fatalf("MigrateDatabase(%+v) returned error: %s (expected %s)", test.req, err, test.want)
			}
		} else {
			if err == nil {
				t.Fatalf("MigrateDatabase(%+v) succeeded (expected %s)", test.req, test.want)
			} else if status.Code(err) != codes.InvalidArgument {
				t.Fatalf("MigrateDatabase(%+v) returned error: %s (expected %s)", test.req, err, test.want)
			}
		}
	}
}
