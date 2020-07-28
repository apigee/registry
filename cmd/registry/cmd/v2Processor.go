// Copyright 2020 Google LLC. All Rights Reserved.
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

package cmd

import (
	metrics "github.com/googleapis/gnostic/metrics"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
)

func processOperationV2(operation *openapi_v2.Operation, operationID, parameters map[string]int) {
	if operation.OperationId != "" {
		operationID[operation.OperationId]++
	}
	for _, item := range operation.Parameters {
		switch t := item.Oneof.(type) {
		case *openapi_v2.ParametersItem_Parameter:
			switch t2 := t.Parameter.Oneof.(type) {
			case *openapi_v2.Parameter_BodyParameter:
				parameters[t2.BodyParameter.Name]++
			case *openapi_v2.Parameter_NonBodyParameter:
				nonBodyParam := t2.NonBodyParameter
				processOperationParametersV2(operation, parameters, nonBodyParam)
			}
		}
	}
}

func processOperationParametersV2(operation *openapi_v2.Operation, parameters map[string]int, nonBodyParam *openapi_v2.NonBodyParameter) {
	switch t3 := nonBodyParam.Oneof.(type) {
	case *openapi_v2.NonBodyParameter_FormDataParameterSubSchema:
		parameters[t3.FormDataParameterSubSchema.Name]++
	case *openapi_v2.NonBodyParameter_HeaderParameterSubSchema:
		parameters[t3.HeaderParameterSubSchema.Name]++
	case *openapi_v2.NonBodyParameter_PathParameterSubSchema:
		parameters[t3.PathParameterSubSchema.Name]++
	case *openapi_v2.NonBodyParameter_QueryParameterSubSchema:
		parameters[t3.QueryParameterSubSchema.Name]++
	}
}

func processSchemaV2(schema *openapi_v2.Schema, properties map[string]int) {
	if schema.Properties == nil {
		return
	}
	for _, pair := range schema.Properties.AdditionalProperties {
		properties[pair.Name]++
	}
}

func processDocumentV2(document *openapi_v2.Document) *metrics.Vocabulary {
	schemas := make(map[string]int)
	operationID := make(map[string]int)
	parameters := make(map[string]int)
	properties := make(map[string]int)

	if document.Definitions != nil {
		for _, pair := range document.Definitions.AdditionalProperties {
			schemas[pair.Name]++
			processSchemaV2(pair.Value, properties)
		}
	}
	for _, pair := range document.Paths.Path {
		v := pair.Value
		if v.Get != nil {
			processOperationV2(v.Get, operationID, parameters)
		}
		if v.Post != nil {
			processOperationV2(v.Post, operationID, parameters)
		}
		if v.Put != nil {
			processOperationV2(v.Put, operationID, parameters)
		}
		if v.Patch != nil {
			processOperationV2(v.Patch, operationID, parameters)
		}
		if v.Delete != nil {
			processOperationV2(v.Delete, operationID, parameters)
		}
	}

	vocab := &metrics.Vocabulary{
		Schemas:    fillProtoStructures(schemas),
		Operations: fillProtoStructures(operationID),
		Parameters: fillProtoStructures(parameters),
		Properties: fillProtoStructures(properties),
	}

	return vocab
}
