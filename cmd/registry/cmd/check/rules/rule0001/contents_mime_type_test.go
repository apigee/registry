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
	"fmt"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/proto"
)

func TestContentsMimeType_general(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		contents []byte
		problems []lint.Problem
	}{
		{
			"empty content",
			"text/plain",
			nil,
			nil,
		},
		{
			"type match",
			"text/plain",
			[]byte("string"),
			nil,
		},
		{
			"type and parameter match",
			"text/plain;charset=utf-8",
			[]byte("string"),
			nil,
		},
		{
			"wrong type",
			"text/html",
			[]byte("string"),
			[]lint.Problem{{
				Message:    `Unexpected mime_type "text/html" for contents.`,
				Suggestion: `Detected mime_type "text/plain; charset=utf-8" for contents.`,
				Severity:   lint.WARNING,
			}},
		},
		{
			"empty type",
			"",
			[]byte("string"),
			[]lint.Problem{{
				Message:    `Unexpected mime_type "" for contents.`,
				Suggestion: `Detected mime_type "text/plain; charset=utf-8" for contents.`,
				Severity:   lint.WARNING,
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
					for i := range test.problems {
						test.problems[i].Location = resource.GetName() + "::MimeType"
					}
					if diff := cmp.Diff(test.problems, got, cmpopts.IgnoreUnexported(lint.Problem{})); diff != "" {
						t.Errorf("Unexpected diff (-want +got):\n%s", diff)
					}
				})
			}
		})
	}
}

func TestContentsMimeType_score(t *testing.T) {
	const scoreType = "application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.Score"
	var score, _ = proto.Marshal(&rpc.Score{
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

	resource := &rpc.Artifact{
		Name:     "Artifact",
		MimeType: scoreType,
		Contents: score,
	}

	want := []lint.Problem{{
		Message:    fmt.Sprintf(`Unexpected contents %q for mime_type.`, scoreType),
		Suggestion: `Internal contents may be corrupted.`,
		Severity:   lint.ERROR,
		Location:   "Artifact::MimeType",
	}}

	got := contentsMimeType.Apply(resource)

	if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(lint.Problem{})); diff != "" {
		t.Errorf("Unexpected diff (-want +got):\n%s", diff)
	}
}
