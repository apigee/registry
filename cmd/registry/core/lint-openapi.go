// Copyright 2021 Google LLC. All Rights Reserved.
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

package core

import (
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/apigee/registry/rpc"
	linter "github.com/googleapis/gnostic/metrics/lint"
	sourceinfo "github.com/googleapis/gnostic/metrics/sourceinfo"
	"google.golang.org/protobuf/proto"
)

// NewLintFromOpenAPIv2 runs the API linter and returns the results.
func NewLintFromOpenAPIv2(name string, b []byte) (*rpc.Lint, error) {
	// create a tmp directory
	root, err := ioutil.TempDir("", "registry-openapi-")
	if err != nil {
		return nil, err
	}
	log.Printf("running in %s", root)
	name = filepath.Base(name)
	// whenever we finish, delete the tmp directory
	//	defer os.RemoveAll(root)
	// write the file to the temp directory
	spec, err := GUnzippedBytes(b)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(root+"/"+name, spec, 0644)
	if err != nil {
		return nil, err
	}
	// run the linter on the spec
	lint := &rpc.Lint{}
	lint.Name = name
	// run the linter on the spec
	lintFile, err := lintFileForOpenAPIv2(name, root)
	if err != nil {
		return nil, err
	}
	lint.Files = append(lint.Files, lintFile)
	return lint, err
}

func lintFileForOpenAPIv2(path string, root string) (*rpc.LintFile, error) {
	cmd := exec.Command("gnostic", path, "--linter-out=.")
	cmd.Dir = root
	_, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(root + "/linter.pb")
	if err != nil {
		return nil, err
	}
	output := &linter.Linter{}
	if err := proto.Unmarshal(b, output); err != nil {
		return nil, err
	}
	problems := make([]*rpc.LintProblem, 0)
	for _, message := range output.Messages {
		problem := &rpc.LintProblem{}
		problem.Message = message.Message
		problem.Suggestion = message.Suggestion
		problems = append(problems, problem)
		keys := message.Keys[0 : len(message.Keys)-1]
		token := message.Keys[len(message.Keys)-1]
		log.Printf("%+v", keys)
		log.Printf("%+v", token)
		node, err := sourceinfo.FindNode(root+"/"+path, keys, token)
		log.Printf("%+v %+v", err, node)

	}
	result := &rpc.LintFile{}
	result.Problems = problems
	return result, err
}

func NewLintFromOpenAPIv3(name string, b []byte) (*rpc.Lint, error) {
	lint := &rpc.Lint{}
	return lint, nil
}
