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
)

func Extensions() cel.EnvOption {
	return cel.Lib(extensionLib{})
}

type extensionLib struct{}

func (extensionLib) CompileOptions() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Function(
			"sum",
			cel.Overload("sum_int", []*cel.Type{cel.ListType(cel.IntType)}, cel.IntType, cel.UnaryBinding(unary(function(sum_int, []string{"list<int>"}, "int")))),
		),
	}
}

func (extensionLib) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

func sum_int(vals []int64) (int64, error) {
	var rv int64
	for _, v := range vals {
		rv = rv + v
	}
	return rv, nil
}
