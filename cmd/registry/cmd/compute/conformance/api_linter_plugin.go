package conformance

import (
	"fmt"

	"github.com/apigee/registry/rpc"
)

type ApiLinter struct {
	Rules map[string][]string
}

func (linter ApiLinter) AddRule(mimeType string, rule string) {
	linter.Rules[mimeType] = append(linter.Rules[mimeType], rule)
}

func (linter ApiLinter) GetName() string {
	return "api-linter"
}

func (linter ApiLinter) SupportsMimeType(mimeType string) bool {
	return true
}

func (linter ApiLinter) Lint(mimeType string, path string) (*rpc.Lint, error) {
	fmt.Println("Linter got mime type:", mimeType, "and path:", path)
	return &rpc.Lint{}, nil
}
