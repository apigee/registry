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
	"testing"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

func refVal(i interface{}) ref.Val {
	return types.DefaultTypeAdapter.NativeToValue(i)
}

func TestFunction(t *testing.T) {
	tests := []struct {
		desc       string
		fn         interface{}
		paramTypes []string
		returnType string
		args       []ref.Val
		wantErr    bool
	}{
		{
			desc:       "fn(int, int) int",
			fn:         func(x int64, y int64) (int64, error) { return x + y, nil },
			paramTypes: []string{"int", "int"},
			returnType: "int",
			args:       []ref.Val{refVal(1), refVal(2)},
		},
		{
			desc:       "fn(int, int) list<int>",
			fn:         func(x int64, y int64) ([]int64, error) { return []int64{x, y}, nil },
			paramTypes: []string{"int", "int"},
			returnType: "list<int>",
			args:       []ref.Val{refVal(1), refVal(2)},
		},
		{
			desc:       "missing err in fn",
			fn:         func(x int64, y int64) int64 { return x + y },
			paramTypes: []string{"int", "int"},
			returnType: "int",
			args:       []ref.Val{refVal(1), refVal(2)},
			wantErr:    true,
		},
		{
			desc:       "non-error second arg in fn",
			fn:         func(x int64, y int64) (int64, int64) { return x + y, 10 },
			paramTypes: []string{"int", "int"},
			returnType: "int",
			args:       []ref.Val{refVal(1), refVal(2)},
			wantErr:    true,
		},
		{
			desc:       "paramType mismatch in implementation",
			fn:         func(x int64, y int64) ([]int64, error) { return []int64{x, y}, nil },
			paramTypes: []string{"double", "double"},
			returnType: "list<int>",
			args:       []ref.Val{refVal(0.1), refVal(0.2)},
			wantErr:    true,
		},
		{
			desc:       "returnType mismatch in implementation",
			fn:         func(x int64, y int64) (int64, error) { return x + y, nil },
			paramTypes: []string{"int", "int"},
			returnType: "float64",
			args:       []ref.Val{refVal(1), refVal(2)},
			wantErr:    true,
		},
		{
			desc:       "paramType mismatch in args",
			fn:         func(x int64, y int64) ([]int64, error) { return []int64{x, y}, nil },
			paramTypes: []string{"int", "int"},
			returnType: "list<int>",
			args:       []ref.Val{refVal(0.1), refVal(0.2)},
			wantErr:    true,
		},
		{
			desc:       "error from native func",
			fn:         func(x int) (int, error) { return 0, fmt.Errorf("Sample error") },
			paramTypes: []string{"int"},
			returnType: "int",
			args:       []ref.Val{refVal(1)},
			wantErr:    true,
		},
		{
			desc:       "unsupported paramType",
			fn:         func(x int) (int, error) { return 0, nil },
			paramTypes: []string{"int64"},
			returnType: "int",
			args:       []ref.Val{refVal(1)},
			wantErr:    true,
		},
		{
			desc:       "unsupported returnType",
			fn:         func(x int) (int, error) { return 0, nil },
			paramTypes: []string{"int"},
			returnType: "int64",
			args:       []ref.Val{refVal(1)},
			wantErr:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := function(test.fn, test.paramTypes, test.returnType)(test.args...)
			if test.wantErr && !types.IsError(got) {
				t.Errorf("expected %s(%s) to return error", test.fn, test.args)
			}

			if !test.wantErr && types.IsError(got) {
				t.Errorf("%s(%s) returned unexpected error: %s", test.fn, test.args, got.Value())
			}
		})
	}
}
