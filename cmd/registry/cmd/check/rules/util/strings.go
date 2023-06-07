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

package util

import (
	"fmt"
	"unicode/utf8"

	"github.com/apigee/registry/pkg/application/check"
)

func CheckUTF(fieldName string, value interface{}, maxLen int) []*check.Problem {
	v := value.(string)
	if !utf8.ValidString(v) {
		return []*check.Problem{{
			Severity:   check.Problem_ERROR,
			Message:    fmt.Sprintf("%s must contain only UTF-8 characters.", fieldName),
			Suggestion: fmt.Sprintf("Fix %s.", fieldName),
		}}
	}
	if utf8.RuneCountInString(v) > maxLen {
		return []*check.Problem{{
			Severity:   check.Problem_ERROR,
			Message:    fmt.Sprintf("%s exceeds limit of %d characters.", fieldName, maxLen),
			Suggestion: fmt.Sprintf("Fix %s.", fieldName),
		}}
	}
	return nil
}
