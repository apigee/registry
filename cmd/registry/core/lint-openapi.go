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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/apigee/registry/rpc"
	linter "github.com/googleapis/gnostic/metrics/lint"
	"google.golang.org/protobuf/proto"
	yaml "gopkg.in/yaml.v3"
)

// NewLintFromOpenAPI runs the API linter and returns the results.
func NewLintFromOpenAPI(name string, b []byte) (*rpc.Lint, error) {
	// create a tmp directory
	root, err := ioutil.TempDir("", "registry-openapi-")
	if err != nil {
		return nil, err
	}
	// log.Printf("running in %s", root)
	name = filepath.Base(name)
	// whenever we finish, delete the tmp directory
	defer os.RemoveAll(root)
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
	lintFile, err := lintFileForOpenAPI(name, root)
	if err != nil {
		return nil, err
	}
	lint.Files = append(lint.Files, lintFile)
	return lint, err
}

func lintFileForOpenAPI(path string, root string) (*rpc.LintFile, error) {
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
		keys := message.Keys
		node, err := FindNode(root+"/"+path, keys)
		if err == nil {
			l := int32(node.Line)
			c := int32(node.Column)
			problem.Location = &rpc.LintLocation{}
			problem.Location.StartPosition = &rpc.LintPosition{
				LineNumber:   l,
				ColumnNumber: c,
			}
			problem.Location.EndPosition = &rpc.LintPosition{
				LineNumber:   l,
				ColumnNumber: c + int32(len(node.Value)-1),
			}
		}
		problems = append(problems, problem)
	}
	result := &rpc.LintFile{}
	result.Problems = problems
	return result, err
}

// findNode recursively iterates through the yaml file using the node feature. The function
// will continue until the token is found at the max depth. If the token is not found, an
// empty node is returned.
func findNode(node *yaml.Node, keyIndex int, maxDepth int, keys []string) (*yaml.Node, error) {
	if keyIndex > maxDepth {
		return node, nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		if keyIndex+1 == maxDepth && val.Value == keys[maxDepth] {
			return val, nil
		}
		if key.Value == keys[keyIndex] {
			switch val.Kind {
			case yaml.SequenceNode:
				nextKeyIndex, err := strconv.Atoi(keys[keyIndex+1])
				if err != nil {
					return nil, err
				}
				return findNode(val.Content[nextKeyIndex], keyIndex+2, maxDepth, keys)
			default:
				return findNode(val, keyIndex+1, maxDepth, keys)
			}
		} else {
			continue
		}

	}
	return &yaml.Node{}, nil
}

// FindNode returns a node object pointing to the given token in a yaml file. The node contains
// information such as the string value, line number, bordering commments, etc.
func FindNode(filename string, keys []string) (*yaml.Node, error) {
	data, _ := ioutil.ReadFile(filename)
	var node yaml.Node
	err := yaml.Unmarshal(data, &node)
	if err != nil {
		fmt.Printf("%+v", err)
	}
	return findNode(node.Content[0], 0, len(keys)-1, keys)
}
