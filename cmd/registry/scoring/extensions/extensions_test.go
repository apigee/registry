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
	"reflect"
	"testing"

	"github.com/google/cel-go/cel"
)

func TestSumInt(t *testing.T) {
	expression := "sum([1,2,3])"

	env, err := cel.NewEnv(Extensions())
	if err != nil {
		t.Fatalf("error creating CEL environment: %s", err)
	}
	ast, issues := env.Parse(expression)
	if issues != nil && issues.Err() != nil {
		t.Errorf("error parsing score_expression, %q: %s", expression, issues)
		return
	}
	prg, err := env.Program(ast)
	if err != nil {
		t.Errorf("program construction error, %q: %s", expression, err)
		return
	}

	out, _, err := prg.Eval(cel.NoVars())
	if err != nil {
		t.Errorf("evaluating expression %q returned unexpected error: %s", expression, err)
		return
	}
	want := 6
	nativeOut, err := out.ConvertToNative(reflect.TypeOf(want))
	if err != nil {
		t.Errorf("evaluating expression %q returned unexpected response type: want %s, got %s", expression, reflect.TypeOf(want), out.Type())
	}
	if nativeOut != want {
		t.Errorf("evaluating expression %q returned unexpected response: want %d, got %d", expression, want, nativeOut)
	}
}
