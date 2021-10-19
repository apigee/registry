// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conformance

import "github.com/apigee/registry/rpc"

// Linter is an interface to lint specs in the registry
type Linter interface {
	// Add a new rule to the linter.
	AddRule(mimeType string, rule string) error

	// Gets the name of the linter.
	GetName() string

	// Returns whether the linter supports the provided mime type.
	SupportsMimeType(mimeType string) bool

	// Lints a provided specification of given mime type and returns a
	// LintFile object.
	LintSpec(mimeType string, specPath string) ([]*rpc.LintProblem, error)
}
