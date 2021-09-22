package conformance

import (
	"fmt"

	"github.com/apigee/registry/rpc"
)

type ApiLinter struct {
	Rules map[string][]string
}

func NewApiLinter() ApiLinter {
	return ApiLinter{
		Rules: make(map[string][]string),
	}
}

func (linter ApiLinter) AddRule(mimeType string, rule string) error {
	linter.Rules[mimeType] = append(linter.Rules[mimeType], rule)
	return nil
}

func (linter ApiLinter) GetName() string {
	return "api-linter"
}

func (linter ApiLinter) SupportsMimeType(mimeType string) bool {
	return true
}

func (linter ApiLinter) LintSpec(mimeType string, specPath string) ([]*rpc.LintProblem, error) {
	fmt.Println("Linter got mime type:", mimeType, "and path:", specPath)
	return []*rpc.LintProblem{}, nil
}
