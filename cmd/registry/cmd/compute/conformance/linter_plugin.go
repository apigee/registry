package conformance

import "github.com/apigee/registry/rpc"

// Linter is an interface to lint specs in the registry
type Linter interface {
	// Add a new rule to the linter.
	AddRule(mimeType string, rule string)

	// Gets the name of the linter.
	GetName() string

	// Returns whether the linter supports the provided mime type.
	SupportsMimeType(mimeType string) bool

	// Lints a provided specification and returns a Lint object.
	Lint(mimeType string, path string) (*rpc.Lint, error)
}
