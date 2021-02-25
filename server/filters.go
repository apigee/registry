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

package server

import (
	"log"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

type filterArgType int

const (
	filterArgTypeString      filterArgType = iota
	filterArgTypeInt                       = iota
	filterArgTypeTimestamp                 = iota
	filterArgTypeStringArray               = iota
	filterArgTypeStringMap                 = iota
)

type filterArg struct {
	argName string
	argType filterArgType
}

func createFilterOperator(filter string, args []filterArg) (cel.Program, error) {
	if filter == "" {
		return nil, nil
	}
	dd := make([]*exprpb.Decl, 0)
	for _, pair := range args {
		switch pair.argType {
		case filterArgTypeString:
			dd = append(dd, decls.NewIdent(pair.argName, decls.String, nil))
		case filterArgTypeInt:
			dd = append(dd, decls.NewIdent(pair.argName, decls.Int, nil))
		case filterArgTypeTimestamp:
			dd = append(dd, decls.NewIdent(pair.argName, decls.Timestamp, nil))
		case filterArgTypeStringArray:
			dd = append(dd, decls.NewIdent(pair.argName, decls.NewListType(decls.String), nil))
		case filterArgTypeStringMap:
			dd = append(dd, decls.NewIdent(pair.argName, decls.NewMapType(decls.String, decls.String), nil))
		default:
			log.Fatalf("unknown filter argument type")
		}
	}
	d := cel.Declarations(dd...)
	env, err := cel.NewEnv(cel.Container("filter"), d)
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	ast, iss := env.Compile(filter)
	if iss.Err() != nil {
		return nil, invalidArgumentError(iss.Err())
	}
	prg, err := env.Program(ast)
	if err != nil {
		return nil, invalidArgumentError(err)
	}
	return prg, nil
}
