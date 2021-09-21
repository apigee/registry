package conformance

import "github.com/apigee/registry/rpc"

// Linter is an interface to lint specs in the registry
type Linter interface {
	// Add a new rule to the linter.
	AddRule(mimeType string, rule string) error

	// Gets the name of the linter.
	GetName() string

	// Returns whether the linter supports the provided mime type.
	SupportsMimeType(mimeType string) bool

	// Lints a provided specification of given mime type and returns a
	// LintFile object.
	LintSpec(mimeType string, specPath string) ([]*rpc.LintProblem, error)
}
