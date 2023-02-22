// Copyright 2023 Google LLC
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

package rule1002

import (
	"context"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/pkg/application/apihub"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestAddRules(t *testing.T) {
	if err := AddRules(lint.NewRuleRegistry()); err != nil {
		t.Errorf("AddRules got an error: %v", err)
	}
}

func TestInternalMimeTypeContents(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		bytes    []byte
		problems []*check.Problem
	}{
		{
			"empty contents",
			"application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.Lifecycle",
			[]byte{},
			nil,
		},
		{
			"good contents",
			"application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.Lifecycle",
			marshalMessage(t,
				&apihub.Lifecycle{
					Description: "Lifecycle",
				},
			),
			nil,
		},
		{
			"unknown type",
			"application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.UnknownType",
			marshalMessage(t,
				&apihub.Lifecycle{
					Description: "Lifecycle",
				},
			),
			[]*check.Problem{{
				Message:    `Error loading contents into proto type Lifecycle.`,
				Suggestion: `Fix mime_type.`,
				Severity:   check.Problem_ERROR,
			}},
		},
		{
			"not a message",
			"application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.Lifecycle",
			[]byte("not a score"),
			[]*check.Problem{{
				Message:    `Unknown internal mime_type: "application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.Lifecycle".`,
				Suggestion: `Fix mime_type or contents.`,
				Severity:   check.Problem_ERROR,
			}},
		},
		{
			"wrong mime type",
			"application/octet-stream;type=google.cloud.apigeeregistry.v1.apihub.ReferenceList",
			marshalMessage(t,
				&apihub.Lifecycle{
					Description: "Lifecycle",
				},
			),
			// TODO: Would like a Problem, but Unmarshal is too lenient and this verification is impossible?
			nil,
			// []*check.Problem{{
			// 	Message:    `Error loading contents into proto type Lifecycle.`,
			// 	Suggestion: `Fix mime_type or contents.`,
			// 	Severity:   check.Problem_ERROR,
			// 	Location:   "Artifact::MimeType",
			// }},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			artifact := &rpc.Artifact{
				Name:     "Artifact",
				MimeType: test.mimeType,
				Contents: test.bytes,
			}
			spec := &rpc.ApiSpec{
				Name:     "Spec",
				MimeType: test.mimeType,
				Contents: test.bytes,
			}

			opts := []cmp.Option{
				cmpopts.IgnoreFields(check.Problem{}, "Message", "Location"),
				cmpopts.IgnoreUnexported(check.Problem{}),
			}
			if internalMimeTypeContents.OnlyIf(artifact, "MimeType") {
				result := internalMimeTypeContents.Apply(context.Background(), spec)
				if diff := cmp.Diff(test.problems, result, opts...); diff != "" {
					t.Errorf("Unexpected diff (-want +got):\n%s", diff)
				}

				result = internalMimeTypeContents.Apply(context.Background(), artifact)
				if diff := cmp.Diff(test.problems, result, opts...); diff != "" {
					t.Errorf("Unexpected diff (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func marshalMessage(t *testing.T, message protoreflect.ProtoMessage) []byte {
	var bytes, err = proto.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}
