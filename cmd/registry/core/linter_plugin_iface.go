package core

import "github.com/apigee/registry/rpc"

type Linter interface {
	AddRule(mimeType string, rule string)
	GetRules(mimeType string) []string
	GetName() string
	SupportsMimeType(mimeType string) bool
	Lint(mimeType string, path string) (*rpc.Lint, error)
}
