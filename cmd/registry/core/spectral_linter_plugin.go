package core

import (
	"fmt"

	"github.com/apigee/registry/rpc"
)

type SpectralLinter struct {
	Name  string
	Rules map[string][]string
}

func (linter SpectralLinter) AddRule(mimeType string, rule string) {
	linter.Rules[mimeType] = append(linter.Rules[mimeType], rule)
}

func (linter SpectralLinter) GetRules(mimeType string) []string {
	return linter.Rules[mimeType]
}

func (linter SpectralLinter) GetName() string {
	return linter.Name
}

func (linter SpectralLinter) SupportsMimeType(mimeType string) bool {
	// Spectral supports OpenAPI and AsyncAPI
	return true
}

func (linter SpectralLinter) Lint(mimeType string, path string) (*rpc.Lint, error) {
	fmt.Println("Linter got mime type:", mimeType, "and path:", path)
	return &rpc.Lint{}, nil
}
