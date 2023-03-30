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
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/apigee/registry/cmd/registry/compress"
	"github.com/apigee/registry/pkg/application/style"
	"github.com/apigee/registry/pkg/connection"
	"github.com/apigee/registry/pkg/log"
	"github.com/apigee/registry/pkg/mime"
	"github.com/apigee/registry/pkg/visitor"
	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/proto"
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

func SimpleLinterMetadata(linter string) *linterMetadata {
	return &linterMetadata{name: linter}
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

func WriteSpecForLinting(ctx context.Context, client connection.RegistryClient, spec *rpc.ApiSpec) (string, error) {
	err := visitor.FetchSpecContents(ctx, client, spec)
	if err != nil {
		return "", err
	}
	// Put the spec in a temporary directory.
	root, err := os.MkdirTemp("", "registry-spec-")
	if err != nil {
		return "", err
	}
	filename := spec.GetFilename()
	if filename == "" {
		return root, fmt.Errorf("%s does not specify a filename", spec.GetName())
	}
	name := filepath.Base(filename)

	if mime.IsZipArchive(spec.GetMimeType()) {
		_, err = compress.UnzipArchiveToPath(spec.GetContents(), root)
	} else {
		// Write the file to the temporary directory.
		err = os.WriteFile(filepath.Join(root, name), spec.GetContents(), 0644)
	}
	if err != nil {
		return root, err
	}
	return root, nil
}

func RunLinter(ctx context.Context,
	specDirectory string,
	metadata *linterMetadata) (*style.LinterResponse, error) {
	// Formulate the request.
	requestBytes, err := proto.Marshal(&style.LinterRequest{
		SpecDirectory: specDirectory,
		RuleIds:       metadata.rules,
	})
	if err != nil {
		return nil, fmt.Errorf("failed marshaling linterRequest, Error: %s ", err)
	}

	executableName := getLinterBinaryName(metadata.name)
	cmd := exec.Command(executableName)
	cmd.Stdin = bytes.NewReader(requestBytes)
	cmd.Stderr = os.Stderr

	pluginStartTime := time.Now()
	// Run the linter.
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running the plugin %s return error: %s", executableName, err)
	}

	pluginElapsedTime := time.Since(pluginStartTime)
	log.Debugf(ctx, "Plugin %s ran in time %s", executableName, pluginElapsedTime)

	// Unmarshal the output bytes into a response object. If there's a failure, log and continue.
	linterResponse := &style.LinterResponse{}
	err = proto.Unmarshal(output, linterResponse)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshalling LinterResponse (plugins must write log messages to stderr, not stdout): %s", err)
	}

	// Check if there were any errors in the plugin.
	if len(linterResponse.GetErrors()) > 0 {
		return nil, fmt.Errorf("plugin %s encountered errors: %v", executableName, linterResponse.GetErrors())
	}

	return linterResponse, nil
}
