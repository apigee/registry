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

package linter

import (
	"fmt"
	"io"
	"os"

	"github.com/apigee/registry/pkg/application/style"
	"google.golang.org/protobuf/proto"
)

// RespondWithErrorAndExit takes in a sequence of errors, sets them in the response,
// responds, and then exits.
func respondWithErrorAndExit(errs ...error) {
	errorMessages := make([]string, len(errs))
	for i, err := range errs {
		errorMessages[i] = err.Error()
	}
	response := &style.LinterResponse{
		Errors: errorMessages,
	}
	respondAndExit(response)
}

// RespondAndExit serializes and writes the plugin response to STDOUT, and then exits.
func respondAndExit(response *style.LinterResponse) {
	responseBytes, _ := proto.Marshal(response)
	os.Stdout.Write(responseBytes)
	os.Exit(0)
}

// GetRequest constructs a LinterRequest object from standard input.
func getRequest() (*style.LinterRequest, error) {
	// Read from stdin.
	pluginData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	if len(pluginData) == 0 && err != nil {
		return nil, fmt.Errorf("no input data")
	}

	// Deserialize the request from the input into a request object.
	linterRequest := &style.LinterRequest{}
	if err := proto.Unmarshal(pluginData, linterRequest); err != nil {
		return nil, err
	}

	return linterRequest, nil
}

// Main reads the request from STDIN, runs the linter plugin, and
// writes the response to STDOUT.
func Main(runner LinterRunner) {
	req, err := getRequest()
	if err != nil {
		respondWithErrorAndExit(err)
	}

	resp, err := runner.Run(req)
	if err != nil {
		respondWithErrorAndExit(err)
	}

	respondAndExit(resp)
}
