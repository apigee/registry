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
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/interpreter/functions"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func Extensions() cel.EnvOption {
	return cel.Lib(extensionLib{})
}

type extensionLib struct{}

func (extensionLib) CompileOptions() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Declarations(
			decls.NewFunction("sum",
				decls.NewOverload("sum_int",
					[]*exprpb.Type{decls.NewListType(decls.Int)},
					decls.Int),
			),
		),
	}
}

func (extensionLib) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{
		cel.Functions(
			&functions.Overload{
				Operator: "sum",
				Unary:    unary(function(sum_int, []string{"list<int>"}, "int")),
			},
			&functions.Overload{
				Operator: "sum_int",
				Unary:    unary(function(sum_int, []string{"list<int>"}, "int")),
			},
		),
	}
}

func sum_int(vals []int64) (int64, error) {
	var rv int64
	for _, v := range vals {
		rv = rv + v
	}
	return rv, nil
}
