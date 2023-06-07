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
	"net/url"

	"github.com/apigee/registry/pkg/application/check"
)

func CheckURI(field, uri string) []*check.Problem {
	if uri != "" {
		u, err := url.ParseRequestURI(uri)
		if err != nil || u.Host == "" {
			return []*check.Problem{{
				Severity:   check.Problem_ERROR,
				Message:    fmt.Sprintf(`%s must be an absolute URI.`, field),
				Suggestion: fmt.Sprintf(`Ensure %s includes a host.`, field),
			}}
		}
	}
	return nil
}
