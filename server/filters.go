package server

import (
	"log"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

type filterArgType int

const (
	filterArgTypeString filterArgType = iota
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
