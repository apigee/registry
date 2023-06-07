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

package rule113

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

func TestAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		in       map[string]string
		expected []*check.Problem
	}{
		{"nil", nil, nil},
		{"empty", map[string]string{}, nil},
		{
			"good",
			map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			nil,
		},
		{
			"bad",
			map[string]string{
				"key": "value",
				"*":   "*",
			},
			[]*check.Problem{
				{
					Message:    `Key "*" has illegal first character '*'.`,
					Suggestion: "Fix key.",
					Severity:   check.Problem_ERROR,
				},
			},
		},
		{
			"big one",
			map[string]string{
				"key": strings.Repeat("x", totalSizeLimit-3),
			},
			nil,
		},
		{
			"too big one",
			map[string]string{
				"key2": strings.Repeat("x", totalSizeLimit-3),
			},
			[]*check.Problem{
				{
					Message:    `Maximum size of all annotations is 256k.`,
					Suggestion: `Reduce size by 1 bytes.`,
					Severity:   check.Problem_ERROR,
				},
			},
		},
		{
			"too big multiple",
			map[string]string{
				"key1": strings.Repeat("x", totalSizeLimit/3),
				"key2": strings.Repeat("x", totalSizeLimit/3),
				"key3": strings.Repeat("x", totalSizeLimit/3),
			},
			[]*check.Problem{
				{
					Message:    `Maximum size of all annotations is 256k.`,
					Suggestion: `Reduce size by 11 bytes.`,
					Severity:   check.Problem_ERROR,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			a := &rpc.ApiSpec{
				Labels: test.in,
			}
			if annotations.OnlyIf(a, fieldName) {
				got := annotations.ApplyToField(ctx, a, fieldName, test.in)
				if diff := cmp.Diff(test.expected, got, cmpopts.IgnoreUnexported(check.Problem{})); diff != "" {
					t.Errorf("Unexpected diff (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestCheckAnnotation(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected []*check.Problem
	}{
		{"good", "alphanum", "value1_2-", nil},
		{"period", "key.", ".", []*check.Problem{
			{
				Message:    `Key "key." contains illegal character '.'.`,
				Suggestion: "Fix key.",
				Severity:   check.Problem_ERROR,
			},
		}},
		{"uppercase", "keY", "valuE", []*check.Problem{
			{
				Message:    `Key "keY" contains illegal character 'Y'.`,
				Suggestion: "Fix key.",
				Severity:   check.Problem_ERROR,
			},
		}},
		{"long", strings.Repeat("y", 63), strings.Repeat("y", 63), nil},
		{"too long", strings.Repeat("n", 64), strings.Repeat("n", 64), []*check.Problem{
			{
				Message:    fmt.Sprintf(`Key %q exceeds max length of 63 characters.`, strings.Repeat("n", 64)),
				Suggestion: "Fix key.",
				Severity:   check.Problem_ERROR,
			},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := checkAnnotation(test.key, test.value)
			if diff := cmp.Diff(test.expected, got, cmpopts.IgnoreUnexported(check.Problem{})); diff != "" {
				t.Errorf("Unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}
