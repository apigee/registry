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

package rule109

import (
	"context"
	"fmt"
	"strings"
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

func TestDisplayName(t *testing.T) {
	bad := []*check.Problem{{
		Severity:   check.Problem_ERROR,
		Message:    fmt.Sprintf("%s must contain only UTF-8 characters.", fieldName),
		Suggestion: fmt.Sprintf("Fix %s.", fieldName)}}

	tooLong := []*check.Problem{{
		Severity:   check.Problem_ERROR,
		Message:    fmt.Sprintf("%s exceeds limit of 64 characters.", fieldName),
		Suggestion: fmt.Sprintf("Fix %s.", fieldName)}}

	tests := []struct {
		name     string
		in       string
		expected []*check.Problem
	}{
		{"empty", "", nil},
		{"invalid", string([]byte{0xff}), bad},
		{"long", strings.Repeat("x", 64), nil},
		{"too long", strings.Repeat("y", 65), tooLong},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			a := &rpc.ApiSpec{
				Description: test.in,
			}
			if displayName.OnlyIf(a, fieldName) {
				got := displayName.ApplyToField(ctx, a, fieldName, test.in)
				if diff := cmp.Diff(test.expected, got, cmpopts.IgnoreUnexported(check.Problem{})); diff != "" {
					t.Errorf("Unexpected diff (-want +got):\n%s", diff)
				}
			}
		})
	}
}
