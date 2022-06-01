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
