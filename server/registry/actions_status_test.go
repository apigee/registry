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
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestGetStatus(t *testing.T) {
	ctx := context.Background()
	server := defaultTestServer(t)

	req := &emptypb.Empty{}
	want := &rpc.Status{
		Message: "running",
	}

	got, err := server.GetStatus(ctx, req)
	if err != nil {
		t.Fatalf("GetStatus(%+v) returned error: %s", req, err)
	}

	opts := cmp.Options{
		protocmp.Transform(),
		// Ignore fields that are build-dependent or excluded from Go 	< 1.18.
		protocmp.IgnoreFields(&rpc.Status{}, "build"),
	}

	if !cmp.Equal(want, got, opts) {
		t.Errorf("GetStatus(%+v) returned unexpected diff (-want +got):\n%s", req, cmp.Diff(want, got, opts))
	}
}
