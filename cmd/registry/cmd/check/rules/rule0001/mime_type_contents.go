// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rule0001

import (
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/rpc"
)

var mimeTypeContents = &lint.FieldRule{
	Name: lint.NewRuleName(1, "mime-type-contents"),
	OnlyIf: func(resource lint.Resource, field string) bool {
		return field == "MimeType"
	},
	ApplyToField: func(resource lint.Resource, field string, value interface{}) []lint.Problem {
		var declared string
		var contents []byte
		switch t := resource.(type) {
		case *rpc.ApiSpec:
			declared = t.GetMimeType()
			contents = t.GetContents()
		case *rpc.Artifact:
			declared = t.GetMimeType()
			contents = t.GetContents()
		}
		if len(contents) == 0 {
			return nil
		}
		detected := http.DetectContentType(contents)

		if strings.TrimSpace(declared) == "" {
			return []lint.Problem{{
				Severity:   lint.ERROR,
				Message:    "Empty mime_type.",
				Suggestion: fmt.Sprintf("Detected mime_type: %q.", detected),
			}}
		}

		declaredType, _, err := mime.ParseMediaType(declared)
		if err != nil {
			return []lint.Problem{{
				Severity: lint.ERROR,
				Message:  fmt.Sprintf("Unable to parse mime_type %q: %s.", declared, err),
			}}
		}

		detectedType, _, _ := mime.ParseMediaType(detected)
		if declaredType != detectedType {
			return []lint.Problem{{
				Severity:   lint.WARNING,
				Message:    fmt.Sprintf("Unexpected mime_type %q for contents.", declared),
				Suggestion: fmt.Sprintf("Detected mime_type: %q.", detected),
			}}
		}
		return nil
	},
}
