// Copyright 2023 Google LLC
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

package rule1002

import (
	"context"
	"fmt"
	"strings"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/rpc"
)

var ruleNum = 1002
var ruleName = lint.NewRuleName(ruleNum, "internal-mime-type-contents")

// AddRules accepts a register function and registers each of
// this rules' checks to the RuleRegistry.
func AddRules(r lint.RuleRegistry) error {
	return r.Register(
		ruleNum,
		internalMimeTypeContents,
	)
}

var internalMimeTypeContents = &lint.FieldRule{
	Name: ruleName,
	OnlyIf: func(resource lint.Resource, field string) bool {
		return field == "MimeType"
	},
	ApplyToField: func(ctx context.Context, resource lint.Resource, field string, value interface{}) []*check.Problem {
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
		_, declaredParameters := parseMime(declared)

		// internal types, see github.com/apigee/registry/cmd/registry/types
		typeParam := declaredParameters["type"]
		if strings.HasPrefix(typeParam, "google.cloud.apigeeregistry.") || strings.HasPrefix(typeParam, "gnostic.metrics.") {
			message, err := mime.MessageForMimeType(declared)
			if err != nil {
				return []*check.Problem{{
					Severity:   check.Problem_ERROR,
					Message:    fmt.Sprintf("Unknown internal mime_type: %q.", declared),
					Suggestion: "Fix mime_type.",
				}}
			}

			// does not validate contents, just proves type compatibility
			err = patch.UnmarshalContents(contents, declared, message)
			if err != nil {
				return []*check.Problem{{
					Severity:   check.Problem_ERROR,
					Message:    fmt.Sprintf("Error loading contents into proto type %q: %v.", declared, err),
					Suggestion: "Fix mime_type or contents.",
				}}
			}
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
