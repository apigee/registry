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

package filtering

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

type FieldType int

const (
	String    FieldType = iota
	Int       FieldType = iota
	Timestamp FieldType = iota
	StringMap FieldType = iota
)

type Field struct {
	Name string
	Type FieldType
}

type Filter struct {
	program cel.Program
}

func (f *Filter) Matches(model map[string]interface{}) (bool, error) {
	if f.program == nil {
		return true, nil
	}

	out, _, err := f.program.Eval(model)
	if err != nil {
		return false, status.Error(codes.InvalidArgument, err.Error())
	}

	match, ok := out.Value().(bool)
	if !ok {
		return false, status.Errorf(codes.InvalidArgument, "filter expression evaluation returned unexpected type: got %T, want bool", out.Value())
	}

	return match, nil
}

func NewFilter(filter string, fields []Field) (Filter, error) {
	if filter == "" {
		return Filter{}, nil
	}

	declarations := make([]*exprpb.Decl, 0)
	for _, field := range fields {
		switch field.Type {
		case String:
			declarations = append(declarations, decls.NewConst(field.Name, decls.String, nil))
		case Int:
			declarations = append(declarations, decls.NewConst(field.Name, decls.Int, nil))
		case Timestamp:
			declarations = append(declarations, decls.NewConst(field.Name, decls.Timestamp, nil))
		case StringMap:
			declarations = append(declarations, decls.NewConst(field.Name, decls.NewMapType(decls.String, decls.String), nil))
		default:
			return Filter{}, status.Errorf(codes.InvalidArgument, "unknown filter argument type")
		}
	}

	env, err := cel.NewEnv(cel.Container("filter"), cel.Declarations(declarations...))
	if err != nil {
		return Filter{}, status.Error(codes.InvalidArgument, err.Error())
	}

	ast, iss := env.Compile(filter)
	if iss.Err() != nil {
		return Filter{}, status.Error(codes.InvalidArgument, iss.Err().Error())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return Filter{}, status.Error(codes.InvalidArgument, err.Error())
	}

	return Filter{program: prg}, nil
}
