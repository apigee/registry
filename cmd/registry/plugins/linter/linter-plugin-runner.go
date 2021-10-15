package lint

import "github.com/apigee/registry/rpc"

// LinterPluginRunner is an interface through which a linter executes.
type LinterPluginRunner interface {

	// Runs the linter with a provided linter request.
	Run(request *rpc.LinterRequest) (*rpc.LinterResponse, error)
}
