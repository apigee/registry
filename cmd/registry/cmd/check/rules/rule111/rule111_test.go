// Copyright 2023 Google LLC.
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

package rule111

import (
	"bytes"
	"compress/gzip"
	"context"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/pkg/application/scoring"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/proto"
)

func TestAddRules(t *testing.T) {
	if err := AddRules(lint.NewRuleRegistry()); err != nil {
		t.Errorf("AddRules got an error: %v", err)
	}
}

func TestMimeTypeContents(t *testing.T) {
	var zbuf bytes.Buffer
	{
		zw, _ := gzip.NewWriterLevel(&zbuf, gzip.BestCompression)
		_, err := zw.Write([]byte("string"))
		if err != nil {
			t.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name     string
		mimeType string
		contents []byte
		problems []*check.Problem
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
			[]*check.Problem{{
				Message:    `Unexpected mime_type "text/html" for contents.`,
				Suggestion: `Detected mime_type: "text/plain; charset=utf-8".`,
				Severity:   check.Problem_WARNING,
			}},
		},
		{
			"empty type",
			"",
			[]byte("string"),
			[]*check.Problem{{
				Message:    `Empty mime_type.`,
				Suggestion: `Detected mime_type: "text/plain; charset=utf-8".`,
				Severity:   check.Problem_ERROR,
			}},
		},
		{
			"bad type",
			"bad/",
			[]byte("string"),
			[]*check.Problem{{
				Message:  `Unable to parse mime_type "bad/": mime: expected token after slash.`,
				Severity: check.Problem_ERROR,
			}},
		},
		{
			"score",
			"application/octet-stream;type=google.cloud.apigeeregistry.v1.scoring.Score",
			createScore(),
			nil,
		},
		{
			"bad gzip in allowed prefix",
			"application/octet-stream+gzip;type=google.cloud.apigeeregistry.v1.scoring.Score",
			createScore(),
			[]*check.Problem{{
				Message:  `Mime "application/octet-stream+gzip;type=google.cloud.apigeeregistry.v1.scoring.Score" indicates gzip, but contents does not have a valid gzip header`,
				Severity: check.Problem_ERROR,
			}},
		},
		{
			"good gzip",
			"application/octet-stream+gzip",
			zbuf.Bytes(),
			nil,
		},
		{
			"bad gzip",
			"application/octet-stream+gzip",
			createScore(),
			[]*check.Problem{{
				Message:  `Mime "application/octet-stream+gzip" indicates gzip, but contents does not have a valid gzip header`,
				Severity: check.Problem_ERROR,
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
					if !mimeTypeContents.OnlyIf(resource, "MimeType") {
						t.Error("invalid OnlyIf")
					}
					got := mimeTypeContents.Apply(context.Background(), resource)
					for i := range test.problems {
						test.problems[i].Location = resource.GetName() + "::MimeType"
					}
					if diff := cmp.Diff(test.problems, got, cmpopts.IgnoreUnexported(check.Problem{})); diff != "" {
						t.Errorf("Unexpected diff (-want +got):\n%s", diff)
					}
				})
			}
		})
	}
}

func createScore() []byte {
	s := &scoring.Score{
		Id:    "score",
		Kind:  "Score",
		Value: &scoring.Score_IntegerValue{},
	}
	b, _ := proto.Marshal(s)
	return b
}
