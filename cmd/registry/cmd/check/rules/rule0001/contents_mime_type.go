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
	"net/http"
	"strings"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/rpc"
)

var contentsMimeType = &lint.FieldRule{
	Name: lint.NewRuleName(1, "contents-mime-type"),
	OnlyIf: func(resource lint.Resource, field string) bool {
		return field == "MimeType"
	},
	ApplyToField: func(resource lint.Resource, field string, value interface{}) []lint.Problem {
		var mimeType string
		var contents []byte
		switch t := resource.(type) {
		case *rpc.ApiSpec:
			mimeType = t.GetMimeType()
			contents = t.GetContents()
		case *rpc.Artifact:
			mimeType = t.GetMimeType()
			contents = t.GetContents()
		}
		mimeType = strings.ToLower(mimeType)
		detected := http.DetectContentType(contents)
		simpleDetected := strings.Split(detected, ";")[0]
		if mimeType != detected && mimeType != simpleDetected {
			return []lint.Problem{{
				Severity:   1,
				Message:    fmt.Sprintf("Unexpected mime type: %q", mimeType),
				Suggestion: fmt.Sprintf("Expected mime type: %q", detected),
			}}
		}
		return nil
	},
}
