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

package extensions

import (
	"fmt"
	"reflect"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
)

// holds the supported types
var nativeTypeMap = map[string]reflect.Type{
	"list<int>":    reflect.TypeOf([]int64{}),
	"int":          reflect.TypeOf(int64(0)),
	"list<double>": reflect.TypeOf([]float64{}),
	"double":       reflect.TypeOf(float64(0)),
}

// Converts a go func into a CEL unary operation (single argument)
func unary(fn functions.FunctionOp) functions.UnaryOp {
	return func(arg ref.Val) ref.Val {
		return fn(arg)
	}
}

// Converts a go func into a CEL operation (any no. of arguments)
func function(fn interface{}, paramTypes []string, returnType string) functions.FunctionOp {
	return func(params ...ref.Val) ref.Val {
		fnValue := reflect.ValueOf(fn)

		// Check if the supplied, expected and implemented number of arguments are equal
		if len(params) != len(paramTypes) || len(paramTypes) != fnValue.Type().NumIn() {
			return types.NoSuchOverloadErr()
		}

		nativeParams := make([]reflect.Value, 0, len(params))

		// Check the expected types for params
		for i, p := range params {
			expectedType, ok := nativeTypeMap[paramTypes[i]]
			if !ok {
				return types.NewErr("type %s is not supported in the extensions, please add support.", paramTypes[i])
			}
			// convert the supplied param into native type to validate
			nativeP, err := p.ConvertToNative(expectedType)
			if err != nil {
				return types.NewErr(fmt.Sprintf("%s expects %dth argument to be of type %s", fn, i+1, expectedType))
			}

			// check if the native fn expects the supplied paramType
			implType := fnValue.Type().In(i)
			if implType != expectedType {
				return types.NewErr(fmt.Sprintf("%dth argument of %s is configured to be of type %s, native fn expects %s", i+1, fn, expectedType, implType))
			}

			nativeParams = append(nativeParams, reflect.ValueOf(nativeP))
		}

		response := fnValue.Call(nativeParams)
		if len(response) != 2 {
			return types.NewErr("native function has %d return values, needs one plus error", len(response))
		}

		// Check for error
		maybeErr := response[1].Interface()
		err, ok := maybeErr.(error)
		if maybeErr != nil && !ok {
			return types.NewErr("second return value of native function must be error")
		}
		if err != nil {
			return types.NewErr(err.Error())
		}

		// Check for returned value
		expectedType, ok := nativeTypeMap[returnType]
		if !ok {
			return types.NewErr("type %s is not supported in the extensions, please add support.", returnType)
		}
		returnVal := response[0]
		if returnVal.Type() != expectedType {
			return types.NewErr("unexpected return type (got %s, expected %s)", returnVal.Type(), expectedType)
		}

		return types.DefaultTypeAdapter.NativeToValue(returnVal.Interface())
	}
}
