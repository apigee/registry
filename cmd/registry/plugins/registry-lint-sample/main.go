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

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/apigee/registry/rpc"
	"google.golang.org/protobuf/proto"
)

// RespondWithErrorAndExit takes in a sequence of errors, sets them in the response,
// responds, and then exits.
func RespondWithErrorAndExit(errs ...error) {
	errorMessages := make([]string, len(errs))
	for i, err := range errs {
		errorMessages[i] = err.Error()
	}
	response := &rpc.LinterResponse{
		Errors: errorMessages,
	}
	RespondAndExit(response)
}

// RespondAndExit serializes and writes the plugin response to STDOUT, and then exits.
func RespondAndExit(response *rpc.LinterResponse) {
	responseBytes, _ := proto.Marshal(response)
	os.Stdout.Write(responseBytes)
	os.Exit(0)
}

// GetRequest constructs a LinterRequest object from standard input.
func GetRequest() (*rpc.LinterRequest, error) {
	// Read from stdin.
	pluginData, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	if len(pluginData) == 0 && err != nil {
		return nil, fmt.Errorf("no input data")
	}

	// Deserialize the request from the input into a request object.
	linterRequest := &rpc.LinterRequest{}
	if err := proto.Unmarshal(pluginData, linterRequest); err != nil {
		return nil, err
	}

	return linterRequest, nil
}

func main() {
	req, err := GetRequest()
	if err != nil {
		RespondWithErrorAndExit(err)
	}

	// Formulate the response. In this sample plugin, we will simply return a fake rule violation /
	// lint problem for every rule that the user specifies, on the given file that is provided.
	lintFile := &rpc.LintFile{
		FilePath: req.SpecPath,
	}
	for _, rule := range req.RuleIds {
		lintFile.Problems = append(lintFile.Problems, &rpc.LintProblem{
			RuleId:  rule,
			Message: fmt.Sprintf("This is a sample violation of the rule %s", rule),
		})
	}
	response := &rpc.LinterResponse{
		Lint: &rpc.Lint{
			Name: "registry-lint-sample",
			Files: []*rpc.LintFile{
				lintFile,
			},
		},
	}

	// Respond by writing response to STDOUT and exiting.
	RespondAndExit(response)
}
