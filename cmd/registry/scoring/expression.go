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

package scoring

import (
	"fmt"
	"reflect"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/rpc"
	"github.com/google/cel-go/cel"
	metrics "github.com/google/gnostic/metrics"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

var nativeTypeMap = map[string]reflect.Type{
	"bool":   reflect.TypeOf(false),
	"int":    reflect.TypeOf(0),
	"double": reflect.TypeOf(0.5),
}

func evaluateScoreExpression(expression string, mimeType string, contents []byte) (interface{}, error) {

	// https://github.com/google/cel-spec/blob/master/doc/langdef.md#dynamic-values
	structProto, err := getStructProto(contents, mimeType)
	if err != nil {
		return nil, err
	}

	// CEL expression
	env, err := cel.NewEnv()
	if err != nil {
		return nil, fmt.Errorf("error creating CEL environment: %s", err)
	}

	ast, issues := env.Parse(expression)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("error parsing score_expression, %q: %s", expression, issues)
	}
	prg, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("program construction error, %q: %s", expression, err)
	}

	out, _, err := prg.Eval(structProto.AsMap())
	if err != nil {
		return nil, fmt.Errorf("error in evaluating expression %q: %s", expression, err)
	}

	switch out.Type().TypeName() {
	case "int":
		expectedType := nativeTypeMap["int"]
		nativeVal, err := out.ConvertToNative(expectedType)
		if err != nil {
			return nil, fmt.Errorf("failed converting output from type %s to %s: %s", out.Type().TypeName(), expectedType, err)
		}
		return nativeVal, nil
	case "double":
		expectedType := nativeTypeMap["double"]
		nativeVal, err := out.ConvertToNative(expectedType)
		if err != nil {
			return nil, fmt.Errorf("failed converting output from type %s to %s: %s", out.Type().TypeName(), expectedType, err)
		}
		return nativeVal, nil
	case "bool":
		expectedType := nativeTypeMap["bool"]
		nativeVal, err := out.ConvertToNative(expectedType)
		if err != nil {
			return nil, fmt.Errorf("failed converting output from type %s to %s: %s", out.Type().TypeName(), expectedType, err)
		}
		return nativeVal, nil
	default:
		return nil, fmt.Errorf("evaluating expression %q generated an unexpected output type %s: should be one of [int, float, bool]", expression, out.Type().TypeName())

	}
}

func getStructProto(contents []byte, mimeType string) (*structpb.Struct, error) {
	messageType, err := core.MessageTypeForMimeType(mimeType)
	if err != nil {
		return nil, fmt.Errorf("failed extracting message type from %q", mimeType)
	}

	switch messageType {
	case "gnostic.metrics.Complexity":
		return unmarshalAndStruct(contents, &metrics.Complexity{})
	case "gnostic.metrics.Vocabulary":
		return unmarshalAndStruct(contents, &metrics.Vocabulary{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.ConformanceReport":
		return unmarshalAndStruct(contents, &rpc.ConformanceReport{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.Index":
		return unmarshalAndStruct(contents, &rpc.Index{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.Lint":
		return unmarshalAndStruct(contents, &rpc.Lint{})
	case "google.cloud.apigeeregistry.v1.apihub.ReferenceList":
		return unmarshalAndStruct(contents, &rpc.ReferenceList{})
	case "google.cloud.apigeeregistry.v1.controller.Receipt":
		return unmarshalAndStruct(contents, &rpc.Receipt{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.References":
		return unmarshalAndStruct(contents, &rpc.References{})
	case "google.cloud.apigeeregistry.v1.scoring.Score":
		return unmarshalAndStruct(contents, &rpc.Score{})
	case "google.cloud.apigeeregistry.v1.scoring.ScoreCard":
		return unmarshalAndStruct(contents, &rpc.ScoreCard{})
	// TODO: Add support for JSON artifacts
	default:
		return nil, fmt.Errorf("unsupported artifact type: %s", messageType)
	}
}

func unmarshalAndStruct(contents []byte, message proto.Message) (*structpb.Struct, error) {
	structProto := &structpb.Struct{}

	err := proto.Unmarshal(contents, message)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshling: %s", err)
	}

	// Convert to Struct proto
	jsonData, err := protojson.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed converting to json: %s", err)
	}

	err = protojson.Unmarshal(jsonData, structProto)
	if err != nil {
		return nil, fmt.Errorf("failed converting to structpb.Struct{}: %s", err)
	}

	return structProto, nil
}
