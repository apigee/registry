// Copyright 2021 Google LLC.
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

package lint

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"google.golang.org/protobuf/proto"

	"github.com/apigee/registry/pkg/application/style"
	linter "github.com/google/gnostic/metrics/lint"
	yaml "gopkg.in/yaml.v3"
)

func lintFileForOpenAPIWithGnostic(path string, root string) (*style.LintFile, error) {
	cmd := exec.Command("gnostic", path, "--linter-out=.")
	cmd.Dir = root
	_, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join(root, "/linter.pb"))
	if err != nil {
		return nil, err
	}
	output := &linter.Linter{}
	if err := proto.Unmarshal(b, output); err != nil {
		return nil, err
	}
	nodeFinder, err := newNodeFinder(filepath.Join(root, path))
	if err != nil {
		return nil, err
	}
	problems := make([]*style.LintProblem, 0)
	for _, message := range output.Messages {
		problem := &style.LintProblem{
			Message:    message.Message,
			Suggestion: message.Suggestion,
		}
		node, err := nodeFinder.findNode(message.Keys)
		if err == nil {
			l := int32(node.Line)
			c := int32(node.Column)
			problem.Location = &style.LintLocation{
				StartPosition: &style.LintPosition{
					LineNumber:   l,
					ColumnNumber: c,
				},
				EndPosition: &style.LintPosition{
					LineNumber:   l,
					ColumnNumber: c + int32(len(node.Value)-1),
				},
			}
		}
		problems = append(problems, problem)
	}
	result := &style.LintFile{Problems: problems}
	return result, nil
}

type nodeFinder struct {
	node *yaml.Node
}

func newNodeFinder(filename string) (*nodeFinder, error) {
	data, _ := os.ReadFile(filename)
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, err
	}
	return &nodeFinder{node: &node}, nil
}

// FindNode returns a node object pointing to the given token in a yaml file. The node contains
// information such as the string value, line number, bordering comments, etc.
func (nf *nodeFinder) findNode(keys []string) (*yaml.Node, error) {
	return findNode(nf.node.Content[0], 0, len(keys)-1, keys)
}

// findNode recursively iterates through the yaml file using the node feature. The function
// will continue until the token is found at the max depth. If the token is not found, an
// empty node is returned.
func findNode(node *yaml.Node, keyIndex, maxDepth int, keys []string) (*yaml.Node, error) {
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
		}
	}
	return &yaml.Node{}, nil
}
