// Copyright 2022 Google LLC
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

	"github.com/apigee/registry/pkg/application/style"
)

type ruleMetadata struct {
	guidelineRule *style.Rule      // Rule object associated with the linter-rule.
	guideline     *style.Guideline // Guideline object associated with the linter-rule.
}

type linterMetadata struct {
	name          string
	rules         []string
	rulesMetadata map[string]*ruleMetadata
}

func getLinterBinaryName(linterName string) string {
	return "registry-lint-" + linterName
}

func GenerateLinterMetadata(styleguide *style.StyleGuide) (map[string]*linterMetadata, error) {
	linterNameToMetadata := make(map[string]*linterMetadata)

	// Iterate through all the guidelines of the style guide.
	for _, guideline := range styleguide.GetGuidelines() {
		// Iterate through all the rules of the style guide.
		for _, rule := range guideline.GetRules() {
			// Get the name of the linter associated with the rule.
			linterName := rule.GetLinter()
			if len(linterName) == 0 {
				continue
			}

			metadata, ok := linterNameToMetadata[linterName]
			if !ok {
				metadata = &linterMetadata{
					name:          linterName,
					rules:         make([]string, 0),
					rulesMetadata: make(map[string]*ruleMetadata),
				}
				linterNameToMetadata[linterName] = metadata
			}

			linterRuleName := rule.GetLinterRulename()
			if len(linterRuleName) == 0 {
				continue
			}

			//Populate required metadata
			metadata.rules = append(metadata.rules, linterRuleName)

			if _, ok := metadata.rulesMetadata[linterRuleName]; !ok {
				metadata.rulesMetadata[linterRuleName] = &ruleMetadata{}
			}
			metadata.rulesMetadata[linterRuleName].guideline = guideline
			metadata.rulesMetadata[linterRuleName].guidelineRule = rule
		}
	}

	if len(linterNameToMetadata) == 0 {
		return nil, fmt.Errorf("empty linter metadata")
	}
	return linterNameToMetadata, nil
}
