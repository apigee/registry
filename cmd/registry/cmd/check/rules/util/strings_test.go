// Copyright 2023 Google LLC. All Rights Reserved.
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

package util

import (
	"strings"
	"testing"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestCheckUTF(t *testing.T) {
	bad := []lint.Problem{{
		Severity:   lint.ERROR,
		Message:    `xxx must contain only UTF-8 characters.`,
		Suggestion: `Fix xxx.`}}

	tooLong := []lint.Problem{{
		Severity:   lint.ERROR,
		Message:    `xxx exceeds limit of 64 characters.`,
		Suggestion: `Fix xxx.`}}

	tests := []struct {
		name     string
		in       string
		expected []lint.Problem
	}{
		{"empty", "", nil},
		{"invalid", string([]byte{0xff}), bad},
		{"long", strings.Repeat("y", 64), nil},
		{"too long", strings.Repeat("n", 65), tooLong},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CheckUTF("xxx", test.in, 64)
			if diff := cmp.Diff(test.expected, got, cmpopts.IgnoreUnexported(lint.Problem{})); diff != "" {
				t.Errorf("Unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}
