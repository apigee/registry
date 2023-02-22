// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lint

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/apigee/registry/pkg/application/check"
	"gopkg.in/yaml.v2"
)

func TestProblemJSON(t *testing.T) {
	problem := &check.Problem{
		Message:  "foo bar",
		Location: "test/location",
		RuleId:   "core::0131",
	}
	serialized, err := json.Marshal(problem)
	if err != nil {
		t.Fatalf("Could not marshal Problem to JSON.")
	}
	tests := []struct {
		testName string
		token    string
	}{
		{"Message", `"message":"foo bar"`},
		{"Location", `"location":"test/location"`},
		{"RuleId", `"rule_id":"core::0131"`},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			if !strings.Contains(string(serialized), test.token) {
				t.Errorf("Got\n%v\nExpected `%s` to be present.", string(serialized), test.token)
			}
		})
	}
}

func TestProblemYAML(t *testing.T) {
	problem := &check.Problem{
		Message:  "foo bar",
		Location: "test/location",
		RuleId:   "core::0131",
		Severity: check.Problem_ERROR,
	}
	serialized, err := yaml.Marshal(problem)
	if err != nil {
		t.Fatalf("Could not marshal Problem to YAML.")
	}
	tests := []struct {
		testName string
		token    string
	}{
		{"Message", `message: foo bar`},
		{"Location", `location: test/location`},
		{"RuleId", `ruleid: core::0131`},
		{"Severity", fmt.Sprintf(`severity: %d`, check.Problem_ERROR.Number())},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			if !strings.Contains(string(serialized), test.token) {
				t.Errorf("Got\n%v\nExpected `%s` to be present.", string(serialized), test.token)
			}
		})
	}
}
