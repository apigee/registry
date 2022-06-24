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

package diff

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/apigee/registry/rpc"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/diff"
)

// change represents one change in the diff.
type change struct {
	fieldPath  stack
	changeType string
}

type stack []string

func (s stack) String() string {
	return strings.Join(s, ".")
}

func (s stack) isEmpty() bool {
	return len(s) == 0
}

func (s *stack) push(str string) {
	*s = append(*s, str)
}

func (s *stack) pop() {
	if s.isEmpty() {
		return
	}
	index := len(*s) - 1
	*s = (*s)[:index]
}

// GetDiff takes two yaml or json Open API 3 Specs and diffs them.
func GetDiff(base, revision []byte) (*rpc.Diff, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	baseSpec, err := loader.LoadFromData(base)
	if err != nil {
		err := fmt.Errorf("failed to load base spec from %q with %v", base, err)
		return nil, err
	}
	revisionSpec, err := loader.LoadFromData(revision)
	if err != nil {
		err := fmt.Errorf("failed to load revision spec from %q with %v", revision, err)
		return nil, err
	}
	diffReport, err := diff.Get(&diff.Config{
		ExcludeExamples:    false,
		ExcludeDescription: false,
	}, baseSpec, revisionSpec)
	if err != nil {
		err := fmt.Errorf("diff failed with %v", err)
		return nil, err
	}
	return getChanges(diffReport)
}

func addToDiffProto(diffProto *rpc.Diff, changePath *change) {
	fieldName := changePath.fieldPath.String()
	switch changePath.changeType {
	case "added":
		diffProto.Additions = append(diffProto.Additions, fieldName)
	case "deleted":
		diffProto.Deletions = append(diffProto.Deletions, fieldName)
	}
}

// getChanges creates a protodif report from a diff.Diff struct.
func getChanges(diff *diff.Diff) (*rpc.Diff, error) {
	diffProto := &rpc.Diff{
		Additions:     []string{},
		Deletions:     []string{},
		Modifications: make(map[string]*rpc.Diff_ValueChange),
	}
	change := &change{
		fieldPath:  stack{},
		changeType: "",
	}
	diffNode := reflect.ValueOf(diff)
	err := searchNode(diffNode, diffProto, change)
	return diffProto, err
}

func searchNode(value reflect.Value, diffProto *rpc.Diff, changePath *change) error {
	// Some values might be pointers, so we should dereference before continuing.
	value = reflect.Indirect(value)

	// Invalid values aren't relevant to the diff.
	if !value.IsValid() {
		return nil
	}

	switch value.Kind() {
	case reflect.Map:
		return searchMapType(value, diffProto, changePath)
	case reflect.Array, reflect.Slice:
		return searchArrayAndSliceType(value, diffProto, changePath)
	case reflect.Struct:
		return searchStructType(value, diffProto, changePath)
	case reflect.Float64, reflect.String, reflect.Bool, reflect.Int:
		valueString := scalarToString(value)
		changePath.fieldPath.push(valueString)
		addToDiffProto(diffProto, changePath)
		changePath.fieldPath.pop()
		return nil
	default:
		return fmt.Errorf("field %q has unknown type %s with value %v", changePath.fieldPath, value.Type(), value)
	}
}

func searchMapType(mapNode reflect.Value, diffProto *rpc.Diff, changePath *change) error {
	if mapNode.Kind() != reflect.Map {
		panic("searchMapType called with invalid type")
	}

	for _, childNodeKey := range mapNode.MapKeys() {
		childNode := mapNode.MapIndex(childNodeKey)
		if childNode.IsZero() {
			continue
		}
		switch childNodeKey.Interface().(type) {
		case float64, string, bool, int:
			childNodeKeyName := scalarToString(childNodeKey)
			changePath.fieldPath.push(childNodeKeyName)
			err := searchNode(childNode, diffProto, changePath)
			if err != nil {
				return err
			}
			changePath.fieldPath.pop()
			continue

		case diff.Endpoint:
			if endpoint, ok := childNodeKey.Interface().(diff.Endpoint); ok {
				changePath.fieldPath.push(fmt.Sprintf("{%s %s}", endpoint.Method, endpoint.Path))
				err := searchNode(childNode, diffProto, changePath)
				if err != nil {
					return err
				}
				changePath.fieldPath.pop()
				continue
			}
			return fmt.Errorf("searchMapType called with invalid diff.Endpoint type: %v", childNodeKey)
		default:
			return fmt.Errorf("map node key %v is not supported", childNodeKey)
		}
	}
	return nil
}

func searchArrayAndSliceType(arrayNode reflect.Value, diffProto *rpc.Diff, changePath *change) error {
	if arrayNode.Kind() != reflect.Slice && arrayNode.Kind() != reflect.Array {
		panic("searchArrayAndSliceType called with invalid type")
	}

	for i := 0; i < (arrayNode.Len()); i++ {
		childNode := arrayNode.Index(i)
		if childNode.IsZero() {
			continue
		}
		switch childNode.Interface().(type) {
		case string:
			changePath.fieldPath.push(childNode.String())
			addToDiffProto(diffProto, changePath)
			changePath.fieldPath.pop()
			continue
		case diff.Endpoint:
			if endpoint, ok := childNode.Interface().(diff.Endpoint); ok {
				changePath.fieldPath.push(fmt.Sprintf("{%s %s}", endpoint.Method, endpoint.Path))
				addToDiffProto(diffProto, changePath)
				changePath.fieldPath.pop()
				continue
			}
			return fmt.Errorf("searchArrayAndSliceType called with invalid diff.Endpoint type: %v", childNode)
		default:
			return fmt.Errorf("array child node %v is not supported", childNode)
		}
	}
	return nil
}

func searchStructType(structNode reflect.Value, diffProto *rpc.Diff, changePath *change) error {
	if structNode.Kind() != reflect.Struct {
		panic("searchStructType called with invalid type")
	}

	if vd, ok := structNode.Interface().(diff.ValueDiff); ok {
		diffProto.Modifications[changePath.fieldPath.String()] = &rpc.Diff_ValueChange{
			From: scalarToString(reflect.ValueOf(vd.From)),
			To:   scalarToString(reflect.ValueOf(vd.To)),
		}
		return nil
	}
	for i := 0; i < structNode.NumField(); i++ {
		tag, ok := structNode.Type().Field(i).Tag.Lookup("json")
		if !ok {
			// Fields that don't have a JSON name aren't part of the diff.
			continue
		}

		// Empty fields in the diff are redundant. Skip them.
		childNode := structNode.Field(i)
		if childNode.IsZero() {
			continue
		}

		// Struct field tags have the format "fieldname,omitempty". We only want the field name.
		fieldName := strings.Split(tag, ",")[0]

		err := handleStructField(childNode, fieldName, diffProto, changePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleStructField(value reflect.Value, name string, diffProto *rpc.Diff, changePath *change) error {
	// Empty fields in the diff are redundant. Skip them.
	if value.IsZero() {
		return nil
	}

	switch name {
	case "added", "deleted", "modified":
		changePath.changeType = name
	default:
		changePath.fieldPath.push(name)
		defer changePath.fieldPath.pop()
	}

	err := searchNode(value, diffProto, changePath)
	if err != nil {
		return err
	}

	return nil
}

func scalarToString(node reflect.Value) string {
	switch node.Kind() {
	case reflect.Float64:
		return fmt.Sprintf("%f", node.Float())
	case reflect.String:
		return node.String()
	case reflect.Bool:
		return fmt.Sprintf("%t", node.Bool())
	case reflect.Int:
		return fmt.Sprintf("%d", node.Int())
	default:
		return ""
	}
}
