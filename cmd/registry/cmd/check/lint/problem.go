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
)

type Severity int32

const (
	INFO Severity = iota
	WARNING
	ERROR
)

func (s Severity) String() string {
	switch s {
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	}
	return "UNKNOWN"
}

// Problem contains information about a result produced by an API Linter.
//
// All rules return []Problem. Most lint rules return 0 or 1 problems, but
// occasionally there are rules that may return more than one.
type Problem struct {
	// Message provides a short description of the problem.
	// This should be no more than a single sentence.
	Message string

	// Suggestion provides a suggested fix, if applicable.
	Suggestion string

	// Location provides the location of the problem.
	// If for a Resource, it is the Resource name.
	// If for a field, this is the Resource name + "::" + field name.
	// The linter sets this automatically.
	Location string

	// RuleID provides the ID of the rule that this problem belongs to.
	// The linter sets this automatically.
	RuleID RuleName

	// Severity provides information on the criticality of the Problem.
	Severity Severity
}

// MarshalJSON defines how to represent a Problem in JSON.
func (p Problem) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.marshal())
}

// MarshalYAML defines how to represent a Problem in YAML.
func (p Problem) MarshalYAML() (interface{}, error) {
	return p.marshal(), nil
}

// Marshal defines how to represent a serialized Problem.
func (p Problem) marshal() interface{} {
	// Return a marshal-able structure.
	return struct {
		Message    string   `json:"message" yaml:"message"`
		Suggestion string   `json:"suggestion,omitempty" yaml:"suggestion,omitempty"`
		Location   string   `json:"location" yaml:"location"`
		RuleID     RuleName `json:"rule_id" yaml:"rule_id"`
		RuleDocURI string   `json:"rule_doc_uri" yaml:"rule_doc_uri"`
		Severity   string   `json:"severity" yaml:"severity"`
	}{
		p.Message,
		p.Suggestion,
		p.Location,
		p.RuleID,
		p.GetRuleURI(),
		p.Severity.String(),
	}
}

// GetRuleURI returns a URI to learn more about the problem.
func (p Problem) GetRuleURI() string {
	return getRuleURL(string(p.RuleID), ruleURLMappings)
}
