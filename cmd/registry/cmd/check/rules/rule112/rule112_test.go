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

package rule112

import (
	"context"
	"fmt"
	"strconv"
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

func TestLabels(t *testing.T) {
	many := make(map[string]string, 64)
	for i := 0; i < 64; i++ {
		many["key"+strconv.Itoa(i)] = "value" + strconv.Itoa(i)
	}
	tooMany := make(map[string]string, 65)
	for k, v := range many {
		tooMany[k] = v
	}
	tooMany["final"] = "straw"
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
				{
					Message:    `Value for key "*" contains illegal character '*'.`,
					Suggestion: "Fix value.",
					Severity:   check.Problem_ERROR,
				},
			},
		},
		{
			"many",
			many,
			nil,
		},
		{
			"too many",
			tooMany,
			[]*check.Problem{
				{
					Message:    `Maximum number of labels is 64.`,
					Suggestion: "Delete some entries.",
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
			if labels.OnlyIf(a, fieldName) {
				got := labels.ApplyToField(ctx, a, fieldName, test.in)
				if diff := cmp.Diff(test.expected, got, cmpopts.IgnoreUnexported(check.Problem{})); diff != "" {
					t.Errorf("Unexpected diff (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestCheckLabel(t *testing.T) {
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
			{
				Message:    `Value for key "key." contains illegal character '.'.`,
				Suggestion: "Fix value.",
				Severity:   check.Problem_ERROR,
			},
		}},
		{"uppercase", "keY", "valuE", []*check.Problem{
			{
				Message:    `Key "keY" contains illegal character 'Y'.`,
				Suggestion: "Fix key.",
				Severity:   check.Problem_ERROR,
			},
			{
				Message:    `Value for key "keY" contains illegal character 'E'.`,
				Suggestion: "Fix value.",
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
			{
				Message:    fmt.Sprintf(`Value for key %q exceeds max length of 63 characters.`, strings.Repeat("n", 64)),
				Suggestion: "Fix value.",
				Severity:   check.Problem_ERROR,
			},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := checkLabel(test.key, test.value)
			if diff := cmp.Diff(test.expected, got, cmpopts.IgnoreUnexported(check.Problem{})); diff != "" {
				t.Errorf("Unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}
