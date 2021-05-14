package list

import (
	"github.com/apigee/registry/cmd/control_loop/resources"
	"github.com/apigee/registry/rpc"
)

func GenerateApiHandler(result *[]resources.Resource) func(*rpc.Api) {
	return func(api *rpc.Api) {
		resource := resources.ApiResource{Api: api}
		(*result) = append((*result), resource)
	}
}


