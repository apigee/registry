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
	"github.com/apigee/registry/cmd/registry/types"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/encoding/protojson"
)

var contentsMimeType = &lint.FieldRule{
	Name: lint.NewRuleName(1, "contents-mime-type"),
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
		declaredType, declaredParameters := parseMime(declared)

		// internal types, see github.com/apigee/registry/cmd/registry/types
		typeParam := declaredParameters["type"]
		if strings.HasPrefix(typeParam, "google.cloud.apigeeregistry.") ||
			strings.HasPrefix(typeParam, "gnostic.metrics.") {
			m, err := types.MessageForMimeType(declared)
			if err != nil {
				return []lint.Problem{{
					Severity:   lint.ERROR,
					Message:    fmt.Sprintf("Unknown internal mime_type: %q.", declared),
					Suggestion: "Fix mime_type.",
				}}
			}

			err = protojson.Unmarshal(contents, m)
			if err != nil {
				return []lint.Problem{{
					Severity:   lint.ERROR,
					Message:    fmt.Sprintf("Error loading contents into proto type %q: %v.", declared, err),
					Suggestion: "Fix mime_type or contents.",
				}}
			}

			return nil
		}

		detected := http.DetectContentType(contents)
		detectedType, _ := parseMime(detected)

		if declaredType != detectedType {
			return []lint.Problem{{
				Severity:   lint.WARNING,
				Message:    fmt.Sprintf("Unexpected mime_type %q for contents.", declared),
				Suggestion: fmt.Sprintf("Detected mime_type %q for contents.", detected),
			}}
		}
		return nil
	},
}

func parseMime(mime string) (baseType string, parameters map[string]string) {
	mime = strings.ToLower(mime)
	splits := strings.Split(mime, ";")
	baseType = strings.TrimSpace(splits[0])
	parameters = make(map[string]string)
	if len(splits) > 1 {
		for _, s := range splits[1:] {
			p := strings.Split(s, "=")
			parameters[p[0]] = p[1]
		}
	}
	return
}
