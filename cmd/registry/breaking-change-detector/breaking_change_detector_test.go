// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package breakingchangedetector

import (
	"testing"

	"github.com/apigee/registry/pkg/connection/grpctest"
	"github.com/apigee/registry/rpc"
	"github.com/apigee/registry/server/registry"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
)

// TestMain will set up a local RegistryServer and grpc.Server for all
// tests in this package if APG_REGISTRY_ADDRESS env var is not set
// for the client.
func TestMain(m *testing.M) {
	grpctest.TestMain(m, registry.Config{})
}

func TestChanges(t *testing.T) {
	tests := []struct {
		desc      string
		diffProto *rpc.Diff
		wantProto *rpc.ChangeDetails
	}{
		{
			desc: "Components.Required field Addition Breaking Test",
			diffProto: &rpc.Diff{
				Additions: []string{"components.schemas.x.required.x"},
			},
			wantProto: &rpc.ChangeDetails{
				BreakingChanges: &rpc.Diff{
					Additions: []string{"components.schemas.x.required.x"},
				},
				NonBreakingChanges: &rpc.Diff{},
				UnknownChanges:     &rpc.Diff{},
			},
		},
		{
			desc: "Components.Schemas field Deletion Breaking Test",
			diffProto: &rpc.Diff{
				Deletions: []string{"components.schemas.x.x"},
			},
			wantProto: &rpc.ChangeDetails{
				BreakingChanges: &rpc.Diff{
					Deletions: []string{"components.schemas.x.x"},
				},
				NonBreakingChanges: &rpc.Diff{},
				UnknownChanges:     &rpc.Diff{},
			},
		},
		{
			desc: "Components.Schema.Type field Modification Breaking Test",
			diffProto: &rpc.Diff{
				Modifications: map[string]*rpc.Diff_ValueChange{
					"components.schemas.x.properties.type": {
						To:   "float",
						From: "int64",
					},
				},
			},
			wantProto: &rpc.ChangeDetails{
				BreakingChanges: &rpc.Diff{
					Modifications: map[string]*rpc.Diff_ValueChange{
						"components.schemas.x.properties.type": {
							To:   "float",
							From: "int64",
						},
					},
				},
				NonBreakingChanges: &rpc.Diff{},
				UnknownChanges:     &rpc.Diff{},
			},
		},
		{
			desc: "Info field Addition NonBreaking Test",
			diffProto: &rpc.Diff{
				Additions: []string{"info.x.x"},
			},
			wantProto: &rpc.ChangeDetails{
				BreakingChanges: &rpc.Diff{},
				NonBreakingChanges: &rpc.Diff{
					Additions: []string{"info.x.x"},
				},
				UnknownChanges: &rpc.Diff{},
			},
		},
		{
			desc: "Info field Deletion NonBreaking Test",
			diffProto: &rpc.Diff{
				Deletions: []string{"info.x.x"},
			},
			wantProto: &rpc.ChangeDetails{
				BreakingChanges: &rpc.Diff{},
				NonBreakingChanges: &rpc.Diff{
					Deletions: []string{"info.x.x"},
				},
				UnknownChanges: &rpc.Diff{},
			},
		},
		{
			desc: "Info field Modification NonBreaking Test",
			diffProto: &rpc.Diff{
				Modifications: map[string]*rpc.Diff_ValueChange{
					"info.x.x.x": {
						To:   "to",
						From: "from",
					},
				},
			},
			wantProto: &rpc.ChangeDetails{
				BreakingChanges: &rpc.Diff{},
				NonBreakingChanges: &rpc.Diff{
					Modifications: map[string]*rpc.Diff_ValueChange{
						"info.x.x.x": {
							To:   "to",
							From: "from",
						},
					},
				},
				UnknownChanges: &rpc.Diff{},
			},
		},
		{
			desc: "Components.Schemas field Addition NonBreaking Test",
			diffProto: &rpc.Diff{
				Additions: []string{"components.schemas.x.x"},
			},
			wantProto: &rpc.ChangeDetails{
				BreakingChanges: &rpc.Diff{},
				NonBreakingChanges: &rpc.Diff{
					Additions: []string{"components.schemas.x.x"},
				},
				UnknownChanges: &rpc.Diff{},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotProto := GetChangeDetails(test.diffProto)
			opts := cmp.Options{
				protocmp.Transform(),
				cmpopts.SortSlices(func(a, b string) bool { return a < b }),
			}
			if !cmp.Equal(test.wantProto, gotProto, opts) {
				t.Errorf("GetDiff returned unexpected diff (-want +got):\n%s", cmp.Diff(test.wantProto, gotProto, opts))
			}
		})
	}
}
