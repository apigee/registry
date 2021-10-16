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

import (
	"fmt"
)

// CreateLinter returns a Linter object when provided the name of a linter
func CreateLinter(linter_name string) (Linter, error) {
	if linter_name == "spectral" {
		return NewSpectralLinter(), nil
	} else if linter_name == "api-linter" {
		return NewApiLinter(), nil
	}

	return nil, fmt.Errorf("unknown linter: %s", linter_name)
}
