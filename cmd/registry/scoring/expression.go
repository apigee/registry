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
	"encoding/json"
	"fmt"

	"github.com/apigee/registry/cmd/registry/core"
	"github.com/apigee/registry/rpc"
	"github.com/google/cel-go/cel"
	metrics "github.com/google/gnostic/metrics"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func evaluateScoreExpression(expression, mimeType string, contents []byte) (interface{}, error) {

	// https://github.com/google/cel-spec/blob/master/doc/langdef.md#dynamic-values
	structProto, err := getMap(contents, mimeType)
	if err != nil {
		return nil, err
	}

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

	out, _, err := prg.Eval(structProto)
	if err != nil {
		return nil, fmt.Errorf("error in evaluating expression %q: %s", expression, err)
	}

	switch value := out.Value().(type) {
	case int64, float64, bool:
		return value, nil
	default:
		return nil, fmt.Errorf("evaluating expression %q generated an unexpected output type %T: should be one of [int, double, bool]", expression, value)
	}
}

func getMap(contents []byte, mimeType string) (map[string]interface{}, error) {
	messageType, err := core.MessageTypeForMimeType(mimeType)
	if err != nil {
		return nil, fmt.Errorf("failed extracting message type from %q", mimeType)
	}

	switch messageType {
	case "gnostic.metrics.Complexity":
		return unmarshalAndMap(contents, &metrics.Complexity{})
	case "gnostic.metrics.Vocabulary":
		return unmarshalAndMap(contents, &metrics.Vocabulary{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.ConformanceReport":
		return unmarshalAndMap(contents, &rpc.ConformanceReport{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.Index":
		return unmarshalAndMap(contents, &rpc.Index{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.Lint":
		return unmarshalAndMap(contents, &rpc.Lint{})
	case "google.cloud.apigeeregistry.v1.apihub.ReferenceList":
		return unmarshalAndMap(contents, &rpc.ReferenceList{})
	case "google.cloud.apigeeregistry.v1.controller.Receipt":
		return unmarshalAndMap(contents, &rpc.Receipt{})
	case "google.cloud.apigeeregistry.applications.v1alpha1.References":
		return unmarshalAndMap(contents, &rpc.References{})
	case "google.cloud.apigeeregistry.v1.scoring.Score":
		return unmarshalAndMap(contents, &rpc.Score{})
	case "google.cloud.apigeeregistry.v1.scoring.ScoreCard":
		return unmarshalAndMap(contents, &rpc.ScoreCard{})
	// TODO: Add support for JSON artifacts
	default:
		return nil, fmt.Errorf("unsupported artifact type: %s", messageType)
	}
}

func unmarshalAndMap(contents []byte, message proto.Message) (map[string]interface{}, error) {
	mapValue := make(map[string]interface{}, 0)

	// Convert to proto
	err := proto.Unmarshal(contents, message)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshling: %s", err)
	}

	// Convert proto to json
	jsonData, err := protojson.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed converting to json: %s", err)
	}

	// Convert json to map
	err = json.Unmarshal(jsonData, &mapValue)
	if err != nil {
		return nil, fmt.Errorf("failed converting to map: %s", err)
	}

	return mapValue, nil
}
