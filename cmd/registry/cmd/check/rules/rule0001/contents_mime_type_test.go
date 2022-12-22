// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rule0001

import (
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestContentsMimeType(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		contents []byte
		problems []lint.Problem
	}{
		{
			"Partial match",
			"text/plain",
			[]byte("string"),
			nil,
		},
		{
			"Complete match",
			"text/plain; charset=utf-8",
			[]byte("string"),
			nil,
		},
		{
			"Failed match",
			"text/html",
			[]byte("string"),
			[]lint.Problem{{
				Message:    `Unexpected mime type: "text/html"`,
				Suggestion: `Expected mime type: "text/plain; charset=utf-8"`,
				Severity:   1,
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resources := []lint.Resource{
				&rpc.ApiSpec{
					Name:     "ApiSpec",
					MimeType: test.mimeType,
					Contents: test.contents,
				},
				&rpc.Artifact{
					Name:     "Artifact",
					MimeType: test.mimeType,
					Contents: test.contents,
				},
			}
			for _, resource := range resources {
				t.Run(resource.GetName(), func(t *testing.T) {
					if !contentsMimeType.OnlyIf(resource, "MimeType") {
						t.Error("invalid OnlyIf")
					}
					got := contentsMimeType.Apply(resource)
					if diff := cmp.Diff(test.problems, got, cmpopts.IgnoreUnexported(lint.Problem{})); diff != "" {
						t.Errorf("Unexpected diff (-want +got):\n%s", diff)
					}
				})
			}
		})
	}
}
