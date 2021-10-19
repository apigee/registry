package linter

import "github.com/apigee/registry/rpc"

// LinterRunner is an interface through which a linter executes.
type LinterRunner interface {

	// Runs the linter with a provided linter request.
	Run(request *rpc.LinterRequest) (*rpc.LinterResponse, error)
}
