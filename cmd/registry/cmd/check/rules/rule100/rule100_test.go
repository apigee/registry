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

package rule100

import (
	"context"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/rpc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestAddRules(t *testing.T) {
	if err := AddRules(lint.NewRuleRegistry()); err != nil {
		t.Errorf("AddRules got an error: %v", err)
	}
}

func Test_availabilitySingleWord(t *testing.T) {
	prob := []*check.Problem{{
		Severity:   check.Problem_INFO,
		Message:    `Availability is free-form, but we expect single words that describe availability.`,
		Suggestion: `Use single words like: "NONE", "TESTING", "PREVIEW", "GENERAL", "DEPRECATED", "SHUTDOWN"`,
	}}

	for _, tt := range []struct {
		in       string
		expected []*check.Problem
	}{
		{"", nil},
		{"x", nil},
		{"x x", prob},
	} {
		api := &rpc.Api{
			Availability: tt.in,
		}
		if availabilitySingleWord.OnlyIf(api) {
			got := availabilitySingleWord.ApplyToApi(context.Background(), api)
			if diff := cmp.Diff(got, tt.expected, cmpopts.IgnoreUnexported(check.Problem{})); diff != "" {
				t.Errorf("unexpected diff: (-want +got):\n%s", diff)
			}
		}
	}
}
