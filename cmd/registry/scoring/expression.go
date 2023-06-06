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

	"github.com/apigee/registry/cmd/registry/patch"
	"github.com/apigee/registry/cmd/registry/scoring/extensions"
	"github.com/apigee/registry/pkg/application/apihub"
	"github.com/apigee/registry/pkg/application/controller"
	"github.com/apigee/registry/pkg/application/scoring"
	"github.com/apigee/registry/pkg/application/style"
	"github.com/apigee/registry/pkg/mime"
	"github.com/google/cel-go/cel"
	metrics "github.com/google/gnostic/metrics"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// https://github.com/google/cel-spec/blob/master/doc/langdef.md#dynamic-values
func evaluateScoreExpression(expression string, artifactMap map[string]interface{}) (interface{}, error) {
	env, err := cel.NewEnv(extensions.Extensions())
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

	out, _, err := prg.Eval(artifactMap)
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
	messageType, err := mime.MessageTypeForMimeType(mimeType)
	if err != nil {
		return nil, fmt.Errorf("failed extracting message type from %q", mimeType)
	}

	switch messageType {
	case "gnostic.metrics.Complexity":
		return unmarshalAndMap(contents, mimeType, &metrics.Complexity{})
	case "gnostic.metrics.Vocabulary":
		return unmarshalAndMap(contents, mimeType, &metrics.Vocabulary{})
	case "google.cloud.apigeeregistry.v1.style.ConformanceReport":
		return unmarshalAndMap(contents, mimeType, &style.ConformanceReport{})
	case "google.cloud.apigeeregistry.v1.style.Lint":
		return unmarshalAndMap(contents, mimeType, &style.Lint{})
	case "google.cloud.apigeeregistry.v1.apihub.ReferenceList":
		return unmarshalAndMap(contents, mimeType, &apihub.ReferenceList{})
	case "google.cloud.apigeeregistry.v1.controller.Receipt":
		return unmarshalAndMap(contents, mimeType, &controller.Receipt{})
	case "google.cloud.apigeeregistry.v1.scoring.Score":
		return unmarshalAndMap(contents, mimeType, &scoring.Score{})
	case "google.cloud.apigeeregistry.v1.scoring.ScoreCard":
		return unmarshalAndMap(contents, mimeType, &scoring.ScoreCard{})
	// TODO: Add support for JSON artifacts
	default:
		return nil, fmt.Errorf("unsupported artifact type: %s", messageType)
	}
}

func unmarshalAndMap(contents []byte, mimeType string, message proto.Message) (map[string]interface{}, error) {
	// Convert to proto
	err := patch.UnmarshalContents(contents, mimeType, message)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshling: %s", err)
	}

	// Convert proto to json
	jsonData, err := protojson.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed converting to json: %s", err)
	}

	var mapValue map[string]interface{}
	// Convert json to map
	err = json.Unmarshal(jsonData, &mapValue)
	if err != nil {
		return nil, fmt.Errorf("failed converting to map: %s", err)
	}

	return mapValue, nil
}
