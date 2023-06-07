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

package rule111

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/apigee/registry/cmd/registry/cmd/check/lint"
	"github.com/apigee/registry/pkg/application/check"
	"github.com/apigee/registry/rpc"
)

var ruleNum = 111
var ruleName = lint.NewRuleName(ruleNum, "mime-type-detected-contents")

// AddRules accepts a register function and registers each of
// this rules' checks to the RuleRegistry.
func AddRules(r lint.RuleRegistry) error {
	return r.Register(
		ruleNum,
		mimeTypeContents,
	)
}

var allowPrefixes = []string{
	"application/octet-stream",
	"application/x.",
	"application/x-",
	"application/vnd",
}

var mimeTypeContents = &lint.FieldRule{
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
		for _, p := range allowPrefixes {
			if strings.HasPrefix(declared, p) {
				return checkGzip(declared, contents)
			}
		}

		detected := http.DetectContentType(contents)

		if strings.TrimSpace(declared) == "" {
			return []*check.Problem{{
				Severity:   check.Problem_ERROR,
				Message:    "Empty mime_type.",
				Suggestion: fmt.Sprintf("Detected mime_type: %q.", detected),
			}}
		}

		declaredType, _, err := mime.ParseMediaType(declared)
		if err != nil {
			return []*check.Problem{{
				Severity: check.Problem_ERROR,
				Message:  fmt.Sprintf("Unable to parse mime_type %q: %s.", declared, err),
			}}
		}

		detectedType, _, _ := mime.ParseMediaType(detected)
		if declaredType != detectedType {
			return []*check.Problem{{
				Severity:   check.Problem_WARNING,
				Message:    fmt.Sprintf("Unexpected mime_type %q for contents.", declared),
				Suggestion: fmt.Sprintf("Detected mime_type: %q.", detected),
			}}
		}

		return checkGzip(declared, contents)
	},
}

func checkGzip(declared string, contents []byte) []*check.Problem {
	if strings.Contains(declared, "+gzip") {
		buf := bytes.NewBuffer(contents)
		_, err := gzip.NewReader(buf)
		if err != nil {
			return []*check.Problem{{
				Severity: check.Problem_ERROR,
				Message:  fmt.Sprintf("Mime %q indicates gzip, but contents does not have a valid gzip header", declared),
			}}
		}
	}
	return nil
}
