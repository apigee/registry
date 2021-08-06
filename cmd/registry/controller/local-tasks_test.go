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

package controller

import (
	"context"
	"strings"
	"testing"
)

// Test that GetCommand() is returning the correct instance for each command string
func TestGetCommand(t *testing.T) {
	tests := []struct {
		desc   string
		action string
		want   string
	}{
		{
			desc:   "annotate action",
			action: "annotate projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml owner=registry",
			want:   "annotate",
		},
		{
			desc:   "compute action",
			action: "compute lint projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml",
			want:   "compute",
		},
		{
			desc:   "delete action",
			action: "delete projects/demo/apis/-/versions/-/specs/-",
			want:   "delete",
		},
		{
			desc:   "export action",
			action: "export csv projects/demo/apis/petstore/versions/1.0.0",
			want:   "export",
		},
		{
			desc:   "get action",
			action: "get projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml",
			want:   "get",
		},
		{
			// Index is currently experimental
			// This is a placeholder test case not necessarily the right usage
			desc:   "index action",
			action: "index projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml",
			want:   "index",
		},
		{
			desc:   "label action",
			action: "label projects/demo/apis/petstore/versions/1.0.0/specs/openapi.yaml env=test --overwrite",
			want:   "label",
		},
		{
			desc:   "list action",
			action: "list projects/demo/apis/-/versions/-/specs/-",
			want:   "list",
		},
		{
			// Search is currently experimental
			// This is a placeholder test case not necessarily the right usage
			desc:   "search action",
			action: "search 'API'",
			want:   "search",
		},
		{
			desc:   "upload action",
			action: "upload manifest cmd/registry/controller/testdata/manifest_e2e.yaml --project_id=demo",
			want:   "upload",
		},
		{
			desc:   "vocabulary action",
			action: "vocabulary union projects/demo/apis/petstore/versions/2.0.0/specs/openapi.yaml/artifacts/vocabulary projects/demo/apis/petstore/versions/5.0.0/specs/openapi.yaml/artifacts/vocabulary",
			want:   "vocabulary",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			cmd, err := getCommand(ctx, strings.Fields(test.action))
			if err != nil {
				t.Fatalf("Error executing GetCommand(%q) %s", test.action, err)
			}

			if cmd.Name() != test.want {
				t.Errorf("GetCommand(%q) created unexpected Command instance. Want: %s, got: %s", test.action, test.want, cmd.Name())
			}
		})
	}
}

// Test the error scenarios for GetCommand()
func TestErrorCases(t *testing.T) {
	tests := []struct {
		desc   string
		action string
		want   string
	}{
		{
			desc:   "resolve action",
			action: "resolve projects/demo/artifacts/test-manifest",
		},
		{
			desc:   "invalid action",
			action: "random projects/demo/apis/petstore",
		},
		{
			desc:   "empty action",
			action: "",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ctx := context.Background()
			_, err := getCommand(ctx, strings.Fields(test.action))
			if err == nil {
				t.Errorf("Expected GetCommand() to return error.")
			}
		})
	}
}
